#!/bin/bash
echo "Starting Go Runner with Hot Reload..."
export CGO_ENABLED=1
go run cmd/hotreload/main.go
