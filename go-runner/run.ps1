# PowerShell script to run Go Runner with CGO enabled
Write-Host "Starting Go Runner with CGO enabled..." -ForegroundColor Green
$env:CGO_ENABLED = "1"
go run cmd/server/main.go
