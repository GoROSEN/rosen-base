package rpc

import (
	"context"
	"fmt"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/google/martian/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpRpcServer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (s *AmqpRpcServer) Start(handle func([]byte) (string, error)) error {
	config := config.GetConfig()
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", config.Rpc.Amqp.User, config.Rpc.Amqp.Password, config.Rpc.Amqp.Host, config.Rpc.Amqp.Port))
	if err != nil {
		log.Errorf("cannot connect to amqp server: %v", err)
		return err
	}
	s.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Failed to open a channel: %v", err)
		return err
	}
	s.ch = ch

	listenQueue := config.Rpc.Queues["listen"]
	log.Infof("listenning queue: %v", listenQueue)
	q, err := ch.QueueDeclare(
		listenQueue, // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Errorf("Failed to declare a queue: %v", err)
		return err
	}

	if err := ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		log.Errorf("Failed to set QoS: %v", err)
		return err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
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
			// 执行功能
			response, _ := handle(d.Body)

			// 发送结果
			err = ch.PublishWithContext(
				context.Background(),
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				})
			if err != nil {
				log.Errorf("Failed to publish a message: %v", err)
				return
			}

			d.Ack(false)
		}
	}()

	return nil
}

func (s *AmqpRpcServer) Stop() {
	s.ch.Close()
	s.conn.Close()
}
