##### WSO-Backend WSO 2.0 #####
### Variable definitions

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

# Publish docs/swagger.yaml as a GitHub Release asset (openapi-vYYYY.MM.DD[-N])
.PHONY: publish-api
publish-api:
	.github/scripts/publish-api.sh

# API contract: make api (regen) / make check-api (swag freshness + Spectral + oasdiff)
OPENAPI_BASE_REF ?= origin/master

.PHONY: api
api:
ifndef SWAGGER
	$(error swag not found; install with: go install github.com/swaggo/swag/cmd/swag@v1.16.3)
endif
	swag init -g server/router.go
ifdef GOIMPORTS
	goimports -w docs/docs.go
endif

.PHONY: check-api
check-api:
	OPENAPI_BASE_REF=$(OPENAPI_BASE_REF) .github/scripts/check-openapi.sh

services/*/responses/%.go: services/*/responses/%.json
	go run $(GIT_REPO)/lib/generate/service_responses/cmd -in $< -out $@

services/words/words_data.go: services/words/words.json
	go generate $(GIT_REPO)/services/words

### Job definitions
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

.PHONY: job-user-home-info
job-user-home-info:
	go build -tags=jsoniter -o job-user-home-info ./jobs/user_home_info/cmd

.PHONY: job-dining-update
job-dining-update:
	go build -tags=jsoniter -o job-dining-update ./jobs/dining_update/cmd

PHONY: job-schedule-notifs
job-schedule-notifs:
	go build -tags=jsoniter -o job-schedule-notifs ./jobs/schedule_notifs/cmd

.PHONY: job-update-on-campus-semesters
job-update-on-campus-semesters:
	go build -tags jsoniter -o job-update-on-campus-semesters ./jobs/update_on_campus_semesters/cmd

.PHONY: job-initialize-on-campus-semesters
job-initialize-on-campus-semesters:
	go build -tags jsoniter -o job-increment-on-campus-semesters ./jobs/update_on_campus_semesters/initial-calculation

.PHONY: job-update_profs_areas_of_study
job-update_profs_areas_of_study:
	go build -tags jsoniter -o job-update_profs_areas_of_study ./jobs/update_profs_areas_of_study/cmd

.PHONY: job-ephmatch-reset
job-ephmatch-reset:
	go build -tags jsoniter -o job-ephmatch-reset ./jobs/ephmatch_reset

.PHONY: job-ephmatch-update-dates
job-ephmatch-update-dates:
	go build -tags jsoniter -o job-ephmatch_update_dates ./jobs/ephmatch_update_dates

.PHONY: job-library-hours-update
 job-library-hours-update:
	go build -tags jsoniter -o job-library-hours-update ./jobs/job_library_hours_update

.PHONY: job-mobile-fetcher
job-mobile-fetcher:
	go build -tags=jsoniter -o job-mobile-fetcher ./jobs/mobile_fetcher/cmd

.PHONY: job-jobctl
job-jobctl:
	go build -tags=jsoniter -o wso-jobctl ./jobs/jobctl/cmd

### Build definitions
$(BINARY_NAME): $(BUILD_DEPS)
	go build -tags=jsoniter -o $(BINARY_NAME) ./server/cmd

.PHONY: build-dev
build: $(BINARY_NAME)

.PHONY: build-prod-linux
build-prod-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o $(BINARY_NAME)_linux ./server/cmd

.PHONY: build-jobs
build-jobs:
	go build  -tags=jsoniter -o job-catalog-update ./jobs/catalog_update/cmd
	go build  -tags=jsoniter -o job-update-all-factrak-survey-deficits ./jobs/update_all_factrak_survey_deficits/cmd
	go build  -tags=jsoniter -o job-update-all-users-from-ldap ./jobs/update_all_users_from_ldap/cmd
	go build  -tags=jsoniter -o job-dorms-update ./jobs/dorms_update/cmd
	go build  -tags=jsoniter -o job-frosh-photos ./jobs/frosh_photos/cmd
	go build  -tags=jsoniter -o job-user-csv-data ./jobs/user_csv_data/cmd
	go build  -tags=jsoniter -o job-user-home-info ./jobs/user_home_info/cmd
	go build  -tags=jsoniter -o job-dining-update ./jobs/dining_update/cmd
	go build  -tags=jsoniter -o job-schedule-notifs ./jobs/schedule_notifs/cmd
	go build  -tags=jsoniter -o job-update-on-campus-semesters ./jobs/update_on_campus_semesters/cmd
	go build -tags=jsoniter -o job-initialize-on-campus-semesters ./jobs/update_on_campus_semesters/initial-calculation
	go build -tags=jsoniter -o job-update_profs_areas_of_study ./jobs/update_profs_areas_of_study/cmd
	go build -tags=jsoniter -o job-ephmatch-reset ./jobs/ephmatch_reset
	go build -tags=jsoniter -o job-ephmatch_update_dates ./jobs/ephmatch_update_dates
	go build -tags jsoniter -o job-library-hours-update ./jobs/library_hours_update/cmd
	go build -tags=jsoniter -o job-mobile-fetcher ./jobs/mobile_fetcher/cmd
	go build -tags=jsoniter -o wso-jobctl ./jobs/jobctl/cmd

