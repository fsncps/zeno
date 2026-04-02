APP=zeno
BUILD_DIR=build
CONFIG_DIR=$(HOME)/.config/zeno

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
	mkdir -p $(CONFIG_DIR)
	if [ ! -f $(CONFIG_DIR)/config.yaml ]; then \
		install -m644 config.default.yaml $(CONFIG_DIR)/config.yaml; \
	fi
	if [ ! -f $(CONFIG_DIR)/.env ]; then \
		install -m600 env.example $(CONFIG_DIR)/.env; \
	fi
	@echo "Installed config files to $(CONFIG_DIR)"
	@echo "Edit $(CONFIG_DIR)/.env with your database credentials"