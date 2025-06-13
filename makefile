# Configuration
BIN_NAME      := kontext # 二进制名称变量化
INSTALL_DIR   := /usr/local/bin
LOCAL_BIN     := $(HOME)/.local/bin

.PHONY: all build clean sudo-install user-install

all: build

build:
	@echo "Building Go binary..."
	go build -o $(BIN_NAME)

sudo-install: build
	@echo "Installing system-wide (requires sudo)"
	sudo mkdir -p $(INSTALL_DIR)
	sudo cp $(BIN_NAME) $(INSTALL_DIR)/
	@echo "Installed to $(INSTALL_DIR)/$(BIN_NAME)"

user-install: build
	@echo "Installing to user directory"
	mkdir -p $(LOCAL_BIN)
	cp $(BIN_NAME) $(LOCAL_BIN)/
	@echo "Installed to $(LOCAL_BIN)/$(BIN_NAME)"
	@echo "NOTE: Ensure $(LOCAL_BIN) is in your PATH"

clean:
	@echo "Full cleanup..."
	rm -f $(BIN_NAME)
