BINARY_NAME := kznginx
MODULE      := github.com/karoza/kz-nginx-generator
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS     := -s -w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT)

.PHONY: build test test-cover lint fmt vet run-ui run-generate clean tidy

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .

test:
	go test -v -cover ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

fmt:
	gofmt -l .

vet:
	go vet ./...

run-ui: build
	./bin/$(BINARY_NAME) ui --port 8080

run-generate: build
	./bin/$(BINARY_NAME) generate --domain example.com --proxy http://localhost:8000

tidy:
	go mod tidy

clean:
	rm -rf bin dist coverage.out
