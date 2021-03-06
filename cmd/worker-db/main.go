package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ezeql/appcues-increment-simple/internal/storage"

	"github.com/go-redis/redis/v7"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	// Redis host environment var name
	redisHostEnvName = "REDIS"
	dbHostEnvName    = "DATABASE_URL"

	batchSize     = 100
	prefetchLimit = 500

	pollDuration = 100 * time.Millisecond
	batchTimeout = time.Second
)

func main() {
	log.SetOutput(os.Stdout)
	log.Printf("worker-db started")

	// check for redis env var
	redisHost, found := os.LookupEnv(redisHostEnvName)
	if !found {
		log.Fatalf("required environment var not defined: %v\n", redisHostEnvName)
	}

	// check for db env var
	dbHost, found := os.LookupEnv(dbHostEnvName)
	if !found {
		log.Fatalf("required environment var not defined: %v\n", dbHost)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisHost})

	pool, err := pgxpool.Connect(context.Background(), dbHost)
	if err != nil {
		log.Fatalf("failed to open a connection to pgsql: %v\n", err)
	}
	defer pool.Close()

	worker, err := storage.NewStoreWorker(redisClient, pool, &storage.StoreConfig{
		PrefetchLimit: prefetchLimit,
		PollDuration:  pollDuration,
	})

	if err != nil {
		log.Fatalf("couldn't start  worker: %v\n", err)
	}

	defer worker.StopAll()

	// TODO: Missing handle graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT)
	defer signal.Stop(signals)

	<-signals // wait for signal
	go func() {
		<-signals // hard exit on second signal (in case shutdown gets stuck)
		os.Exit(1)
	}()
}
