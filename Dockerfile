FROM golang:1.10.3-alpine3.8

COPY main.go /src/main.go

RUN apk add git gcc libc-dev

WORKDIR /src/

RUN go get github.com/eclipse/paho.mqtt.golang && \
    go get github.com/iZonex/go-dht

RUN go build main.go