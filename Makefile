PAC_VERSION := $(shell sed -n '/[ ]*github.com\/openshift-pipelines\/pipelines-as-code v[0-9]*\.[0-9]*\.[0-9]*/ { s/.* v//;p ;}' go.mod)
TKN_VERSION := $(shell sed -n '/[ ]*github.com\/tektoncd\/cli v[0-9]*\.[0-9]*\.[0-9]*/ { s/.* v//;p ;}' go.mod)
RESULTS_VERSION := $(shell sed -n '/[ ]*github.com\/tektoncd\/results v[0-9]*\.[0-9]*\.[0-9]*/ { s/.* v//;p ;}' go.mod)
MAG_VERSION := $(shell sed -n '/[ ]*github.com\/openshift-pipelines\/manual-approval-gate v[0-9]*\.[0-9]*\.[0-9]*/ { s/.* v//;p ;}' go.mod)

GO := go
GOVERSION := 1.22
OPC_VERSION := devel
BINARYNAME := opc
GOLANGCI_LINT := golangci-lint

FLAGS := -ldflags "-X github.com/tektoncd/cli/pkg/cmd/version.clientVersion=$(TKN_VERSION) \
		   -X github.com/openshift-pipelines/pipelines-as-code/pkg/params/version.Version=$(PAC_VERSION) \
		   -X github.com/openshift-pipelines/pipelines-as-code/pkg/params/settings.TknBinaryName=$(BINARYNAME)" $(LDFLAGS)

all: build

vendor: tidy
	$(GO) mod vendor

mkbin: # makes bin directory
	mkdir -p ./bin

build: mkbin generate ## builds binary and updates version in pkg/version
	$(GO) build -v $(FLAGS) -mod=vendor -o bin/$(BINARYNAME) main.go

windows: mkbin generate
	env GOOS=windows GOARCH=amd64 $(GO) build -mod=vendor $(FLAGS)  -v -o bin/$(BINARYNAME).exe main.go

generate: version-file ## updates version of pipeline-as-code, cli, mag and results in pkg/version file
version-file:
	echo '{"pac": "$(PAC_VERSION)", "tkn": "$(TKN_VERSION)", "results": "$(RESULTS_VERSION)", "manualapprovalgate": "$(MAG_VERSION)", "opc": "$(OPC_VERSION)"}' > pkg/version.json

version-updates: ## updates pipeline-as-code, cli, mag and results version in go.mod
	$(GO) get -u github.com/openshift-pipelines/pipelines-as-code
	$(GO) mod vendor
	$(GO) get -u github.com/openshift-pipelines/manual-approval-gate
	$(GO) mod vendor
	$(GO) get -u github.com/tektoncd/cli
	$(GO) mod vendor
	$(GO) get -u github.com/tektoncd/results
	$(GO) mod vendor
	$(GO) mod tidy

tidy:
	$(GO) mod tidy -compat=$(GOVERSION)

lint-go: ## runs go linter on all go files
	@echo "Linting go files..."
	@$(GOLANGCI_LINT) run ./... --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--timeout 10m

test: ## runs unit tests
	@echo "Running unit tests..."
	@$(GO) test -v -mod=vendor ./...

.PHONY: generate version-file version-updates updates build all vendor tidy lint-go test mkbin

.PHONY: help
help: ## print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {gsub("\\\\n",sprintf("\n%22c",""), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
