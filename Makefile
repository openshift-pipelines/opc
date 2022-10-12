GO := go
GOVERSION := 1.18
BINARYNAME := opc
GOLANGCI_LINT=golangci-lint

all: build

vendor: tidy
	$(GO) mod vendor

.PHONY: build
build: vendor mkbin
	$(GO) build -o bin/$(BINARYNAME) main.go

.PHONY: tidy
tidy:
	$(GO) mod tidy -compat=$(GOVERSION)

.PHONY: lint-go
lint-go: ## runs go linter on all go files
	@echo "Linting go files..."
	@$(GOLANGCI_LINT) run ./... --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--deadline 5m


mkbin:
	mkdir -p ./bin
