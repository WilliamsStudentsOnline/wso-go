FROM golang:1.12.7

ADD . /go/src/github.com/WilliamsStudentsOnline/wso-go

WORKDIR /go/src/github.com/WilliamsStudentsOnline/wso-go

ENV GO111MODULE=on

# Need this for go caching
ENV XDG_CACHE_HOME=/tmp/.cache
RUN mkdir /tmp/.cache
RUN chmod -R 777 /tmp/.cache

RUN go mod download
RUN go mod verify

RUN GOOS=linux go build -ldflags "-w -s" -tags=jsoniter -o /go/bin/wso-go

CMD go test -race ./...
