BINARY := pulsenet
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: test vet fmt build release clean

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./cmd ./internal

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/pulsenet

release: test vet
	mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/pulsenet-windows-amd64.exe ./cmd/pulsenet
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/pulsenet-linux-amd64 ./cmd/pulsenet
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/pulsenet-darwin-amd64 ./cmd/pulsenet
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o dist/pulsenet-darwin-arm64 ./cmd/pulsenet

clean:
	rm -rf dist $(BINARY) $(BINARY).exe
