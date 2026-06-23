bin := if os_family() == "windows" { "build/pinter.exe" } else { "build/pinter" }
run_bin := if os_family() == "windows" { "./build/pinter.exe" } else { "./build/pinter" }

build:
	go build -o {{bin}} ./cmd/pinter

run: build
	{{run_bin}}

[windows]
test:
	GOFLAGS=-buildmode=exe go test ./...

[unix]
test:
	go test ./...

clean:
	go clean
