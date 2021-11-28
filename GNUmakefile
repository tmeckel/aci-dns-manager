GO ?= go
GOFMT ?= gofumpt -l -d -extra
GOFILES := $(shell find . -name "*.go" -type f)
VERSION := $(shell git describe --tags | sed 's/^v//')
SHORT_COMMIT := $(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-s -w -X=main.version=$(VERSION) -X=main.commit=$(SHORT_COMMIT)"
DIST_FOLDER = dist

ifeq ($(GOPATH),)
	GOPATH:=$(shell go env GOPATH)
endif

.EXPORT_ALL_VARIABLES:
  GO111MODULE=on

.PHONY: default
default: build

.PHONY: tools
tools:
	@echo "==> installing required tooling..."
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(GOPATH)/bin"

.PHONY: build
build:
	@mkdir -p $(DIST_FOLDER)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a $(LDFLAGS) -o $(DIST_FOLDER)/aci-dns-manager cmd/main.go

.PHONY: fmt
fmt:
	@hash gofumpt > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install mvdan.cc/gofumpt@v0.2.0; \
	fi
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@echo "==> Checking source code formatting ..."
	@hash gofumpt > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install mvdan.cc/gofumpt@v0.2.0; \
	fi
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: clean
clean:
	rm -rf $(DIST_FOLDER)

.PHONY: docker-lint
docker-lint:
	docker run --pull always --rm -v ${PWD}/docker/hadolint.yml:/.config/hadolint.yaml -i ghcr.io/hadolint/hadolint < docker/Dockerfile

.PHONY: docker-image
docker-image: build docker-lint
	docker build -t aci-dns-manager:latest -f docker/Dockerfile $(DIST_FOLDER)

.PHONY: docker-scan
docker-scan: docker-image
	docker run --rm -ti -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/aquasecurity/trivy:latest i aci-dns-manager:latest

.PHONY: lint
lint: fmt-check
	@echo "==> Checking source code against linters..."
	@if command -v golangci-lint; then (golangci-lint run ./...); else ($(GOPATH)/bin/golangci-lint run ./...); fi
