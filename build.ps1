# Parent dir is root
$scriptDir = Get-Location
$outDir = Join-Path -Path $scriptDir -ChildPath "/_out"

# Windows x64
$env:GOARCH = "amd64"
$env:GOOS = "windows"
$env:CGO_ENABLED = 0
go build -o $outDir/brother-cert.exe ./cmd/brother-cert

# Linux x64
$env:GOARCH = "amd64"
$env:GOOS = "linux"
$env:CGO_ENABLED = 0
go build -o $outDir/brother-cert-amd64 ./cmd/brother-cert

# Linux arm64
$env:GOARCH = "arm64"
$env:GOOS = "linux"
$env:CGO_ENABLED = 0
go build -o $outDir/brother-cert-arm64 ./cmd/brother-cert
