# rpc-guard

A small Go prototype for exploring RPC reliability under degraded conditions.

## Why

In distributed systems, outages are often amplified by recovery behavior:
- retries create extra load
- clients keep hitting degraded endpoints
- partial failures get misread as total failures

This tool explores a few basic countermeasures:
- endpoint scoring
- retry budgeting
- jittered backoff
- circuit breaking
- endpoint switching

## What it is

Not a production-ready client.

A thinking artifact:
a minimal control layer for exploring how RPC clients behave under partial failure.

## Current ideas

- choose the best available endpoint instead of random retrying
- temporarily quarantine degraded endpoints
- make client behavior observable
- treat partial failure as a first-class condition

## Example output

Single request with endpoint scoring and latency tracking:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x17b85f6",
  "error": null
}
[
  {
    "available": true,
    "avg_latency": "639.736332ms",
    "consec_fails": 0,
    "failures": 0,
    "last_error": "",
    "opened_until": "0001-01-01T00:00:00Z",
    "successes": 1,
    "url": "https://ethereum-rpc.publicnode.com"
  },
  {
    "available": true,
    "avg_latency": "500ms",
    "consec_fails": 0,
    "failures": 0,
    "last_error": "",
    "opened_until": "0001-01-01T00:00:00Z",
    "successes": 0,
    "url": "https://rpc.ankr.com/eth"
  },
  {
    "available": true,
    "avg_latency": "500ms",
    "consec_fails": 0,
    "failures": 0,
    "last_error": "",
    "opened_until": "0001-01-01T00:00:00Z",
    "successes": 0,
    "url": "https://cloudflare-eth.com"
  }
]
```

Endpoints are scored dynamically based on latency and observed failure patterns.

