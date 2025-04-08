package consumer

import (
	"errors"
	"fmt"
	model "github.com/cristiandpt/healthcare/measures-consumer/internal/model"
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

const (
	reconnectDelay = 5 * time.Second
	reInitDelay    = 2 * time.Second
)

var (
	errNotConnected = errors.New("not connected to a server")
)

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
		case model.ConsumeMessage:
			actor.handleConsume()
		case model.CloseMessage:
			actor.handleClose()
		case model.ProcessMessage:
			actor.processDelivery(m.Delivery)
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

// connect will create a new AMQP connection.
func (actor *ConsumerActor) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}
	actor.changeConnection(conn)
	actor.isReady = true
	actor.logger.Println("Connected to RabbitMQ!")
	return conn, nil
}

// handleReInit will wait for a channel error and continuously attempt to re-initialize the channel.
func (actor *ConsumerActor) handleReInit(conn *amqp.Connection) bool {
	for {
		if err := actor.init(conn); err != nil {
			actor.logger.Printf("Failed to initialize channel: %s. Retrying in %s...\n", err, reInitDelay)
			select {
			case <-time.After(reInitDelay):
			case <-actor.notifyConnClose:
				actor.logger.Println("Connection closed. Reconnecting...")
				return false
			case <-actor.mailbox: // Allow exiting if the actor is closed during re-init
				return true
			}
			continue
		}

		select {
		case <-actor.mailbox: // Allow exiting if the actor is closed
			return true
		case <-actor.notifyConnClose:
			actor.logger.Println("Connection closed. Reconnecting...")
			return false
		case <-actor.notifyChanClose:
			actor.logger.Println("Channel closed. Re-initializing...")
		}
	}
}

// init will initialize the channel and declare the queue.
func (actor *ConsumerActor) init(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		actor.queueName, // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return err
	}

	if err := ch.Qos(
		1,     // prefetchCount: process one message at a time
		0,     // prefetchSize: no limit
		false, // global: apply per consumer
	); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	actor.changeChannel(ch)
	actor.logger.Println("Channel initialized and queue declared.")
	return nil
}

// changeConnection takes a new connection and updates the close listener.
func (actor *ConsumerActor) changeConnection(connection *amqp.Connection) {
	actor.conn = connection
	actor.notifyConnClose = make(chan *amqp.Error, 1)
	actor.conn.NotifyClose(actor.notifyConnClose)
}

// changeChannel takes a new channel and updates the channel listeners.
func (actor *ConsumerActor) changeChannel(channel *amqp.Channel) {
	actor.channel = channel
	actor.notifyChanClose = make(chan *amqp.Error, 1)
	actor.channel.NotifyClose(actor.notifyChanClose)
}

// handleConsume starts consuming messages from the queue.
func (actor *ConsumerActor) handleConsume() {
	if !actor.isReady {
		actor.logger.Println("Not connected, cannot start consuming.")
		return
	}

	deliveries, err := actor.channel.Consume(
		actor.queueName,   // queue
		actor.consumerTag, // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		actor.logger.Printf("Failed to start consuming: %s\n", err)
		return
	}
	actor.consumeChan = deliveries
	actor.logger.Printf("Started consuming with tag: %s\n", actor.consumerTag)

	// Start a new goroutine to continuously process deliveries
	go actor.processDeliveries()
}

// processDeliveries continuously reads deliveries from the consume channel and sends them
// as ProcessMessage to the actor's mailbox for processing.
func (actor *ConsumerActor) processDeliveries() {
	for d := range actor.consumeChan {
		actor.mailbox <- model.ProcessMessage{Delivery: model.Delivery(d)}
	}
	actor.logger.Println("Consumption stopped.")
}

// processDelivery handles the processing of a single received message.
func (actor *ConsumerActor) processDelivery(d model.Delivery) {
	actor.logger.Printf("Received message [%v]: %q\n", d.DeliveryTag, d.Body)
	// Saving to MomgoDB
	// Acknowledge the message to remove it from the queue
	if err := d.Ack(false); err != nil {
		actor.logger.Printf("Error acknowledging message [%v]: %s\n", d.DeliveryTag, err)
		// Optionally, you can nack the message and requeue or discard it
		// d.Nack(false, true) // requeue
		// d.Nack(false, false) // discard
	} else {
		actor.logger.Printf("Message [%v] acknowledged\n", d.DeliveryTag)
	}
}

// Consume sends a message to the actor's mailbox to start consuming.
func (actor *ConsumerActor) Consume() {
	actor.mailbox <- model.ConsumeMessage{}
}

// Close signals the actor to shut down.
func (actor *ConsumerActor) Close() {
	actor.mailbox <- model.CloseMessage{}
	actor.wg.Wait() // Wait for the actor to finish
	actor.logger.Println("Consumer actor stopped.")
}

// handleClose performs the closing of the channel and connection.
func (actor *ConsumerActor) handleClose() {
	actor.logger.Println("Closing connection...")
	if actor.channel != nil {
		if err := actor.channel.Cancel(actor.consumerTag, false); err != nil {
			actor.logger.Printf("Error cancelling consumer [%s]: %s\n", actor.consumerTag, err)
		}
		if err := actor.channel.Close(); err != nil {
			actor.logger.Printf("Error closing channel: %s\n", err)
		}
	}
	if actor.conn != nil {
		if err := actor.conn.Close(); err != nil {
			actor.logger.Printf("Error closing connection: %s\n", err)
		}
	}
	actor.isReady = false
	close(actor.mailbox) // Close the mailbox to signal the run loop to exit
}
