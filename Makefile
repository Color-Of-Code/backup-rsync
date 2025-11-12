# Makefile to build the project and place the binary in the dist/ directory

.PHONY: build clean deps test

deps:
	go mod tidy

build: deps
	@mkdir -p dist
	go build -o dist/backup ./backup/main.go

test:
	go test ./... -v

clean:
	rm -rf dist