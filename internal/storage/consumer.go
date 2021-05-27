package storage

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/adjust/rmq/v3"
	"github.com/ezeql/appcues-increment-simple/internal/incrementsrv"
	"github.com/go-redis/redis/v7"
	"github.com/jackc/pgx/v4/pgxpool"

	cmap "github.com/orcaman/concurrent-map"
)

type consumer struct {
	db    *pgxpool.Pool
	redis *redis.Client
	rmq.BatchConsumer
	kvMap *cmap.ConcurrentMap
	ch    chan *incrementsrv.Input
}

func (consumer *consumer) consume() {
	ticker := time.NewTicker(TickerInterval)
	for {
		select {
		case <-ticker.C:
			items := consumer.kvMap.Items()
			r, err := consumer.sendAndClear()
			if err != nil {
				log.Println(err)
				continue
			}
			if r > 0 {
				log.Println("flush: keys sent to db: ", items)
			}

		case c := <-consumer.ch:
			consumer.kvMap.Upsert(c.Key.String(), c.Value, func(exist bool, valueInMap, newValue interface{}) interface{} {
				if !exist {
					return newValue.(uint64)
				}
				return valueInMap.(uint64) + newValue.(uint64)
			})
		}
	}
}

func (consumer *consumer) Consume(batch rmq.Deliveries) {
	for _, delivery := range batch {
		reader := strings.NewReader(delivery.Payload())

		//check for toxic messages
		incIn, err := incrementsrv.InputFromJSONReader(reader)
		if err != nil {
			delivery.Reject()
		}

		//send to channel
		consumer.ch <- incIn

		//ack message. Implementation handles retry
		if err := delivery.Ack(); err != nil {
			// delivery.Push()
			continue
		}
	}

}

func (consumer *consumer) sendAndClear() (int64, error) {
	var (
		keys   []interface{}
		values []interface{}
	)

	consumer.kvMap.IterCb(func(key string, v interface{}) {
		keys = append(keys, key)
		values = append(values, v)
	})

	res, err := consumer.db.Exec(context.Background(), sqlQuery, keys, values)

	if err != nil {
		return 0, err
	}

	// if all went good, clear map.
	// Otherwise, next run will retry current map plus the new ones until the next.
	consumer.kvMap.Clear()

	return int64(res.RowsAffected()), nil

}
