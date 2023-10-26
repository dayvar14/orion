# Makefile

# Define the output directory for the binaries
OUTPUT_DIR := bin

# List of Go files in the cmd directory
CMD_FILES := $(wildcard cmd/**/*.go)

# Targets
all: build

deps:
	@go mod tidy

build: $(CMD_FILES)
	@mkdir -p $(OUTPUT_DIR)
	@go build -o $(OUTPUT_DIR)/ $(CMD_FILES)
	@echo "Binaries built and stored in $(OUTPUT_DIR)"

build/pi: $(CMD_FILES)
	@mkdir -p $(OUTPUT_DIR)
	@env GOOS=linux GOARCH=arm GOARM=5 go build -o $(OUTPUT_DIR)/ $(CMD_FILES)
	@echo "Binaries built and stored in $(OUTPUT_DIR)"

clean:
	@rm -rf $(OUTPUT_DIR)
	@echo "$(OUTPUT_DIR) directory removed"

.PHONY: all build clean
