param([string]$Output = ("backups/aeroporto-" + (Get-Date -Format "yyyyMMdd-HHmmss") + ".db"))
$ErrorActionPreference = "Stop"
Set-Location (Split-Path -Parent $PSScriptRoot)
go run ./cmd/server backup --output $Output
if ($LASTEXITCODE -ne 0) { throw "Falha no backup" }
