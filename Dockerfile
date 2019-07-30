FROM golang:1.12.7

ADD . /go/src/github.com/WilliamsStudentsOnline/wso-go

WORKDIR /go/src/github.com/WilliamsStudentsOnline/wso-go

ENV GO111MODULE=on

RUN go mod download
RUN go mod verify

RUN GOOS=linux go build -ldflags "-w -s" -tags=jsoniter -o /go/bin/wso-go

CMD go test -race ./...
