FROM golang:alpine as builder

LABEL maintainer="Gleb Lys <barabashka.zzz@gmail.com>"

RUN apk add git

COPY . /go/src/github.com/hardenchant/socks5-list-proxy

WORKDIR /go/src/github.com/hardenchant/socks5-list-proxy

RUN go get && go build

RUN adduser -D socks5-list-proxy

USER socks5-list-proxy

ENTRYPOINT ["socks5-list-proxy"]