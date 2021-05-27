package incrementsrv

import (
	"encoding/json"
	"time"

	"github.com/adjust/rmq/v3"
	"github.com/go-redis/redis/v7"
)

// Redis connection name prefix
const redisPrefix = "appcues-increment-srv"

// queue name for publishing increments
const queueName = "increments"

type Publisher struct {
	Queue rmq.Queue
	conn  rmq.Connection
}

type Consumer struct {
	Publisher *Publisher
}

func NewPublisher(redisClient *redis.Client) (*Publisher, error) {
	//Ignoring network erorrs layer handling
	connection, err := rmq.OpenConnectionWithRedisClient(redisPrefix, redisClient, nil)
	if err != nil {
		return nil, err
	}

	queue, err := connection.OpenQueue(queueName)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		Queue: queue,
		conn:  connection,
	}, nil
}

func (inc *Publisher) Publish(item Input) error {
	if err := item.Valid(); err != nil {
		return err
	}

	b, err := json.Marshal(item)
	if err != nil {
		return err
	}

	if err := inc.Queue.Publish(string(b)); err != nil {
		return err
	}

	return nil
}

func NewConsumer(redisClient *redis.Client, prefetchLimit int64, pollDuration time.Duration) (*Consumer, error) {
	srv, err := NewPublisher(redisClient)
	if err != nil {
		return nil, err
	}

	if err := srv.Queue.StartConsuming(prefetchLimit, pollDuration); err != nil {
		return nil, err
	}
	return &Consumer{Publisher: srv}, nil

}

func (inc *Consumer) AddBatchConsumer(batchSize int64, batchTimeout time.Duration, consumer rmq.BatchConsumer) error {
	_, err := inc.Publisher.Queue.AddBatchConsumer(redisPrefix, batchSize, batchTimeout, consumer)
	return err
}

func (inc *Consumer) StopAllConsuming() <-chan struct{} {
	return inc.Publisher.conn.StopAllConsuming()
}
