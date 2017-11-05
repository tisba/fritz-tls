BINARY=fritz-tls

.PHONY: all build fmt bootstrap lint

all: build

build:
	go build -o ${BINARY}

fmt:
	gofmt -w -s $(shell find . -type f -name '*.go' -not -path "./vendor/*")

bootstrap:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/kisielk/errcheck
	go get -u gopkg.in/alecthomas/gometalinter.v1
	gometalinter.v1 --install
	dep ensure

lint:
	gometalinter.v1 ./... --vendor .
