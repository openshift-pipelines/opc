PAC_VERSION := $(shell sed -n '/[ ]*github.com\/openshift-pipelines\/pipelines-as-code v[0-9]*\.[0-9]*\.[0-9]*/ { s/.* v//;p ;}' go.mod)
TKN_VERSION := $(shell sed -n '/[ ]*github.com\/tektoncd\/cli v[0-9]*\.[0-9]*\.[0-9]*/ { s/.* v//;p ;}' go.mod)

GO := go
GOVERSION := 1.18
OPC_VERSION := devel
BINARYNAME := opc
GOLANGCI_LINT := golangci-lint

FLAGS := -ldflags "-X github.com/tektoncd/cli/pkg/cmd/version.clientVersion=$(TKN_VERSION) \
		   -X github.com/openshift-pipelines/pipelines-as-code/pkg/params/version.Version=$(PAC_VERSION) \
		   -X github.com/openshift-pipelines/pipelines-as-code/pkg/params/settings.TknBinaryName=$(BINARYNAME)" $(LDFLAGS)

all: build

vendor: tidy
	$(GO) mod vendor

mkbin:
	mkdir -p ./bin

build: mkbin generate
	$(GO) build -v $(FLAGS) -mod=vendor -o bin/$(BINARYNAME) main.go

windows: mkbin generate
	env GOOS=windows GOARCH=amd64 $(GO) build -mod=vendor $(FLAGS)  -v -o bin/$(BINARYNAME).exe main.go

generate: version-file
version-file:
	echo '{"pac": "$(PAC_VERSION)", "tkn": "$(TKN_VERSION)", "opc": "$(OPC_VERSION)"}' > pkg/version.json

version-updates:
	$(GO) get -u github.com/openshift-pipelines/pipelines-as-code
	$(GO) mod vendor
	$(GO) get -u github.com/tektoncd/cli
	$(GO) mod vendor
	$(GO) mod tidy

tidy:
	$(GO) mod tidy -compat=$(GOVERSION)

lint-go: ## runs go linter on all go files
	@echo "Linting go files..."
	@$(GOLANGCI_LINT) run ./... --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--deadline 10m

.PHONY: generate version-file version-updates updates build all vendor tidy lint-go mkbin
