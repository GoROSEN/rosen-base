package rpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/google/martian/log"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpRpcClient struct {
	conn           *amqp.Connection
	pubCh          *amqp.Channel
	replyQueueName string
	callbacks      map[string]func([]byte)
}

func (c *AmqpRpcClient) Start() error {
	config := config.GetConfig()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", config.Rpc.Amqp.User, config.Rpc.Amqp.Password, config.Rpc.Amqp.Host, config.Rpc.Amqp.Port))
	if err != nil {
		log.Errorf("Failed to connect to RabbitMQ: %v", err)
		return err
	}
	c.conn = conn
	c.callbacks = make(map[string]func([]byte))

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Failed to open a channel: %v", err)
		return err
	}
	c.pubCh = ch

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		log.Errorf("Failed to declare a queue: %v", err)
		return err
	}

	c.replyQueueName = q.Name
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Errorf("Failed to register a consumer: %v", err)
		return err
	}

	go func() {
		for d := range msgs {
			fn, exists := c.callbacks[d.CorrelationId]
			if exists {
				fn(d.Body)
			} else {
				log.Errorf("callback function for msg id %v does not exist", d.CorrelationId)
			}
		}
	}()

	return nil
}

func (c *AmqpRpcClient) Stop() {
	c.pubCh.Close()
	c.conn.Close()
}

func (c *AmqpRpcClient) Call(qname string, command string, params map[string]interface{}, cb func([]byte)) error {
	var body RpcCallData
	body.Command = command
	if params == nil {
		params = map[string]interface{}{}
	}
	body.Params = params
	bodyData, _ := json.Marshal(&body)
	return c.call(qname, bodyData, cb)
}

func (c *AmqpRpcClient) call(qname string, body []byte, cb func([]byte)) error {

	if c.pubCh.IsClosed() {
		err := fmt.Errorf("rpc publish channel is closed")
		log.Infof("AmqpRpcClient call error: %v", err)
		return err
	}

	corrId := uuid.New().String()
	c.callbacks[corrId] = cb

	if err := c.pubCh.PublishWithContext(
		context.Background(),
		"",    // exchange
		qname, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrId,
			ReplyTo:       c.replyQueueName,
			Body:          body,
		}); err != nil {
		log.Errorf("Failed to publish a message: %v", err)
		return err
	}
	return nil
}
