# PowerShell script to run Go Runner with Hot Reload
Write-Host "Starting Go Runner with Hot Reload..." -ForegroundColor Green
$env:CGO_ENABLED = "1"
go run cmd/hotreload/main.go
