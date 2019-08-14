BINARY_NAME=wso-backend
DOCKER_TAG=wso-backend

.PHONY: build-docs
build-docs:
	swag init -g server/router.go
	goimports -w docs/docs.go

.PHONY: fmt
fmt:
	goimports -w ./

.PHONY: commit
commit: build-docs fmt

build:
	go build -tags=jsoniter -o $(BINARY_NAME) ./server/cmd

.PHONY: run
run: build
	./$(BINARY_NAME)

.PHONY: run-dev
run-dev: build
	./$(BINARY_NAME) --development

.PHONY: test
test:
	go test -race ./...

.PHONY: mod
mod:
	go mod tidy
	go mod download

.PHONY: docker-builder
docker-builder:
	docker build -t $(DOCKER_TAG)/builder -f Dockerfile.builder .
	docker build -t $(DOCKER_TAG)/builder-cert -f Dockerfile.cert .

.PHONY: docker-build-dev
docker-build-dev: docker-builder
	docker build -t $(DOCKER_TAG):dev-latest -f Dockerfile.release .
	docker build -t $(DOCKER_TAG)-jobs:dev-latest -f Dockerfile.release_jobs .

.PHONY: docker-build-rel-dev
docker-build-rel-dev: docker-builder
	docker build -t $(DOCKER_TAG):dev-latest -f Dockerfile.release .

.PHONY: docker-build-jobs-dev
docker-build-jobs-dev: docker-builder
	docker build -t $(DOCKER_TAG)-jobs:dev-latest -f Dockerfile.release_jobs .

.PHONY: k8-apply-dev
k8-apply-dev:
	kubectl apply -k k8s/development
	echo "Backend Service IP:" $(minikube service backend -n development --url)

.PHONY: k8-delete-dev
k8-delete-dev:
	kubectl delete -k k8s/development