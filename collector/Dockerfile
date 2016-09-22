FROM golang:1.7

RUN mkdir -p /go/src/github.com/namely/sense/collector
WORKDIR /go/src/github.com/namely/sense/collector

RUN go get github.com/streadway/amqp

CMD go run main.go

