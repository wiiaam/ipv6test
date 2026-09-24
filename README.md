# ipv6test

A tiny IPv6 status page. Hosted on an **IPv6-only** address, it's the real test:
if a visitor can load the page, their IPv6 works. It then shows the address the
server saw. No JavaScript — the address is rendered server-side.

## What it shows

- **You have IPv6!** — the page is only reachable over IPv6, so loading it proves
  end-to-end IPv6 connectivity.
- **Your IP address** — read from the TCP connection and rendered into the page
  by the server (browser JS can't see its own public address).

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN` | `:8080` | Listen address (bind to an IPv6 address for an IPv6-only deployment, e.g. `[::]:8080`). |
| `TRUST_PROXY` | `false` | Honour `X-Forwarded-For` / `X-Real-IP` when behind a proxy. |

## Run

```sh
go run .
# open http://[::1]:8080
```

### Docker

```sh
docker build -t ipv6test .
docker run --rm -p 8080:8080 ipv6test
```

## Deploying IPv6-only

- Expose the service on an IPv6 address only. Visitors reaching it is the proof
  of connectivity.
- Works fine behind TLS: the page has no mixed content, no cross-origin
  requests, and no JS.

## Future ideas

- **NAT64 test over HTTPS**: host something on a VPS with an AAAA record pointing
  into the NAT64 prefix (e.g. `64:ff9b::<your IPv4 address>`) and a TLS cert for
  that hostname. If a client can connect to it over HTTPS, they reached your IPv4
  address through NAT64. This confirms NAT64 even for dual-stack clients (the
  AAAA is the only record, so there's no IPv4 fallback).

## API

- `GET /` — the page, with the client address rendered in.
- `GET /healthz` — liveness/readiness probe.