FROM golang:1.12.8

# Turn on modules
ENV GO111MODULE=on

# Create our workspace
RUN mkdir -p /wso/jobs
RUN mkdir -p /go/src/github.com/WilliamsStudentsOnline/wso-go
WORKDIR /go/src/github.com/WilliamsStudentsOnline/wso-go

# Copy timezone into wso directory
RUN cp /usr/local/go/lib/time/zoneinfo.zip /wso/zoneinfo.zip

# Copy the go module files, so this can be cached by docker
COPY go.mod .
COPY go.sum .

# Download the go mod files (cached by docker)
RUN go mod download
RUN go mod verify

# Get documentation maker
RUN go get -u github.com/swaggo/swag/cmd/swag

# Copy the rest of the project into the file
COPY . .

# Run generation
RUN go generate github.com/WilliamsStudentsOnline/wso-go/...

# Generate API documentation
RUN swag init -g server/router.go

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

### Build the main server ###
RUN go build -ldflags "-w -s" -tags=jsoniter \
    -o /wso/wso-backend /go/src/github.com/WilliamsStudentsOnline/wso-go/server/cmd

### Build the jobs ###
# Catalog Update
RUN go build -ldflags "-w -s" -tags=jsoniter \
    -o /wso/jobs/catalog-update /go/src/github.com/WilliamsStudentsOnline/wso-go/jobs/catalog_update/cmd
# Update all factrak survey deficits
RUN go build -ldflags "-w -s" -tags=jsoniter \
    -o /wso/jobs/update-all-factrak-survey-deficits \
    /go/src/github.com/WilliamsStudentsOnline/wso-go/jobs/update_all_factrak_survey_deficits/cmd
# Update all users from LDAP
RUN go build -ldflags "-w -s" -tags=jsoniter \
    -o /wso/jobs/update-all-users-from-ldap \
    /go/src/github.com/WilliamsStudentsOnline/wso-go/jobs/update_all_users_from_ldap/cmd
# Dorms Update
RUN go build -ldflags "-w -s" -tags=jsoniter \
    -o /wso/jobs/dorms-update \
    /go/src/github.com/WilliamsStudentsOnline/wso-go/jobs/dorms_update/cmd