## Update Versions here. If you bump to a new version you need to do make
## generate and commit the new files
PAC_VERSION := 0.14.3
TKN_VERSION := 0.28.0

GO := go
GOVERSION := 1.18
OPC_VERSION := devel
BINARYNAME := opc
GOLANGCI_LINT := golangci-lint

all: vendor generate build

vendor: tidy
	$(GO) mod vendor

mkbin:
	mkdir -p ./bin

build: mkbin
	$(GO) build -v -o bin/$(BINARYNAME) main.go

generate: version-file version-updates
version-file:
	echo '{"pac": "$(PAC_VERSION)", "tkn": "$(TKN_VERSION)", "opc": "$(OPC_VERSION)"}' > pkg/version.json

version-updates:
	go get -u github.com/openshift-pipelines/pipelines-as-code@v$(PAC_VERSION)
	go get -u github.com/tektoncd/cli@v$(TKN_VERSION)

tidy:
	$(GO) mod tidy -compat=$(GOVERSION)

lint-go: ## runs go linter on all go files
	@echo "Linting go files..."
	@$(GOLANGCI_LINT) run ./... --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--deadline 5m

.PHONY: generate version-file version-updates updates build all vendor tidy lint-go mkbin
