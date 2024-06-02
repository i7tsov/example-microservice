#!/bin/sh

mkdir -p bin
GOOS=linux CGO_ENABLED=0 \
    go build -ldflags=" \
        -X main.version=$(git describe --tags --always --long --dirty='-dirty')" \
    -o ./bin/ ./cmd/server
