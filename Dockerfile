FROM golang:alpine

RUN apk add --no-cache git

ADD app.linux.amd64 /go/bin/app

ENTRYPOINT /go/bin/app

EXPOSE 8080