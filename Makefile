# Recursive wildcard
rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))

BINARY_NAME=wso-backend
DOCKER_TAG=wso-backend
GIT_REPO=github.com/WilliamsStudentsOnline/wso-go
BUILD_DIRS = config db lib models server services sanitize
BUILD_DEPS = $(call rwildcard, $(BUILD_DIRS), *.go) jobs/jobs.go $(wildcard jobs/*/*.go) jobs/dorms_update/cmd/data.go docs/docs.go
SERVICE_DIRS = $(wildcard services/*)
SWAGGER := $(shell which swag 2>/dev/null)
GOIMPORTS := $(shell which goimports 2>/dev/null)

$(BINARY_NAME): $(BUILD_DEPS)
	go build -tags=jsoniter -o $(BINARY_NAME) ./server/cmd

jobs/dorms_update/cmd/data.go: $(wildcard jobs/dorms_update/data/*) jobs/dorms_update/cmd/gen.go
	go generate $(GIT_REPO)/jobs/dorms_update/cmd

docs/docs.go docs/swagger.json docs/swagger.yaml: $(wildcard models/*.go) $(wildcard services/**/*.go) server/router.go
ifdef SWAGGER
	swag init -g server/router.go
endif
ifdef GOIMPORTS
	goimports -w docs/docs.go
endif

services/*/responses/%.go: services/*/responses/%.json
	go run $(GIT_REPO)/lib/generate/service_responses/cmd -in $< -out $@

services/words/words_data.go: services/words/words.json
	go generate $(GIT_REPO)/services/words

.PHONY: go-gen
go-gen:
	go generate $(GIT_REPO)/...

.PHONY: fmt
fmt:
	goimports -w ./

.PHONY: commit
commit: jobs/dorms_update/cmd/data.go docs/docs.go fmt services/*/responses/*.go

.PHONY: run-dev
run-dev: $(BINARY_NAME)
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

.PHONY: docker-dev
docker-dev: docker-builder
	docker build -t $(DOCKER_TAG):dev-latest -f Dockerfile.release .
	docker build -t $(DOCKER_TAG)-jobs:dev-latest -f Dockerfile.release_jobs .

.PHONY: docker-rel-dev
docker-rel-dev: docker-builder
	docker build -t $(DOCKER_TAG):dev-latest -f Dockerfile.release .

.PHONY: docker-jobs-dev
docker-jobs-dev: docker-builder
	docker build -t $(DOCKER_TAG)-jobs:dev-latest -f Dockerfile.release_jobs .

.PHONY: k8-apply-dev
k8-apply-dev:
	kubectl apply -k k8s/development
	minikube service backend -n development --url

.PHONY: k8-delete-dev
k8-delete-dev:
	kubectl delete -k k8s/development

.PHONY: k8-restart-backend-dev
k8-restart-backend-dev:
	kubectl -n development delete deployments.apps backend
	kubectl -n development apply -k k8s/development
