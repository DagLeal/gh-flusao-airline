$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)
if (-not $env:DEMO_MODE -and -not $env:ADMIN_EMAIL) { $env:DEMO_MODE = "true" }
$env:COOKIE_SECURE = "false"
go run ./cmd/server
if ($LASTEXITCODE -ne 0) { throw "O servidor terminou com erro" }
