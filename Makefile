# Build settings
# Override: make build OS=linux ARCH=amd64

APP        ?= bench
CMD_PKG    ?= ./cmd/bench
OUT_DIR    ?= bin

OS         ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH_RAW   ?= $(shell uname -m)

# Map uname arch to Go arch
ifeq ($(ARCH_RAW),x86_64)
ARCH ?= amd64
else ifeq ($(ARCH_RAW),amd64)
ARCH ?= amd64
else ifeq ($(ARCH_RAW),arm64)
ARCH ?= arm64
else ifeq ($(ARCH_RAW),aarch64)
ARCH ?= arm64
else
ARCH ?= $(ARCH_RAW)
endif

# Map uname OS to Go OS
ifeq ($(OS),darwin)
GOOS ?= darwin
else ifeq ($(OS),linux)
GOOS ?= linux
else ifeq ($(OS),windows)
GOOS ?= windows
else
GOOS ?= $(OS)
endif

GOARCH ?= $(ARCH)

BIN_EXT :=
ifeq ($(GOOS),windows)
BIN_EXT := .exe
endif

BIN := $(OUT_DIR)/$(APP)-$(GOOS)-$(GOARCH)$(BIN_EXT)

.PHONY: help
help:
	@echo "Targets:" \
	 && echo "  make tidy        - go mod tidy" \
	 && echo "  make fmt         - gofmt all" \
	 && echo "  make test        - go test ./..." \
	 && echo "  make build       - build for OS/ARCH (default: host)" \
	 && echo "  make clean       - remove $(OUT_DIR)/" \
	 && echo "" \
	 && echo "Variables:" \
	 && echo "  OS=<linux|darwin|windows>" \
	 && echo "  ARCH=<amd64|arm64>" \
	 && echo "  GOOS/GOARCH can also be set directly" \
	 && echo "" \
	 && echo "Example:" \
	 && echo "  make build OS=linux ARCH=amd64"

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.github/*')

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "-s -w" -o $(BIN) $(CMD_PKG)
	@echo "Built $(BIN)"

.PHONY: clean
clean:
	rm -rf $(OUT_DIR)
