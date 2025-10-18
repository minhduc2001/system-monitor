@echo off
echo Starting Go Runner with CGO enabled...
set CGO_ENABLED=1
go run cmd/server/main.go
