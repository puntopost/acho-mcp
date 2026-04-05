#!/bin/sh
set -e

# Format check
FILES=$(find . -name "*.go" ! -path "./vendor/*" | xargs gofmt -l)
if [ -n "$FILES" ]; then
    echo "gofmt needed:"
    echo "$FILES"
    exit 1
fi

# Vet
go vet ./...

# Build
go build -o bin/acho ./cmd/acho

# Test
go test -v ./...
