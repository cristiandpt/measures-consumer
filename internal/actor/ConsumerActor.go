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


// handleReconnect will wait for a connection error and continuously attempt to reconnect.
func (actor *ConsumerActor) handleReconnect() {
	for {
		actor.isReady = false
		actor.logger.Println("Attempting to connect...")

		conn, err := actor.connect(actor.addr)
		if err != nil {
			actor.logger.Printf("Failed to connect: %s. Retrying in %s...\n", err, reconnectDelay)
			select {
			case <-time.After(reconnectDelay):
			case <-actor.mailbox: // Allow exiting if the actor is closed during reconnect
				return
			}
			continue
		}

		if actor.handleReInit(conn) {
			return // Exit if re-initialization was part of a shutdown
		}
	}
}
