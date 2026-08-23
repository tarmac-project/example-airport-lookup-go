COMPONENTS = functions/src/data/fetch functions/src/data/init functions/src/data/load functions/src/data/seed functions/src/handlers/lookup
ROOT = $(CURDIR)
COVERAGE_DIR ?= $(ROOT)/coverage
TINYGO_IMAGE ?= tinygo/tinygo:0.38.0
TINYGO_FLAGS ?= -scheduler=none --no-debug -target=wasip1 -buildmode=c-shared

.PHONY: all clean tests lint build format benchmarks tidy

all: build tests lint

build:
	@echo "Building all modules..."
	@for dir in $(COMPONENTS); do \
		$(MAKE) -C $$dir build ROOT="$(ROOT)" TINYGO_IMAGE="$(TINYGO_IMAGE)" TINYGO_FLAGS="$(TINYGO_FLAGS)" || exit 1; \
	done

tests:
	@echo "Running tests for all modules..."
	mkdir -p $(COVERAGE_DIR)
	go test -v -race -covermode=atomic -coverprofile=$(COVERAGE_DIR)/root.out ./...
	@for dir in $(COMPONENTS); do \
		$(MAKE) -C $$dir tests COVERAGE_DIR="$(COVERAGE_DIR)" || exit 1; \
	done

benchmarks:
	@echo "Running benchmarks for all modules..."
	go test -bench=. -benchmem ./...
	@for dir in $(COMPONENTS); do \
		$(MAKE) -C $$dir benchmarks || exit 1; \
	done

format:
	@echo "Formatting code..."
	@gofmt -s -w .
	@if command -v golines >/dev/null 2>&1; then \
		golines -w .; \
	else \
		echo "golines not installed, skipping line wrapping"; \
	fi
	@for dir in $(COMPONENTS); do \
		$(MAKE) -C $$dir format || exit 1; \
	done

lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		set -e; \
		golangci-lint run ./...; \
		for dir in $(COMPONENTS); do \
			$(MAKE) -C $$dir lint || exit 1; \
		done; \
	else \
		echo "golangci-lint not installed, skipping lint"; \
	fi

tidy:
	## Run go mod tidy for all function modules
	@set -e; \
	for mod in $$(find functions/src -name go.mod | sort); do \
		dir=$$(dirname $$mod); \
		echo "==> $$dir"; \
		(cd $$dir && go mod tidy); \
	done

docker-compose:
	docker compose up -d mysql redis
	sleep 15
	docker compose up data-manager lookup

docker-compose-background:
	docker compose up -d mysql redis
	sleep 15
	docker compose up -d data-manager lookup

loadtest-setup:
	docker compose -f load-compose.yml up -d mysql
	sleep 15
	docker compose -f load-compose.yml up -d data-manager lookup

run: build docker-compose
run-nobuild: docker-compose
run-background: build docker-compose-background
run-stress: build loadtest-setup
	sleep 600
	k6 run --config tests/k6/stress.json tests/k6/script.js
run-soak: build loadtest-setup
	sleep 600
	k6 run --config tests/k6/soak.json tests/k6/script.js
run-steady: build docker-compose-background
	sleep 600
	k6 run --config tests/k6/steady.json tests/k6/script.js


clean:
	@echo "Cleaning build artifacts..."
	@for dir in $(COMPONENTS); do \
		$(MAKE) -C $$dir clean || exit 1; \
	done
	rm -rf functions/build
	@find . -type f -name "*.test" -delete
	@find . -type f -name "coverage.out" -delete
	@find . -type f -name "coverage.html" -delete
	docker compose down --remove-orphans
