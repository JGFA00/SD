# Use bash explicitly
SHELL := /bin/bash

# Variables
GO_FILES := peer.go poisson.go
APP_NAME := peer_app
HOST_FILE := hosts.txt
LOG_DIR := logs

# Default target
all: build run

# Build the Go application
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(GO_FILES)

# Run all peers
run: build
	@echo "Starting all peers..."
	@mkdir -p logs # Create log directory if it doesn't exist
	@while read -r line || [ -n "$$line" ]; do \
		if [ "$$line" != "" ] && [ $${line:0:1} != "#" ]; then \
			set -- $$line; \
			PEER_NAME=$$1; \
			HOST_PORT=$$2; \
			shift 2; \
			NEIGHBORS=$$@; \
			if [ -z "$$HOST_PORT" ] || [ -z "$$NEIGHBORS" ]; then \
				echo "Skipping invalid line: $$line"; \
				continue; \
			fi; \
			echo "Starting $$PEER_NAME on $$HOST_PORT with neighbors: $$NEIGHBORS"; \
			./$(APP_NAME) $$HOST_PORT $$NEIGHBORS > logs/$$PEER_NAME.log 2>&1 & \
		fi; \
	done < $(HOST_FILE)
	@echo "All peers started. Logs are available in the logs directory."

# Clean up the built application and logs
clean:
	@echo "Cleaning up..."
	rm -f $(APP_NAME)
	rm -rf $(LOG_DIR)

# Stop all running peers
stop:
	@echo "Stopping all peers..."
	pkill -f $(APP_NAME) || true

.PHONY: all build run clean stop
