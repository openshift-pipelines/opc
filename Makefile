## Update Versions here. If you bump to a new version you need to do make
## generate and commit the new files
PAC_VERSION := 0.15.3
TKN_VERSION := 0.29.0

GO := go
GOVERSION := 1.18
OPC_VERSION := devel
BINARYNAME := opc
GOLANGCI_LINT := golangci-lint

FLAGS := -ldflags "-X github.com/tektoncd/cli/pkg/cmd/version.clientVersion=$(TKN_VERSION) \
		   -X github.com/openshift-pipelines/pipelines-as-code/pkg/params/version.Version=$(PAC_VERSION) \
		   -X github.com/openshift-pipelines/pipelines-as-code/pkg/params/settings.TknBinaryName=$(BINARYNAME)" $(LDFLAGS)

all: vendor generate build

vendor: tidy
	$(GO) mod vendor

mkbin:
	mkdir -p ./bin

build: mkbin
	$(GO) build -v $(FLAGS) -mod=vendor -o bin/$(BINARYNAME) main.go

windows: mkbin
	env GOOS=windows GOARCH=amd64 $(GO) build -mod=vendor $(FLAGS)  -v -o bin/$(BINARYNAME).exe main.go

generate: version-file version-updates
version-file:
	echo '{"pac": "$(PAC_VERSION)", "tkn": "$(TKN_VERSION)", "opc": "$(OPC_VERSION)"}' > pkg/version.json

version-updates:
	$(GO) get -u github.com/openshift-pipelines/pipelines-as-code@v$(PAC_VERSION)
	$(GO) mod vendor
	$(GO) get -u github.com/tektoncd/cli@v$(TKN_VERSION)
	$(GO) mod vendor

tidy:
	$(GO) mod tidy -compat=$(GOVERSION)

lint-go: ## runs go linter on all go files
	@echo "Linting go files..."
	@$(GOLANGCI_LINT) run ./... --modules-download-mode=vendor \
							--max-issues-per-linter=0 \
							--max-same-issues=0 \
							--deadline 10m

.PHONY: generate version-file version-updates updates build all vendor tidy lint-go mkbin
