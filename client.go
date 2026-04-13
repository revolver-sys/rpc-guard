package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sort"
	"time"
)

type Client struct {
	endpoints      []*Endpoint
	httpClient     *http.Client
	retryBudget    int
	baseBackoff    time.Duration
	circuitOpenFor time.Duration
}

func NewClient(urls []string) *Client {
	endpoints := make([]*Endpoint, 0, len(urls))
	for _, u := range urls {
		endpoints = append(endpoints, NewEndpoint(u))
	}

	return &Client{
		endpoints: endpoints,
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
		retryBudget:    4,
		baseBackoff:    150 * time.Millisecond,
		circuitOpenFor: 20 * time.Second,
	}
}

func (c *Client) Request(ctx context.Context, method string, params interface{}) (*JSONRPCResponse, error) {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	tried := map[string]bool{}

	for attempt := 0; attempt < c.retryBudget; attempt++ {
		now := time.Now()
		ep := c.pickBestEndpoint(now, tried)
		if ep == nil {
			return nil, fmt.Errorf("no available endpoints: %w", lastErr)
		}
		tried[ep.URL] = true

		start := time.Now()
		resp, err := c.doPost(ctx, ep.URL, data)
		latency := time.Since(start)

		if err == nil && resp.Error == nil {
			ep.RecordSuccess(latency, now)
			return resp, nil
		}

		errMsg := "unknown error"
		if err != nil {
			errMsg = err.Error()
			lastErr = err
		} else if resp != nil && resp.Error != nil {
			errMsg = fmt.Sprintf("json-rpc error: %v", resp.Error)
			lastErr = errors.New(errMsg)
		}

		ep.RecordFailure(errMsg, now, c.circuitOpenFor)

		sleep := c.backoff(attempt)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(sleep):
		}
	}

	if lastErr == nil {
		lastErr = errors.New("request failed after retries")
	}
	return nil, lastErr
}

func (c *Client) doPost(ctx context.Context, url string, payload []byte) (*JSONRPCResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("http status %d: %s", res.StatusCode, string(body))
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &rpcResp, nil
}

func (c *Client) pickBestEndpoint(now time.Time, tried map[string]bool) *Endpoint {
	candidates := make([]*Endpoint, 0, len(c.endpoints))
	for _, ep := range c.endpoints {
		if tried[ep.URL] {
			continue
		}
		if !ep.IsAvailable(now) {
			continue
		}
		candidates = append(candidates, ep)
	}

	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score(now) > candidates[j].Score(now)
	})

	// Small randomization among top candidates to avoid hard pinning.
	if len(candidates) > 1 && rand.Intn(100) < 20 {
		return candidates[1]
	}
	return candidates[0]
}

func (c *Client) backoff(attempt int) time.Duration {
	base := c.baseBackoff * time.Duration(1<<attempt)
	jitter := time.Duration(rand.Intn(120)) * time.Millisecond
	return base + jitter
}

func (c *Client) DebugState() []map[string]any {
	now := time.Now()
	out := make([]map[string]any, 0, len(c.endpoints))
	for _, ep := range c.endpoints {
		out = append(out, ep.Snapshot(now))
	}
	return out
}