.PHONY: build-jobs-prod-linux
build-jobs-prod-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-catalog-update_linux ./jobs/catalog_update/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-update-all-factrak-survey-deficits_linux ./jobs/update_all_factrak_survey_deficits/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-update-all-users-from-ldap_linux ./jobs/update_all_users_from_ldap/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-dorms-update_linux ./jobs/dorms_update/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-frosh-photos_linux ./jobs/frosh_photos/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-user-csv-data_linux ./jobs/user_csv_data/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-user-home-info_linux ./jobs/user_home_info/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-dining-update_linux ./jobs/dining_update/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-schedule-notifs_linux ./jobs/schedule_notifs/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-update-on-campus-semesters_linux ./jobs/update_on_campus_semesters/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-initialize-on-campus-semesters_linux ./jobs/update_on_campus_semesters/initial-calculation
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-update_profs_areas_of_study_linux ./jobs/update_profs_areas_of_study/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-ephmatch-reset_linux ./jobs/ephmatch_reset
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-ephmatch_update_dates_linux ./jobs/ephmatch_update_dates
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-library-hours-update_linux ./jobs/library_hours_update/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o job-mobile-fetcher_linux ./jobs/mobile_fetcher/cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o wso-jobctl_linux ./jobs/jobctl/cmd

.PHONY: sync-prod
sync-prod:
	bash .github/scripts/open-sync-prod.sh $(SHA)

.PHONY: check-prod-integrity
check-prod-integrity:
	bash .github/scripts/check-prod-integrity.sh

### Utility definitions
.PHONY: clean
clean:
	rm wso-backend >/dev/null 2>&1 || true
	rm job-* >/dev/null 2>&1 || true

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

