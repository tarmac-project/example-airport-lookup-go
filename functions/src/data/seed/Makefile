.PHONY: tests build benchmarks format lint clean

ROOT ?= $(abspath ../../../..)
COVERAGE_DIR ?= $(ROOT)/coverage
TINYGO_IMAGE ?= tinygo/tinygo:0.38.0
TINYGO_FLAGS ?= -scheduler=none --no-debug -target=wasip1 -buildmode=c-shared

tests:
	mkdir -p $(COVERAGE_DIR)
	go test -v -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/seed.out ./...

build:
	mkdir -p $(ROOT)/functions/build/data
	docker run --rm -v $(ROOT):/build -w /build/functions/src/data/seed $(TINYGO_IMAGE) tinygo build -o /build/functions/build/data/seed.wasm $(TINYGO_FLAGS) main.go
	go test -run '^$$' ./...

benchmarks:
	go test -bench=. -benchmem ./...

format:
	gofmt -s -w .
	@if command -v golines >/dev/null 2>&1; then \
		golines -w .; \
	else \
		echo "golines not installed, skipping line wrapping"; \
	fi

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

clean:
	find . -type f -name "*.test" -delete
	find . -type f -name "coverage.out" -delete
	find . -type f -name "coverage.html" -delete
