package main

import (
	actor "github.com/cristiandpt/healthcare/measures-consumer/internal/actor"
	"log"
	"os"
)

func main() {

	queueMeasures := "measures_queue"
	addr := os.Getenv("RABBITMQ_ADDR")
	rabbitActorConsumer := actor.NewConsumerActor(queueMeasures, addr)
	defer rabbitActorConsumer.Close()

	// Start consuming messages
	rabbitActorConsumer.Consume()

	// Keep the main goroutine alive to allow the consumer actor to run
	forever := make(chan bool)
	log.Println(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
