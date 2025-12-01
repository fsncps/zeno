APP=zeno
BUILD_DIR=build

.PHONY: all build clean install

all: build

build:
	mkdir -p $(BUILD_DIR)
	go build -v -o $(BUILD_DIR)/$(APP) ./cmd/zeno

clean:
	rm -f $(BUILD_DIR)/$(APP)
	go clean -cache

install: build
	sudo install -m755 $(BUILD_DIR)/$(APP) /usr/local/bin/$(APP)