.PHONY: commit
commit: jobs/dorms_update/cmd/data.go docs/docs.go fmt services/*/responses/*.go

.PHONY: run-dev
run-dev: dev

AIR_VERSION ?= v1.67.0

.PHONY: dev
dev:
	AIR_VERSION=$(AIR_VERSION) ./scripts/devenv.sh up

.PHONY: dev-down
dev-down:
	@./scripts/devenv.sh down

.PHONY: dev-reset
dev-reset:
	@./scripts/devenv.sh reset

# tests for local dev
TEST_COMPOSE = docker compose -p wso-test -f docker-compose.test.yml
TEST_MYSQL_ROOT_PASSWORD ?= secret-mysql-password
TEST_MYSQL_HOST ?= 127.0.0.1
TEST_MYSQL_PORT ?= 3307

.PHONY: _test-mysql-up
_test-mysql-up:
	@if [ -z "$(DOCKER)" ]; then echo "docker not found in PATH"; exit 1; fi
	@$(DOCKER) compose -p wso-test -f docker-compose.test.yml up -d --wait

.PHONY: _test-mysql-down
_test-mysql-down:
	@if [ -z "$(DOCKER)" ]; then echo "docker not found in PATH"; exit 1; fi
	@$(DOCKER) compose -p wso-test -f docker-compose.test.yml down
	@$(DOCKER) rm -f wso-mysql-test >/dev/null 2>&1 || true

# Relative paths so gotestsum --watch can fsnotify them (import paths are not dirs)
GO_MODULE = $(shell go list -m)
# Skip the same packages as CI (.github/coverage-exclude.txt)
ifdef IGNORE_SKIPS
TEST_PKGS = $(shell go list ./... | sed 's|^$(GO_MODULE)|.|')
else
HASH := \#
TEST_SKIP_REGEX = $(shell awk '/^[^$(HASH)[:space:]]/ { print $$1 }' .github/coverage-exclude.txt | paste -sd'|' -)
TEST_PKGS = $(shell go list ./... | grep -Ev '/($(TEST_SKIP_REGEX))(/|$$)' | sed 's|^$(GO_MODULE)|.|')
endif

export WSO_TEST_MYSQL_HOST ?= $(TEST_MYSQL_HOST)
export WSO_TEST_MYSQL_PORT ?= $(TEST_MYSQL_PORT)
export WSO_TEST_MYSQL_USER ?= root
export WSO_TEST_MYSQL_PASSWORD ?= $(TEST_MYSQL_ROOT_PASSWORD)

GOTESTSUM_VERSION ?= latest
GOTESTSUM_FMT = --format=pkgname-and-test-fails --format-hide-empty-pkg --hide-summary=output

ensure_gotestsum = \
	if ! command -v gotestsum >/dev/null 2>&1; then \
		go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION); \
	fi

run_tests = CI=1 gotestsum $(GOTESTSUM_FMT) -- $(1) $(TEST_PKGS)

.PHONY: itest
itest: _test-mysql-up
	@$(ensure_gotestsum)
	@echo "Press a to run all."
	@CI=1 gotestsum --watch $(GOTESTSUM_FMT) \
		--packages="$(TEST_PKGS)" \
		--raw-command -- ./scripts/gotest-filtered.sh

.PHONY: itest-race
itest-race: _test-mysql-up
	@$(ensure_gotestsum)
	@$(call run_tests,-race)

.PHONY: test
test:
	@echo "make test is deprecated; use \`make itest\` (watch) or \`make itest-race\`" >&2
	@exit 1

.PHONY: fast-test t
fast-test t:
	@echo "make $@ is deprecated; use \`make itest\`" >&2
	@exit 1

.PHONY: mod
mod:
	go mod tidy
	go mod download

ATLAS_MYSQL_ROOT_PASSWORD ?= secret-mysql-password
ATLAS_SCHEMA_DSN ?= root:$(ATLAS_MYSQL_ROOT_PASSWORD)@tcp(127.0.0.1:3306)/wso_atlas?parseTime=true&charset=utf8mb4&multiStatements=true
ATLAS_SCHEMA_URL ?= mysql://root:$(ATLAS_MYSQL_ROOT_PASSWORD)@127.0.0.1:3306/wso_atlas
ATLAS := $(shell command -v atlas 2>/dev/null)
DOCKER := $(shell command -v docker 2>/dev/null)

# throwaway MySQL for schema dump / diff (internal use)
.PHONY: atlas-mysql-up
atlas-mysql-up:
	@if [ -z "$(DOCKER)" ]; then echo "docker not found in PATH"; exit 1; fi
	@$(DOCKER) compose up -d --wait mysql
	@$(DOCKER) exec -e MYSQL_PWD=$(ATLAS_MYSQL_ROOT_PASSWORD) wso-mysql \
		mysql -uroot -e "CREATE DATABASE IF NOT EXISTS wso_atlas;"

# regenerate db/atlas/schema.sql from current GORM models (internal use)
.PHONY: atlas-schema
atlas-schema: atlas-mysql-up
	@if [ -z "$(ATLAS)" ]; then echo "atlas CLI not found; install from https://atlasgo.io/docs"; exit 1; fi
	@go run ./db/atlas/cmd/dump_schema -dsn '$(ATLAS_SCHEMA_DSN)'
	@$(ATLAS) schema inspect --url '$(ATLAS_SCHEMA_URL)' --format '{{ sql . "  " }}' -o db/atlas/schema.sql

# plan a new migration: make migrate name=<snake_case_name>
.PHONY: migrate
migrate:
	@if [ -z "$(ATLAS)" ]; then echo "atlas CLI not found; install from https://atlasgo.io/docs"; exit 1; fi
	@if [ -z "$(name)" ]; then echo "usage: make migrate name=<migration_name>"; exit 1; fi
	@$(MAKE) --no-print-directory atlas-schema
	@cd db/atlas && $(ATLAS) migrate diff "$(name)" --env local --config file://atlas.hcl

# checksum + replay on ephemeral MySQL
.PHONY: check-migrations
check-migrations:
	@if [ -z "$(ATLAS)" ]; then echo "atlas CLI not found; install from https://atlasgo.io/docs"; exit 1; fi
	@n=$$(ls db/atlas/migrations/*.up.sql 2>/dev/null | wc -l | tr -d ' '); \
	if [ "$$n" -eq 0 ]; then \
		echo "check-migrations: no migrations detected"; \
		exit 0; \
	fi; \
	cd db/atlas && $(ATLAS) migrate validate --env local --config file://atlas.hcl; \
	echo "check-migrations: passed ($$n version$$([ $$n -eq 1 ] || echo s))"

### Docker definitions
.PHONY: docker-builder
docker-builder:
	@if [ -z "$(DOCKER)" ]; then echo "docker not found in PATH"; exit 1; fi
	$(DOCKER) build -t $(DOCKER_TAG)/builder -f Dockerfile.builder .
	$(DOCKER) build -t $(DOCKER_TAG)/builder-cert -f Dockerfile.cert .

.PHONY: docker-dev
docker-dev: docker-builder
	$(DOCKER) build -t $(DOCKER_TAG):dev-latest -f Dockerfile.release .
	$(DOCKER) build -t $(DOCKER_TAG)-jobs:dev-latest -f Dockerfile.release_jobs .

.PHONY: docker-rel-dev
docker-rel-dev: docker-builder
	$(DOCKER) build -t $(DOCKER_TAG):dev-latest -f Dockerfile.release .

.PHONY: docker-jobs-dev
docker-jobs-dev: docker-builder
	$(DOCKER) build -t $(DOCKER_TAG)-jobs:dev-latest -f Dockerfile.release_jobs .
