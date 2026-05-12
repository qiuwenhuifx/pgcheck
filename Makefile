BIN ?= bin/pgcheck
GO ?= go
VERSION ?= 2.0.1
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.PHONY: build test fmt clean

build:
	$(GO) build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" -o $(BIN) .

test:
	$(GO) test ./...

fmt:
	gofmt -w main.go internal/app/*.go internal/pgexec/*.go internal/queries/*.go

clean:
	rm -rf bin dist coverage.out
