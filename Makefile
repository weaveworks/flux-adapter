.PHONY: all

all: flux-adapter

VERSION := $(shell git describe --tags --dirty)

flux-adapter: cmd/flux-adapter/*.go
	go build -o "$@" -ldflags "-X main.version=$(VERSION)" ./cmd/flux-adapter
