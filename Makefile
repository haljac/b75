# Binary name
BINARY_NAME=b75
MAIN_PATH=cmd/b75/main.go

# Go related variables.
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

.PHONY: all build run dev test clean vet install

all: build

# Build the project
build:
	@echo "  >  Building binary..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run the project (standard environment)
run: build
	@echo "  >  Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run in "dev" mode with a local data directory
# This prevents polluting your actual ~/.local/share/b75 directory while developing
dev: build
	@echo "  >  Running $(BINARY_NAME) in dev mode (data stored in ./tmp_data)..."
	XDG_DATA_HOME=$(GOBASE)/tmp_data ./$(BINARY_NAME)

# Run tests
test:
	@echo "  >  Running tests..."
	go test -v ./...

# Run go vet
vet:
	@echo "  >  Running go vet..."
	go vet ./...

# Clean build artifacts and temp data
clean:
	@echo "  >  Cleaning..."
	go clean
	rm -f $(BINARY_NAME)
	rm -rf tmp_data

# Install to $GOPATH/bin
install:
	@echo "  >  Installing..."
	go install $(MAIN_PATH)
