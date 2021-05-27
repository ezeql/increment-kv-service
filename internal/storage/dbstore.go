package storage

import (
	"log"
	"time"

	"github.com/ezeql/appcues-increment-simple/internal/incrementsrv"
	"github.com/go-redis/redis/v7"

	"github.com/jackc/pgx/v4/pgxpool"

	cmap "github.com/orcaman/concurrent-map"
)

const (
	TickerInterval = time.Second * 9

	sqlQuery = `INSERT INTO increments (key, value)
	SELECT keys.key, values.value
	FROM UNNEST($1::UUID[]) WITH ORDINALITY AS keys(key, idx)
	INNER JOIN UNNEST($2::BIGINT[]) WITH ORDINALITY AS values(value, idx) ON keys.idx = values.idx
	ON CONFLICT (key) DO UPDATE SET value = excluded.value + increments.value`

	batchSize    = 100
	batchTimeout = time.Second
)

type StoreWorker struct {
	db       *pgxpool.Pool
	Redis    *redis.Client
	consumer *incrementsrv.Consumer
}

type StoreConfig struct {
	PrefetchLimit int64
	PollDuration  time.Duration
}

func NewStoreWorker(redisClient *redis.Client, db *pgxpool.Pool, cfg *StoreConfig) (*StoreWorker, error) {
	srv, err := incrementsrv.NewConsumer(redisClient, cfg.PrefetchLimit, cfg.PollDuration)
	if err != nil {
		return nil, err
	}
	worker := &StoreWorker{db: db,
		Redis:    redisClient,
		consumer: srv,
	}

	worker.startConsumer(batchSize, batchTimeout)

	return worker, nil
}

func (w *StoreWorker) startConsumer(batchSize int64, batchTimeout time.Duration) error {
	ch := make(chan *incrementsrv.Input)
	kvMap := cmap.New()

	consumer := &consumer{db: w.db, redis: w.Redis, kvMap: &kvMap, ch: ch}

	log.Println("CONFIG: flushing to database every", TickerInterval.Milliseconds(), "ms ⏲️")

	//launch goroutine in charge
	go consumer.consume()

	return w.consumer.AddBatchConsumer(batchSize, batchTimeout, consumer)
}

func (w *StoreWorker) StopAll() {
	w.consumer.StopAllConsuming()
}
