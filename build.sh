#!/usr/bin/env sh
set -eu
VERSION="${VERSION:-2.4.0}"
mkdir -p bin
go test ./...
go vet ./...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o bin/pulsenet-windows-amd64.exe ./cmd/pulsenet
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o bin/pulsenet-linux-amd64 ./cmd/pulsenet
printf 'Built PulseNet %s for Windows and Linux.\n' "$VERSION"
