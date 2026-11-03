$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)
go run github.com/a-h/templ/cmd/templ@v0.3.1100 generate
if ($LASTEXITCODE -ne 0) { throw "templ generate falhou" }
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
if ($LASTEXITCODE -ne 0) { throw "sqlc generate falhou" }
npm ci
if ($LASTEXITCODE -ne 0) { throw "npm ci falhou" }
npm run build:css
if ($LASTEXITCODE -ne 0) { throw "Tailwind falhou" }
go test ./...
if ($LASTEXITCODE -ne 0) { throw "Os testes falharam" }
