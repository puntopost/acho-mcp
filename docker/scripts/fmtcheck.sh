#!/bin/sh
set -e
FILES=$(find . -name "*.go" ! -path "./vendor/*" | xargs gofmt -l)
if [ -n "$FILES" ]; then
    echo "gofmt needed:"
    echo "$FILES"
    exit 1
fi
