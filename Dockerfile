FROM golang:alpine

RUN apk add --no-cache git

ADD . /go/src/github.com/go-park-mail-ru/2018_2_LSP_GAME

RUN cd /go/src/github.com/go-park-mail-ru/2018_2_LSP_GAME && go get ./...

RUN go install github.com/go-park-mail-ru/2018_2_LSP_GAME

ENTRYPOINT /go/bin/2018_2_LSP_GAME

EXPOSE 8080