FROM golang:1.26.1-trixie

RUN apt-get update && apt-get install -y --no-install-recommends libsqlite3-dev && rm -rf /var/lib/apt/lists/*

ENV CGO_ENABLED=1
ENV GOFLAGS="-mod=vendor -tags=sqlite_fts5 -buildvcs=false"
ENV GOPATH=/tmp/go

WORKDIR /app
