#!/bin/sh
set -e
export GOFLAGS=
go mod tidy
go mod vendor
