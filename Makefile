GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

BINARY_NAME=wso-go

.PHONY: build-docs
build-docs:
	swag init

.PHONY: build
build: build-docs
	$(GOBUILD) -tags=jsoniter -o $(BINARY_NAME) main.go

run: build
	./$(BINARY_NAME)

.PHONY: test
test:
	$(GOTEST) ./...

mod:
	$(GOCMD) mod tidy
	$(GOCMD) mod download
	$(GOCMD) mod vendor