GO := go
GOVERSION := 1.17
BINARYNAME := opc

all: build

build: vendor mkbin
	$(GO) build -o bin/$(BINARYNAME) main.go

tidy:
	$(GO) mod tidy -compat=$(GOVERSION)

vendor: tidy
	$(GO) mod vendor

mkbin:
	mkdir -p ./bin
