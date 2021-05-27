package handlers_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/jackc/pgx/v4/pgxpool"
	cmap "github.com/orcaman/concurrent-map"

	"github.com/ezeql/appcues-increment-simple/internal/hits"
	"github.com/ezeql/appcues-increment-simple/internal/storage"
	"github.com/ezeql/appcues-increment-simple/testdata"
	"github.com/go-redis/redis/v7"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// shared resources
var (
	c  *redis.Client
	db *pgxpool.Pool
)

const (
	// redisHostEnvName = "REDIS"

	batchSize     = 100
	prefetchLimit = 500

	pollDuration = 100 * time.Millisecond
	batchTimeout = time.Second // approx since we want 10 seconds max

)

// Init Deps
func TestMain(m *testing.M) {
	// Redis
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	c = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	//Postgres

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "13.2-alpine", []string{
		"POSTGRES_DB=postgres",
		"POSTGRES_USER=postgres",
		"POSTGRES_PASSWORD=postgres",
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {

		var err error
		db, err = pgxpool.Connect(
			context.Background(),
			fmt.Sprintf(
				"postgres://postgres:postgres@localhost:%s/postgres?sslmode=disable", resource.GetPort("5432/tcp")))

		if err != nil {
			return err
		}
		return db.Ping(context.Background())

	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	db.Exec(context.Background(), `CREATE TABLE increments(
		"key" UUID NOT NULL,
		"value" BIGINT NOT NULL,
		CONSTRAINT increments_pk PRIMARY KEY ("key"));`)

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

type test struct {
	json       string
	statusCode int
}

func TestHitsServer_CreateIncrement(t *testing.T) {
	runTest := func(t *testing.T, test test) {
		cfg := hits.Config{
			RedisClient: c,
		}

		h, err := hits.HitsHTTP(cfg)
		assert.Nil(t, err)

		ts := httptest.NewServer(http.HandlerFunc(h.Router.ServeHTTP))
		defer ts.Close()

		res, err := http.Post(ts.URL+"/increment", "application/json", strings.NewReader(test.json))

		assert.Nil(t, err)
		assert.Equal(t, test.statusCode, res.StatusCode, "they should be equal. JSON %s", test.json)

	}

	tests := []test{
		{json: testdata.ValidIncJSONReq, statusCode: http.StatusOK},
		{json: testdata.InvalidValueJSONReq, statusCode: http.StatusBadRequest},
		{json: testdata.InvalidKeyJSONReq, statusCode: http.StatusBadRequest},
		{json: testdata.MissingValueJSONReq, statusCode: http.StatusBadRequest},
		{json: testdata.MissingKeyJSONReq, statusCode: http.StatusBadRequest},
		{json: testdata.InvalidJSONReq, statusCode: http.StatusBadRequest},
	}

	for _, test := range tests {
		runTest(t, test)
	}
}

func TestHitsServer_Concurrent(t *testing.T) {
	// Clear table ( previous test adds a record )
	_, err := db.Exec(context.Background(), "TRUNCATE TABLE increments")
	assert.Nil(t, err)

	cfg := hits.Config{
		RedisClient: c,
	}

	const (
		TotalPerSecond = 1000            // requests per second
		totalDuration  = 5 * time.Second // total length of the test
	)

	var totalIterations = len(testdata.ListValidKeys) // ammount of concurrent clients sending requests

	h, err := hits.HitsHTTP(cfg)
	assert.Nil(t, err)

	worker, err := storage.NewStoreWorker(c, db, &storage.StoreConfig{
		PrefetchLimit: prefetchLimit,
		PollDuration:  pollDuration,
	})
	assert.Nil(t, err)

	defer worker.StopAll()

	ts := httptest.NewServer(http.HandlerFunc(h.Router.ServeHTTP))
	defer ts.Close()

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(totalIterations)
	m := cmap.New()

	runTestwith := func(k string) {
		rate := vegeta.Rate{Freq: TotalPerSecond, Per: time.Second}

		duration := totalDuration

		targeter := vegeta.NewStaticTargeter(vegeta.Target{
			Method: "POST",
			URL:    ts.URL + "/increment",
			Body:   []byte(testdata.BuildIncrementJSON(k, 1)),
		})
		attacker := vegeta.NewAttacker()

		// attacker
		var metrics vegeta.Metrics
		for res := range attacker.Attack(targeter, rate, duration, k) {
			metrics.Add(res)

			m.Upsert(k, 1, func(exist bool, valueInMap, newValue interface{}) interface{} {
				if !exist {
					return newValue
				}
				return valueInMap.(int) + newValue.(int)
			})
		}

		metrics.Close()
		waitGroup.Done()
	}

	// run load testers in parallel
	for i := 0; i < totalIterations; i++ {
		go runTestwith(testdata.ListValidKeys[i])
	}

	waitGroup.Wait()                   // Make sure all requests are completed
	time.Sleep(storage.TickerInterval) // Wait for the ticker to run once

	var actualValuesSum int

	expectedValuePerKey := TotalPerSecond * int(totalDuration.Seconds())        // value for every key
	expectedValuesTotalSum := expectedValuePerKey * len(testdata.ListValidKeys) // sum of all values

	err = db.QueryRow(
		context.Background(),
		"SELECT SUM(value) FROM increments WHERE value=$1", expectedValuePerKey).Scan(&actualValuesSum)

	assert.Nil(t, err)

	assert.Equal(t, expectedValuesTotalSum, actualValuesSum, "There must be %v unique keys with %v hits each totalling %v", len(testdata.ListValidKeys), expectedValuePerKey, expectedValuesTotalSum)

}
