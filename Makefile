VERSION := $(shell git describe --long --tags --dirty 2> /dev/null || echo "dev")
GOLANGCI_LINT_IMG := golangci/golangci-lint:v1.55-alpine

ARCH           ?= amd64
OS             ?= linux
PLATFORM       ?= $(OS)/$(ARCH)
REGISTRY_NAME  ?= index.docker.io/linode
IMAGE_NAME     ?= k8s-node-decorator
IMAGE_VERSION  ?= $(VERSION)
IMAGE_TAG      ?= $(REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_VERSION)

.PHONY: lint
lint:
	docker run --rm -v $(PWD):/app -w /app ${GOLANGCI_LINT_IMG} golangci-lint run -v

.PHONY: build
build:
	go build -o ./bin/k8s-node-decorator -a -ldflags '-X main.version='${VERSION}' -extldflags "-static"' ./main.go

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build --platform=$(PLATFORM) --progress=plain -t $(IMAGE_TAG) --build-arg VERSION=$(VERSION) -f ./Dockerfile .

.PHONY: fmt
fmt:
	gofumpt -w -l .

