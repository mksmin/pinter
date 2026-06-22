.PHONY: build test

build:
	go build -o build/pinter ./cmd/pinter

test:
	go test ./...
