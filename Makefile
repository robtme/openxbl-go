PROJECT			 := github.com/robtme/openxbl-go
GO				 := $(shell which go 2>/dev/null)
GOFUMPT			 := $(shell which gofumpt 2>/dev/null)
GOLINTER		 := $(shell which golangci-lint 2>/dev/null)
GONILAWAY        := $(shell which nilaway 2>/dev/null)
GO_BENCH_FLAGS	 := -short -bench=. -benchmem
GO_BENCH		 := $(GO) test $(GO_BENCH_FLAGS)
GO_FORMAT		 := $(GOFUMPT) -w
GO_TEST			 := $(GO) test -v -short
GO_TIDY			 := $(GO) mod tidy

all: check-dependencies format lint nil

check-dependencies:
	@echo "Checking dependencies..."
	@if [ -z "${GO}" ]; then \
		echo "Cannot find 'go' in your $$PATH"; \
		exit 1; \
	fi
	@if [ -z "${GOFIELDALIGNMENT}" ]; then \
		echo "Cannot find 'fieldalignment' in your $$PATH"; \
		exit 1; \
	fi

format:
	@if [ -z "${GOFUMPT}" ]; then \
		echo "Cannot find 'gofumpt' in your $$PATH"; \
		exit 1; \
	fi
	@echo "Formatting code..."
	@$(GO_FORMAT) $(PWD)

lint:
	@if [ -z "${GOLINTER}" ]; then \
		echo "Cannot find 'golangci-lint' in your $$PATH"; \
		exit 1; \
	fi
	@echo "Running linter..."
	@$(GOLINTER) run ./...

nil:
	@if [ -z "${GONILAWAY}" ]; then \
		echo "Cannot find 'nilaway' in your $$PATH"; \
		exit 1; \
	fi
	@echo "Running nilaway..."
	@$(GONILAWAY) ./...