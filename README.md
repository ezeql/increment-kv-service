# Platform Engineer Project

## General Description

The flow starts with a client action sending a POST request with a JSON encoded body.

The first layer to take action in the system is [Caddy](https://caddyserver.com/), a reverse proxy. It will distribute the requests from the clients between all available downstream HTTP servers.

These instances that will validate the JSON payload and, if valid, will push a message to a messaging queue implemented in top of Redis using a library called [RMQ](https://github.com/adjust/rmq). It uses the publish-subscribe pattern.

A predefined set of workers (worker-db) are subscribed to this queue.
These will pickup and distribute batches of up to N items from it.

After validating the message(s), it will send the new data to a channel. This is used to share/sync new messages to another goroutine running in the same worker which is in charge of maintaining a synchronized map with the amount of hits for each key.

In the aforementioned goroutine, a [ticker](https://gobyexample.com/tickers) that runs every nine seconds will send the counts hold in the worker's memory, to the database using a single query that includes all keys and their values.

Nine seconds is a value chosen after one of the requirements of the problem that states data available in the database should be no older than 10 seconds.

### Validations

`{
    'key': "uuid"
    'value': int64
}`

1. `key` must be a valid `UUID` string.
2. `value` must be equal or higher than number 1.
3. No extra fields are present in the request Payload.

## Requirements

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

```all``` Build the images, create the table in Postgres and runs the app.

```make dev``` hot reload dev environment

```make bench``` runs a simple benchmark using [hey](https://github.com/rakyll/hey)

## Task list

- [x] /Increment endpoint
- [x] JSON payload Validation
- [x] Requests incremented by given key
- [x] The persisted state must be, at most, ten seconds out of date.
- [x] Rate limit
- [x] Benchmarking
- [ ] Metrics
- [ ] Auth
- [ ] CI / CD
- [ ] Graceful shutdown
- [ ] Retry mechanism when connecting to DB, publish messages, DLQs, transport layer,etc


## Tests

TBD

## Postmortem

TBD



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
