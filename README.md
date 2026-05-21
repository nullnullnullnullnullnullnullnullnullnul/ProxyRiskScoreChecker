# ProxyRiskScoreChecker

[![ci](https://github.com/nullnullnullnullnullnullnullnullnullnul/ProxyRiskScoreChecker/actions/workflows/ci.yml/badge.svg)](https://github.com/nullnullnullnullnullnullnullnullnullnul/ProxyRiskScoreChecker/actions/workflows/ci.yml)
[![Go 1.24](https://img.shields.io/badge/Go-1.24-00ADD8.svg)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

CLI tool that queries the [IPQualityScore (IPQS)](https://www.ipqualityscore.com/) reputation API for a batch of HTTP / HTTPS / SOCKS5 proxies. For each proxy, the tool resolves the outbound IP it presents to upstream services, then looks that IP up in the IPQS API and reports the returned `fraud_score`.

## Overview

Intended for operators of a proxy fleet who want to monitor the IPQS reputation of the exit IPs they expose to third-party services. The tool does not manipulate traffic: for each proxy it makes one outbound-IP probe and one IPQS API call, then writes the proxies whose resolved IP returned `fraud_score=0` to the configured output file.

External services contacted:

- `http://ipinfo.io/json` to resolve the outbound IP a proxy presents.
- `https://ipqualityscore.com/api/json/ip/...` for the reputation lookup (requires an API key).

## Build

```bash
go build -o build/proxyriskscore ./cmd/proxyriskscore
```

## Usage

```text
Usage of proxyriskscore:
  -api-key string
        IPQS API key (falls back to $IPQS_API_KEY)
  -input string
        input file with one proxy per line (default "proxies.txt")
  -output string
        output file for proxies with fraud_score=0 (default "clean.txt")
  -strictness int
        IPQS strictness level (0-3)
  -timeout duration
        per-request timeout (default 10s)
```

The IPQS API key is required, either via `--api-key` or the `IPQS_API_KEY` environment variable.

### Supported proxy input formats

One per line. Lines that don't match any of the formats below are skipped with a warning:

- `protocol://user:pass@host:port`
- `protocol://host:port`
- `user:pass@host:port` (protocol defaults to `http`)
- `host:port:user:pass` (proxy-provider list format)
- `host:port` (protocol defaults to `http`)

Where `protocol` is one of `http`, `https`, `socks5`.

### Example

```bash
export IPQS_API_KEY="your_key_here"
./build/proxyriskscore --input proxies.txt --output clean.txt --strictness 1
```

Sample log output:

```text
time=2026-05-21T18:00:00Z level=INFO msg="loaded proxies" count=120 file=proxies.txt
time=2026-05-21T18:00:00Z level=INFO msg=clean proxy=http://10.0.0.1:8080 ip=185.213.155.74
time=2026-05-21T18:00:01Z level=INFO msg=flagged proxy=http://10.0.0.2:8080 ip=45.33.32.156 score=75
time=2026-05-21T18:00:02Z level=WARN msg="skip: proxy probe failed" proxy=10.0.0.3:8080 err="do request: context deadline exceeded"
time=2026-05-21T18:00:10Z level=INFO msg=done clean=94 total=120 output=clean.txt
```

## Quality

```bash
gofmt -l .          # formatting check (CI fails on any output)
go vet ./...        # stdlib static checks
go test ./...       # unit tests
golangci-lint run   # extended lint suite
```

All four are enforced in CI.

## Conventions

- **Commits**: [Conventional Commits](https://www.conventionalcommits.org/) (`feat`, `fix`, `refactor`, `docs`, `chore`, `ci`, `test`, ...).
- **Branching**: `main` is the only long-lived branch. Changes land on topic branches via PR.
- **PRs**: CI must be green. Squash merge.

## Roadmap

- Concurrency for the per-proxy lookup loop (currently sequential).
- Optional JSON / NDJSON output alongside the plain-text `host:port` format.
- Integration tests covering the `proxy.Validator` path (requires a fixture SOCKS5/HTTP proxy).

## Migration from the prior interactive CLI

This refactor introduces breaking changes relative to the previous interactive version:

- Env var renamed: `API_KEY` is now `IPQS_API_KEY`.
- No interactive stdin prompts. Use flags.
- Default output renamed: `proxies_risk_score_0.txt` is now `clean.txt` (override with `--output`).
- The intermediate `validproxys.txt` file is no longer written; validation now happens in memory only.

## License

MIT. See [LICENSE](LICENSE).
