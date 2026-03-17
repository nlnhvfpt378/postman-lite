APP_NAME=postman-lite
ROOT:=$(shell pwd)
GO?=go
BIN_DIR=$(ROOT)/deliverables/bin
DIST_DIR=$(ROOT)/deliverables
VERSION?=0.3.0

.PHONY: deps build run package-cross clean

deps:
	$(GO) mod tidy

build: deps
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=1 $(GO) build -o $(BIN_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)

run: build
	$(BIN_DIR)/$(APP_NAME)

package-cross: deps
	VERSION=$(VERSION) GO_BIN=$(GO) bash ./build/package-cross.sh

clean:
	rm -rf $(BIN_DIR) $(ROOT)/build/dist \
		$(DIST_DIR)/$(APP_NAME) $(DIST_DIR)/$(APP_NAME).exe \
		$(DIST_DIR)/$(APP_NAME)_$(VERSION)_linux_amd64.tar.gz \
		$(DIST_DIR)/$(APP_NAME)_$(VERSION)_windows_amd64.zip
