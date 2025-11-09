#!/bin/bash
echo "Starting Go Runner with CGO enabled..."
export CGO_ENABLED=1
go run cmd/server/main.go
