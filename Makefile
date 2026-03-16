APP_NAME=postman-lite
ROOT:=$(shell pwd)
GO=/home/node/clawd/.local/go/bin/go
BIN_DIR=$(ROOT)/deliverables/bin
DIST_DIR=$(ROOT)/deliverables
VERSION?=0.1.0

export CGO_ENABLED=0

.PHONY: deps build run package-deb package-cross package-all clean

deps:
	$(GO) mod tidy

build: deps
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

run: build
	$(BIN_DIR)/$(APP_NAME)

package-deb: build
	bash ./build/package-deb.sh $(VERSION)

package-cross: deps
	VERSION=$(VERSION) GO_BIN=$(GO) bash ./build/package-cross.sh

package-all: package-cross package-deb

clean:
	rm -rf $(BIN_DIR) $(ROOT)/build/deb $(ROOT)/build/dist \
		$(DIST_DIR)/$(APP_NAME) $(DIST_DIR)/$(APP_NAME).exe \
		$(DIST_DIR)/$(APP_NAME)_$(VERSION)_linux_amd64.tar.gz \
		$(DIST_DIR)/$(APP_NAME)_$(VERSION)_windows_amd64.zip
