# Usage: .\make.ps1 [build|serve]
param(
    [Parameter(Position = 0)]
    [string]$Target = "build"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Invoke-Build {
    Write-Host "Building spd.exe..."
    go build -o spd.exe
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

switch ($Target) {
    "build" {
        Invoke-Build
    }
    "serve" {
        Invoke-Build
        Write-Host "Starting server..."
        .\spd.exe serve
    }
    default {
        Write-Error "Unknown target '$Target'. Available targets: build, serve"
        exit 1
    }
}
