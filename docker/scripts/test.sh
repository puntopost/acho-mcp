#!/bin/sh
set -e
go build -o bin/acho ./cmd/acho
go test -v ./...
