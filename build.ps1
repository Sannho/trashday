# Embeds the manifest (admin rights + modern controls) only for the build:
# a rsrc.syso left in the folder would also be linked into `go test` binaries,
# which then refuse to start without elevation.
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
bin\tailwindcss.exe -i ui.input.css -o ui.css --minify
if ($LASTEXITCODE) { throw "tailwindcss failed" }
go test .
if ($LASTEXITCODE) { throw "tests failed" }
go run github.com/akavel/rsrc@v0.10.2 -manifest trashday.manifest -ico trashday.ico -o rsrc.syso
try {
    go build -ldflags="-H windowsgui -s -w" -o trashday.exe .
} finally {
    Remove-Item rsrc.syso
}
Write-Host "OK: trashday.exe"
