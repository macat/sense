package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/streadway/amqp"
)

var ic influx.Client

func main() {
	log.Println("INFLUX_HOST", os.Getenv("INFLUX_HOST"))
	log.Println("AMPQ_HOST", os.Getenv("AMPQ_HOST"))

	var err error
	config := influx.HTTPConfig{
		Addr:     os.Getenv("INFLUX_HOST"),
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PASS"),
	}
	ic, err = influx.NewHTTPClient(config)
	if err != nil {
		log.Fatal("Failed to set up InfluxDB.", err)
	}
	defer ic.Close()

	conn, err := amqp.Dial(os.Getenv("AMPQ_HOST"))
	if err != nil {
		log.Fatal("Failed to connect to AMPQ server.", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open channel.")
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("mqtt_metrics", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Queue declaration failed.")
	}

	err = ch.QueueBind(q.Name, "sense.#", "amq.topic", false, nil)
	if err != nil {
		log.Fatal("Queue binding failed.")
	}

	msgs, err := ch.Consume("mqtt_metrics", "collector", true, true, false, false, nil)

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			keys := strings.SplitN(d.RoutingKey, ".", 6)
			if len(keys) == 6 {
				data := strings.SplitN(string(d.Body), ",", 2)
				if len(data) == 2 {
					go write(keys[1], keys[2], keys[3], keys[4], keys[5], data[0], data[1])
				}
			}
		}
	}()
	log.Printf("Collector running.")
	<-forever
}

func write(location, floor, room, deviceID, sensor, timestamp, data string) {
	var err error
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		log.Println("Timestamp conversion failed", err)
		return
	}
	t := time.Unix(i, 0)

	var p *influx.Point
	d, err := strconv.Atoi(data)
	tags := map[string]string{
		"location":  location,
		"floor":     floor,
		"room":      room,
		"device_id": deviceID,
		"sensor":    sensor,
	}
	if err != nil {
		log.Println("Data conversion failed, using string.", err)
		p, err = influx.NewPoint(
			"sense",
			tags,
			map[string]interface{}{
				"s": data,
				"d": 0,
			},
			t,
		)
	} else {
		p, err = influx.NewPoint(
			"sense",
			tags,
			map[string]interface{}{
				"s": "",
				"d": d,
			},
			t,
		)
	}

	if err != nil {
		log.Println("Failed creating influx point", err)
		return
	}

	// Create a new point batch
	bp, _ := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  "sense",
		Precision: "s",
	})

	bp.AddPoint(p)

	if err := ic.Write(bp); err != nil {
		log.Println("Failed writing influx point", err)
		return
	}

}
