BINARY_NAME=wso-backend

.PHONY: build-docs
build-docs:
	swag init -g server/router.go

.PHONY: build
build: build-docs
	go build -tags=jsoniter -o $(BINARY_NAME) ./server/cmd

run: build
	./$(BINARY_NAME)

run-dev: build
	./$(BINARY_NAME) --development

.PHONY: test
test:
	go test -race ./...

.PHONY: mod
mod:
	go mod tidy
	go mod download

docker-builder:
	docker build -t wso-backend/builder -f Dockerfile.builder

.PHONY: docker-build-dev
docker-build-dev: docker-builder
	docker build -t wso-backend:dev-latest -f Dockerfile.release .
