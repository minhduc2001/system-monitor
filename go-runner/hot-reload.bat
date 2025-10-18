@echo off
echo Starting Go Runner with Hot Reload...
set CGO_ENABLED=1
go run cmd/hotreload/main.go
