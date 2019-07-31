FROM golang:1.12.7

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

# Copy the rest of the project into the file
COPY . .

# Build the go file
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -tags=jsoniter -o /go/bin/wso-go

# Default entrypoint. TODO: move this binary to a scratch deployment for minimal size (issue with cgo?)
ENTRYPOINT ["/go/bin/wso-go"]
