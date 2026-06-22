.PHONY: build run test clean

ifeq ($(OS),Windows_NT)
BIN := build/pinter.exe
RUN_BIN := .\build\pinter.exe
else
BIN := build/pinter
RUN_BIN := ./build/pinter
endif

build:
	go build -o $(BIN) ./cmd/pinter

ifeq ($(wildcard $(BIN)),)
run: build
else
run:
endif
	$(RUN_BIN)

test:
	go test ./...

clean:
	go clean
