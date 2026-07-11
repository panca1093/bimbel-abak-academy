#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../backend"

export CGO_ENABLED=0 GOOS=linux GOARCH=amd64

go build -ldflags "-s -w" -o bin/api ./cmd/api
go build -ldflags "-s -w" -o bin/worker ./cmd/worker
