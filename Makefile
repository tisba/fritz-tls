BINARY=fritz-tls

.PHONY: all build fmt bootstrap lint

all: build

build:
	go build -ldflags="-s -w" -o ${BINARY}

fmt:
	go fmt ./...

bootstrap.homebrew:
	brew install --quiet golangci-lint
	brew install --quiet golangci/tap/golangci-lint goreleaser

lint:
	golangci-lint run
