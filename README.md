# ipv6test

A tiny IPv6 status page. It shows the client's IPv6 address and confirms NAT64 on
an IPv6-only network.

It's meant to be hosted on an **IPv6-only** address. That's the real test: if a
visitor can load the page, their IPv6 works. The page then congratulates them and
shows the address the server saw.

## What it shows

- **Your IPv6 address** — read from the TCP connection and rendered into the page
  by the server (browser JS can't see its own public address).
- **NAT64** — the browser probes well-known IPv4 addresses (`1.1.1.1`, `8.8.8.8`)
  encoded under the well-known NAT64 prefix `64:ff9b::/96`. If a request makes it
  back, NAT64 works. The probe is a plain `no-cors` GET, entirely client-side.

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
- The NAT64 probe targets are hardcoded in `web/index.html` (`NAT64_TARGETS`) as
  IPv6 literals under `64:ff9b::/96`. Edit that list to probe other hosts.
- The NAT64 probe is a plain `http://` request to a literal address. If you
  terminate TLS in front of this app, browsers will block that probe as mixed
  content. Serve the app over HTTP on the IPv6 path (or accept that the NAT64
  card will fail under HTTPS).

## API

- `GET /` — the page, with the client address rendered in.
- `GET /healthz` — liveness/readiness probe.
