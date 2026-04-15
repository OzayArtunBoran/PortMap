#!/bin/bash
set -e

echo "Running pre-commit checks..."

echo "-> go vet"
go vet ./...

echo "-> go test (short)"
go test ./... -count=1 -short

echo "All checks passed!"
