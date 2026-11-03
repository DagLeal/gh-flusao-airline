#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
go run github.com/a-h/templ/cmd/templ@v0.3.1100 generate
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
npm ci
npm run build:css
go test ./...
