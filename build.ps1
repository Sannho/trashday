# Builds trashday.exe: Tailwind CSS, tests, Windows resources (manifest + icon), binary.
# The resources (rsrc.syso) are embedded only for the build: left in place, they would
# also be linked into `go test` binaries, which then refuse to start without elevation.
$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

# Tailwind standalone CLI v4.3.3 (no Node), fetched once from the official release and checksum-verified.
if (-not (Test-Path bin\tailwindcss.exe)) {
    New-Item -ItemType Directory bin -Force | Out-Null
    Invoke-WebRequest https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.3/tailwindcss-windows-x64.exe -OutFile bin\tailwindcss.exe
    if ((Get-FileHash bin\tailwindcss.exe -Algorithm SHA256).Hash -ne 'E0E260CE048014E9268F6237FF18F8CCF02CEF521CBD0AE04E82C2CDF7AA3955') {
        Remove-Item bin\tailwindcss.exe
        throw "tailwindcss.exe checksum mismatch"
    }
}
bin\tailwindcss.exe -i ui\style.css -o ui\app.css --minify
if ($LASTEXITCODE) { throw "tailwindcss failed" }

go test ./...
if ($LASTEXITCODE) { throw "tests failed" }

$syso = 'cmd\trashday\rsrc.syso'
go run github.com/akavel/rsrc@v0.10.2 -manifest build\windows\trashday.manifest -ico build\windows\trashday.ico -o $syso
if ($LASTEXITCODE) { throw "rsrc failed" }
try {
    go build -ldflags="-H windowsgui -s -w" -o trashday.exe ./cmd/trashday
    if ($LASTEXITCODE) { throw "go build failed" }
} finally {
    Remove-Item $syso
}
Write-Host "OK: trashday.exe"
