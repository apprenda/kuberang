# Set the build version
ifeq ($(origin VERSION), undefined)
	VERSION := $(shell git describe --tags --always --dirty)
endif

# build date
ifeq ($(origin BUILD_DATE), undefined)
	BUILD_DATE := $(shell date -u)
endif

# Setup some useful vars
HOST_GOOS = $(shell go env GOOS)
HOST_GOARCH = $(shell go env GOARCH)
GLIDE_VERSION = v0.11.1
ifeq ($(origin GLIDE_GOOS), undefined)
	GLIDE_GOOS := $(HOST_GOOS)
endif

BUILD_FLAGS = "-X main.version=$(VERSION) -X 'main.buildDate=$(BUILD_DATE)'"

build: vendor
	go build -o bin/kuberang -ldflags $(BUILD_FLAGS) ./cmd
	GOOS=darwin go build -o bin/darwin/$(HOST_GOARCH)/kuberang -ldflags $(BUILD_FLAGS) ./cmd
	GOOS=linux go build -o bin/linux/$(HOST_GOARCH)/kuberang -ldflags $(BUILD_FLAGS) ./cmd

docker: vendor
	CGO_ENABLED=0 GOOS=linux go build -a -o docker/kuberang -ldflags $(BUILD_FLAGS) ./cmd
	docker build --rm -t kuberang:$(VERSION) docker/

clean:
	rm -rf bin
	rm -rf out
	rm -rf vendor

test: vendor
	go test ./cmd/... ./pkg/... $(TEST_OPTS)

vendor: tools/glide
	./tools/glide install

tools/glide:
	mkdir -p tools
	curl -L https://github.com/Masterminds/glide/releases/download/$(GLIDE_VERSION)/glide-$(GLIDE_VERSION)-$(GLIDE_GOOS)-$(HOST_GOARCH).tar.gz | tar -xz -C tools
	mv tools/$(GLIDE_GOOS)-$(HOST_GOARCH)/glide tools/glide
	rm -r tools/$(GLIDE_GOOS)-$(HOST_GOARCH)

