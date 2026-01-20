.PHONY: build clean

BINARY_NAME := tpsp
BIN_DIR := bin

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/tpsp

clean:
	rm -rf $(BIN_DIR)/tpsp
