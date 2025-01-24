# Recursive wildcard
rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))

BINARY_NAME=wso-backend
DOCKER_TAG=wso-backend
GIT_REPO=github.com/WilliamsStudentsOnline/wso-go
BUILD_DIRS = config db lib models server services sanitize
BUILD_DEPS = $(call rwildcard, $(BUILD_DIRS), *.go) $(wildcard jobs/*/*.go) jobs/dorms_update/cmd/data.go docs/docs.go
SERVICE_DIRS = $(wildcard services/*)
SWAGGER := $(shell which swag 2>/dev/null)
GOIMPORTS := $(shell which goimports 2>/dev/null)

define GOIMPORTS_ERROR
goimports command is missing.

GoImports Installation Instructions
---
Run:
  go get golang.org/x/tools/cmd/goimports

Ensure your go bin is in your $$PATH.
Edit your ~/.bashrc to add this line:
  export PATH=$$PATH:$$(go env GOPATH)/bin

endef

$(BINARY_NAME): $(BUILD_DEPS)
	go build -tags=jsoniter -o $(BINARY_NAME) ./server/cmd

jobs/dorms_update/cmd/data.go: $(wildcard jobs/dorms_update/data/*) jobs/dorms_update/cmd/gen.go
	go generate $(GIT_REPO)/jobs/dorms_update/cmd

docs docs/docs.go docs/swagger.json docs/swagger.yaml: $(wildcard models/*.go) $(wildcard services/**/*.go) $(wildcard services/**/**/*.go) server/router.go
ifdef SWAGGER
	swag init -g server/router.go
endif
ifdef GOIMPORTS
	goimports -w docs/docs.go
else
	$(warning $(GOIMPORTS_ERROR))
endif

services/*/responses/%.go: services/*/responses/%.json
	go run $(GIT_REPO)/lib/generate/service_responses/cmd -in $< -out $@

services/words/words_data.go: services/words/words.json
	go generate $(GIT_REPO)/services/words

.PHONY: job-catalog-update
job-catalog-update:
	go build -tags=jsoniter -o job-catalog-update ./jobs/catalog_update/cmd

.PHONY: job-update-all-factrak-survey-deficits
job-update-all-factrak-survey-deficits:
	go build -tags=jsoniter -o job-update-all-factrak-survey-deficits ./jobs/update_all_factrak_survey_deficits/cmd

.PHONY: job-update-all-users-from-ldap
job-update-all-users-from-ldap:
	go build -tags=jsoniter -o job-update-all-users-from-ldap ./jobs/update_all_users_from_ldap/cmd

.PHONY: job-dorms-update
job-dorms-update:
	go build -tags=jsoniter -o job-dorms-update ./jobs/dorms_update/cmd

.PHONY: job-frosh-photos
job-frosh-photos:
	go build -tags=jsoniter -o job-frosh-photos ./jobs/frosh_photos/cmd

.PHONY: job-user-csv-data
job-user-csv-data:
	go build -tags=jsoniter -o job-user-csv-data ./jobs/user_csv_data/cmd

.PHONY: job-dining-update
job-dining-update:
	go build -tags=jsoniter -o job-dining-update ./jobs/dining_update/cmd

PHONY: job-schedule-notifs
job-schedule-notifs:
	go build -tags=jsoniter -o job-schedule-notifs ./jobs/schedule_notifs/cmd

.PHONY: job-update-on-campus-semesters
job-update-on-campus-semesters:
	go build -tags jsoniter -o job-increment-oncampus-semesters ./jobs/update_on_campus_semesters/cmd

.PHONY: job-initialize-on-campus-semesters
job-initialize-on-campus-semesters:
	go build -tags jsoniter -o job-increment-oncampus-semesters ./jobs/update_on_campus_semesters/initial-calculation

.PHONY: job-update_profs_areas_of_study
job-update_profs_areas_of_study:
	go build -tags jsoniter -o job-update_profs_areas_of_study ./jobs/update_profs_areas_of_study/cmd

.PHONY: job-ephmatch-reset
job-ephmatch-reset:
	go build -tags jsoniter -o job-ephmatch-reset ./jobs/ephmatch_reset

.PHONY: build-prod-linux
build-prod-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o $(BINARY_NAME)_linux ./server/cmd

.PHONY: build-jobs-prod-linux
build-jobs-prod-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-catalog-update_linux ./jobs/catalog_update/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-update-all-factrak-survey-deficits_linux ./jobs/update_all_factrak_survey_deficits/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-update-all-users-from-ldap_linux ./jobs/update_all_users_from_ldap/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-dorms-update_linux ./jobs/dorms_update/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-frosh-photos_linux ./jobs/frosh_photos/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-user-csv-data_linux ./jobs/user_csv_data/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-dining-update_linux ./jobs/dining_update/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-schedule-notifs_linux ./jobs/schedule_notifs/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-update-on-campus-semesters ./jobs/update_on_campus_semesters/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-initialize-on-campus-semesters ./jobs/update_on_campus_semesters/initial-calculation
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-update_profs_areas_of_study ./jobs/update_profs_areas_of_study/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-ephmatch-reset ./jobs/ephmatch_reset

.PHONY: go-gen
go-gen:
	go generate $(GIT_REPO)/...

.PHONY: fmt
fmt:
ifdef GOIMPORTS
	goimports -w ./
else
	$(error $(GOIMPORTS_ERROR))
endif

.PHONY: commit
commit: jobs/dorms_update/cmd/data.go docs/docs.go fmt services/*/responses/*.go

.PHONY: run-dev
run-dev: $(BINARY_NAME)
	./$(BINARY_NAME) --development

.PHONY: run-with-analytics
run-with-analytics:
	./$(BINARY_NAME) --development
	bash ./prod_files/run-analytics.sh

.PHONY: test
test:
	go test -race ./...

.PHONY: fast-test
fast-test:
	go test ./...

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
