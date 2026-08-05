# Contributing

## Development setup

Requirements:

- Go 1.23 or newer.
- Git.

```bash
git clone https://github.com/appoloncel283-debug/pulsenet.git
cd pulsenet
go test ./...
go vet ./...
go build ./cmd/pulsenet
```

## Pull requests

Keep changes focused and include tests for behavior changes. Before opening a pull request, run:

```bash
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

Avoid adding dependencies unless the benefit clearly justifies the additional supply-chain and binary-size cost.

## Bug reports

Include:

- Operating system and architecture.
- PulseNet version.
- Exact command used.
- Expected and actual behavior.
- Redacted output or report when relevant.

Do not publish credentials, internal hostnames, tokens, private IP plans, or sensitive certificate details in public issues.
