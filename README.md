![Appcues Image](./images/appcues.png "Appcues Image")

# Platform Engineer Project

## General Description

The flow starts with a client action sending a POST request with a JSON encoded body.

The first layer to take action in the system is [Caddy](https://caddyserver.com/), a reverse proxy. It will distribute the requests from the clients between all available [HTTP servers](https://github.com/ezeql/kv-service/blob/master/cmd/server/main.go).

Rate limiting runs at this step, before passing the data downstream. It is implemented [with a Caddy extension that returns a HTTP 429](https://github.com/ezeql/kv-service/blob/master/caddy/Caddyfile#L11) if the amount of requests is larger than a threshold value.

[HTTP server](https://github.com/ezeql/kv-service/blob/master/cmd/server/main.go) instances that will validate the JSON payload and, if valid, will push a message to a messaging queue implemented in top of Redis using a library called [RMQ](https://github.com/adjust/rmq). It uses the publish-subscribe pattern.

If by this moment the client submitted a valid request, it will get a HTTP 200.
Otherwise a HTTP 400 is returned.

A predefined [set of workers](https://github.com/ezeql/kv-service/blob/master/cmd/worker-db/main.go) are subscribed to this queue.
These will pickup and distribute batches of up to N items from it.

After validating the message(s), [it will send the new data to a channel](https://github.com/ezeql/kv-service/blob/master/internal/storage/consumer.go#L62). This is used to share/sync new messages to another goroutine running in the same worker which is in charge of maintaining a synchronized map with the amount of hits for each key.

In the aforementioned goroutine, a [ticker](https://gobyexample.com/tickers) [that runs every nine seconds](https://github.com/ezeql/kv-service/blob/master/internal/storage/consumer.go#L29) will send the counts held in the worker's memory, to the database using a single query that includes all keys and their values.

Nine seconds is a value chosen after one of the requirements of the problem that states data available in the database should be no older than 10 seconds.

Metrics are pulled from Caddy every 15 seconds by Prometheus.

### Validations

`{
    'key': "uuid"
    'value': int64
}`

1. `key` must be a valid `UUID` string.
2. `value` must be equal or higher than number 1.
3. No extra fields are present in the request Payload.

## Requirements

[Make](https://en.wikipedia.org/wiki/Make_(software))

### Release

- [Docker](https://docs.docker.com/get-docker/)
- [Docker-compose](https://docs.docker.com/compose/install/)

### Dev

- [Go](https://golang.org/doc/install)
- [Air](https://github.com/cosmtrek/air)

### Benchmark

- [Hey](https://github.com/rakyll/hey)

## Make files / How to run

```make test``` runs Go tests

```make build``` builds all the images

```make db-clear``` clears the table increments

```make all``` Build the images, create the table in Postgres and runs the app.

```make dev``` hot reload dev environment

```make bench``` runs a simple benchmark using [hey](https://github.com/rakyll/hey)

## Task list

- [x] /Increment endpoint
- [x] JSON payload Validation
- [x] Requests incremented by given key
- [x] The persisted state must be, at most, ten seconds out of date.
- [x] Rate limit
- [x] Benchmarking
- [x] Proxy Metrics + Prometheus + Grafana
- [ ] Auth: [Suggested implementation](https://github.com/ezeql/kv-service/blob/master/caddy/Dockerfile#L8)
- [ ] CI / CD
- [ ] Graceful shutdown
- [ ] Retry mechanism when connecting to DB, publish messages, DLQs, transport layer,etc


## Tests

[TestHitsServer_CreateIncrement](https://github.com/ezeql/kv-service/blob/master/internal/hits/handlers/post_test.go#L113): Simple test issuing a valid request expecting a  valid response.


[TestHitsServer_Concurrent](https://github.com/ezeql/kv-service/blob/master/internal/hits/handlers/post_test.go#L146): Integration test. Concurrent stress test issuing multiple concurrent queries using different keys and counting the available keys who made to the database.

[Test_incrementRequestValid](https://github.com/ezeql/kv-service/blob/master/internal/incrementsrv/increment_test.go#L16): Input and expected output tests

[Test_fillMask](https://github.com/ezeql/kv-service/blob/master/testdata/values_test.go): Useful UUID generators and their testing.





## Postmortem

Exercise was fun to work on and allowed me to refresh on some concepts.
One of the issues I found was the fact that introducing multiple consumers pulling from the same worker would create potential deadlocks in Postgres.


The solution applied was to use a single connection per host using a single query which compacts all available keys( after every tick ) into a single query containing all of them using [UNNEST](https://stackoverflow.com/questions/20815028/how-do-i-insert-multiple-values-into-a-postgres-table-at-once).

A potential improvement would be limiting the ammount of entries a single INSERT may contain which is  currently unlimited.

Before going to production with the current solution I would invest some extra time making sure there are no single point of failures. For example, setting Redis to HA maybe by using a managed service like Elasticache.

As a final note, I lost many precious minutes with this issue [I end up submitted a PR to ](https://github.com/rakyll/hey/pull/242)
## (Some) Links digested in this project

1. https://medium.com/avitotech/how-to-work-with-postgres-in-go-bad2dabd13e4
2. https://www.reddit.com/r/golang/comments/h7ontk/how_to_use_connection_pooling_with_pgxgolang/
3. https://rafaelcn.github.io/2020/03/07/an-advice-about-postgresql-drivers-and-go.html
4. https://medium.com/@amoghagarwal/insert-optimisations-in-golang-26884b183b35>
(https://blogtitle.github.io/go-advanced-concurrency-patterns-part-2-timers/)
5. https://medium.com/@jeremieshaker/golang-ticker-best-practices-using-tickers-in-a-multi-threaded-program-without-losing-your-mind-dfc307c6de62
6. https://devandchill.com/posts/2020/05/go-lib/pq-or-pgx-which-performs-better/
7. https://www.postgresonline.com/journal/archives/347-LATERAL-WITH-ORDINALITY---numbering-sets.html
8. https://stackoverflow.com/questions/41717935/preserve-the-order-of-items-in-array-when-doing-join-in-postgres
9. https://stackoverflow.com/questions/8760419/postgresql-unnest-with-element-number
