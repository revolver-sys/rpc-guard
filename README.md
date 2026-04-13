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
a minimal control layer showing how RPC clients can behave more intelligently under failure.

## Current ideas

- choose the best available endpoint instead of random retrying
- temporarily quarantine degraded endpoints
- make client behavior observable
- treat partial failure as a first-class condition
