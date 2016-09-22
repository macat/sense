package main

import (
	"log"
	"os"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial(os.Getenv("AMPQ_SERVER"))
	if err != nil {
		log.Fatal("Failed to connect to AMPQ server.")
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
			log.Println(d.RoutingKey)
			log.Println(string(d.body))
		}
	}()
	log.Printf("Collector running.")
	<-forever
}
