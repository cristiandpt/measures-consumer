package consumer

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"sync"
	"time"
)

// ConsumerActor represents the actor responsible for consuming messages from RabbitMQ.
type ConsumerActor struct {
	queueName       string
	addr            string
	conn            *amqp.Connection
	channel         *amqp.Channel
	notifyConnClose chan *amqp.Error
	notifyChanClose chan *amqp.Error
	isReady         bool
	mailbox         chan interface{} // Actor's mailbox for messages
	wg              sync.WaitGroup
	logger          *log.Logger
	consumerTag     string // Identifier for the consumer
	consumeChan     <-chan amqp.Delivery
}

// NewConsumerActor creates a new RabbitMQ consumer actor.
func NewConsumerActor(queueName, addr string) *ConsumerActor {
	actor := &ConsumerActor{
		queueName:   queueName,
		addr:        addr,
		mailbox:     make(chan interface{}),
		logger:      log.New(os.Stdout, "[ConsumerActor] ", log.LstdFlags),
		consumerTag: fmt.Sprintf("consumer-%d", time.Now().UnixNano()), // Unique consumer tag
	}
	actor.wg.Add(1)
	go actor.run() // Start the actor's processing loop
	return actor
}

func (actor *ConsumerActor) run() {
	defer actor.wg.Done()
	actor.handleReconnect()

	for msg := range actor.mailbox {
		switch m := msg.(type) {
		case ConsumeMessage:
			// TDDO
		case CloseMessage:
			// TDDO
		case ProcessMessage:
			// TDDO
		default:
			actor.logger.Printf("Received unknown message type: %T\n", msg)
		}
	}
		actor.logger.Println("Consumer actor mailbox closed.")

}
