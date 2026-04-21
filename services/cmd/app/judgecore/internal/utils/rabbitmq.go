package utils

import (
	"FeasOJ/app/judgecore/internal/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ConnectRabbitMQ RabbitMQ连接
func ConnectRabbitMQ(rmqConfig config.RabbitMQ) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(rmqConfig.Host)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	prefetch := rmqConfig.Prefetch
	if prefetch <= 0 {
		prefetch = 1
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, err
	}

	exchange := rmqConfig.Exchange
	if exchange == "" {
		exchange = "judge.submission.exchange"
	}
	mainQueue := rmqConfig.MainQueue
	if mainQueue == "" {
		mainQueue = "judge.submission.main"
	}

	if err := ch.ExchangeDeclare(exchange, amqp.ExchangeDirect, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, err
	}
	if _, err = ch.QueueDeclare(
		mainQueue,
		true,
		false,
		false,
		false,
		amqp.Table{"x-dead-letter-exchange": exchange, "x-dead-letter-routing-key": "judge.submission.dead"},
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, err
	}
	if err := ch.QueueBind(mainQueue, "judge.submission.enqueue", exchange, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}
