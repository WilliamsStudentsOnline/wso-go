FROM golang:1.12.7 AS builder

# Turn on modules
ENV GO111MODULE=on

# Create our workspace
RUN mkdir -p /go/src/github.com/WilliamsStudentsOnline/wso-go
WORKDIR /go/src/github.com/WilliamsStudentsOnline/wso-go

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

# Generate API documentation
RUN swag init -g server/router.go

# Build the go file
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o /go/bin/wso-backend /go/src/github.com/WilliamsStudentsOnline/wso-go/server/cmd

# Default entrypoint. TODO: move this binary to a scratch deployment for minimal size (issue with cgo?)
#ENTRYPOINT ["/go/bin/wso-backend"]
