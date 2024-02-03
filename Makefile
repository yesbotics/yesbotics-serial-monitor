SHELL:=/bin/bash

.ONESHELL:

APP_VERSION ?= "0.1.0"
APP_EXECUTABLE ?= "ysm"

BUILD_DIR = ./bin
BUILD_COMMAND = CGO_ENABLED=0 go build -ldflags="-s -w"

all: build

build: FORCE setup build-linux build-mac build-windows tar-gz

build-windows: FORCE
	GOOS=windows GOARCH=amd64 $(BUILD_COMMAND) -o $(BUILD_DIR)/$(APP_EXECUTABLE)_windows-amd64/$(APP_EXECUTABLE).exe .
	GOOS=windows GOARCH=386 $(BUILD_COMMAND) -o $(BUILD_DIR)/$(APP_EXECUTABLE)_windows-386/$(APP_EXECUTABLE).exe .

build-linux: FORCE
	GOOS=linux GOARCH=amd64 ${BUILD_COMMAND} -o $(BUILD_DIR)/$(APP_EXECUTABLE)_linux-amd64/$(APP_EXECUTABLE) .
	GOOS=linux GOARCH=arm $(BUILD_COMMAND) -o $(BUILD_DIR)/$(APP_EXECUTABLE)_linux-arm/$(APP_EXECUTABLE) .
	GOOS=linux GOARCH=arm64 $(BUILD_COMMAND) -o $(BUILD_DIR)/$(APP_EXECUTABLE)_linux-arm64/$(APP_EXECUTABLE) .

build-mac: FORCE
	GOOS=darwin GOARCH=amd64 $(BUILD_COMMAND) -o $(BUILD_DIR)/$(APP_EXECUTABLE)_darwin-amd64/$(APP_EXECUTABLE) .
	GOOS=darwin GOARCH=arm64 $(BUILD_COMMAND) -o $(BUILD_DIR)/$(APP_EXECUTABLE)_darwin-arm64/$(APP_EXECUTABLE) .

build-android: FORCE
	GOOS=android GOARCH=arm $(BUILD_COMMAND) -o $(BUILD_DIR)/$(APP_EXECUTABLE)_android-arm/$(APP_EXECUTABLE) .
	GOOS=android GOARCH=arm64 $(BUILD_COMMAND) -o $(BUILD_DIR)/$(APP_EXECUTABLE)_android-arm64/$(APP_EXECUTABLE) .

zip: FORCE
	cd $(BUILD_DIR)
	@for d in */ ; do \
        zip -r "$${d%/}.zip" "$$d"; \
    done

tar-gz: FORCE
	cd $(BUILD_DIR)
	@for d in */ ; do \
		tar -czvf "$${d%/}.tar.gz" "$$d"; \
		sha256sum "$${d%/}.tar.gz" > "$${d%/}.tar.gz.sha256"; \
	done

install: FORCE build
	go install .

setup: FORCE
	mkdir -p $(BUILD_DIR)
	go mod vendor
	go mod tidy

update: FORCE
	go get -u
	go mod vendor
	go mod tidy


test: FORCE

dev: FORCE
	DEBUG=1 go run .

run: FORCE dev

log: FORCE
	tail -f debug.log


.PHONY: FORCE
FORCE:
