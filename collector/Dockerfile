FROM golang:1.7
RUN go get github.com/streadway/amqp
RUN go get github.com/influxdata/influxdb/client/v2

RUN mkdir -p /go/src/github.com/macat/sense/collector
WORKDIR /go/src/github.com/macat/sense/collector
ADD . /go/src/github.com/macat/sense/collector

CMD go run main.go
