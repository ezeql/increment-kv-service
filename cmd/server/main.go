package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ezeql/appcues-increment-simple/internal/hits"
	"github.com/go-redis/redis/v7"
)

const (
	// Redis host env var name
	redisHostEnvName = "REDIS"
	// service listen address
	listenAddress = ":3333"
)

func main() {
	log.SetOutput(os.Stdout)

	log.Printf("Increments HTTP server started")

	// check for redis env var
	redisHost, found := os.LookupEnv(redisHostEnvName)
	if !found {
		log.Fatalf("required environment var not defined: %v\n", redisHostEnvName)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisHost})
	cfg := hits.Config{
		RedisClient: redisClient,
	}

	hits, err := hits.HitsHTTP(cfg)
	if err != nil {
		log.Fatalf("couldn't create http server: %v\n", err)
	}

	// launch server with sane defaults
	go func() {
		log.Printf("service running at %s", listenAddress)
		s := http.Server{
			Addr:    listenAddress,
			Handler: hits.Router,
		}

		if err := s.ListenAndServe(); err != nil {
			log.Panicf("error while serving service: %s", err)
		}
	}()

	// TODO: Missing handle graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("Stopping API server.")

}
