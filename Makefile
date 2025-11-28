# Makefile to build the project and place the binary in the dist/ directory

# Build command with common flags
BUILD_CMD = CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -tags=prod
PACKAGE = ./backup/main.go

.PHONY: build clean test lint tidy checksums release sanity-check check-mod-tidy lint-config-check  lint-fix format check-clean

format:
	go fmt ./...
	@echo "OK: Code formatted."

lint-config-check:
	@golangci-lint config path
	@golangci-lint config verify
	@echo "OK: Lint configuration is valid."

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

check-clean:
	@git diff --quiet || (echo "ERROR: Working directory has uncommitted changes." && exit 1)
	@echo "OK: Working directory is clean."

check-mod-tidy:
	go mod tidy -diff
	@echo "OK: No untidy module files detected."

sanity-check: format check-clean check-mod-tidy
	@echo "OK: All sanity checks passed."

test:
	go test ./... -v

tidy:
	go mod tidy

build:
	@mkdir -p dist
	$(BUILD_CMD) -o dist/backup $(PACKAGE)

clean:
	rm -rf dist

# Build for specific OS and architecture (e.g., make release-linux-amd64)
release-%:
	@mkdir -p dist
	GOOS=$(word 1,$(subst -, ,$*)) GOARCH=$(word 2,$(subst -, ,$*)) $(BUILD_CMD) -o dist/backup-$* $(PACKAGE)

checksums:
	@for file in dist/*; do \
		if [ "$${file##*.}" != "sha256" ]; then \
			sha256sum "$$file" > "$$file.sha256"; \
		fi; \
	done

release: release-linux-amd64 release-darwin-amd64 release-windows-amd64 checksums
	@echo
	@echo "Binaries with sizes and checksums:"
	@for file in dist/*; do \
		if [ -f "$$file" ] && [ "$${file##*.}" != "sha256" ]; then \
			size=$$(stat --printf="%s" "$$file"); \
			checksum=$$(cat "$$file.sha256" | awk '{print $$1}'); \
			printf "%-40s %-15s %-64s\n" "$$file" "Size: $$size bytes" "Checksum: $$checksum"; \
		fi; \
	done

report-size: build
	go install github.com/Zxilly/go-size-analyzer/cmd/gsa@latest
	gsa --web --listen=":8910" --open dist/backup

report-coverage:
	@mkdir -p coverage
	@go test ./... -count=1 -coverprofile=coverage/coverage.out -coverpkg=./backup/...
	@echo
	@echo "Coverage Summary:"
	@go tool cover -func=coverage/coverage.out
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated at coverage/coverage.html"
