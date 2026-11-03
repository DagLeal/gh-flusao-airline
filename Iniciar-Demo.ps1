$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
$env:DEMO_MODE = "true"
$env:COOKIE_SECURE = "false"
$env:DATABASE_PATH = Join-Path $PSScriptRoot "data/aeroporto.db"
if (-not $env:PORT) { $env:PORT = "8080" }
Write-Host ("Abra http://localhost:" + $env:PORT + " no navegador. Ctrl+C encerra.")
$serverExecutable = Join-Path $PSScriptRoot "bin/aeroporto.exe"
if (Test-Path -LiteralPath $serverExecutable) {
    & $serverExecutable
} else {
    go run ./cmd/server
}
