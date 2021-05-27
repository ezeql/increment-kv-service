dev:
	docker-compose -f docker-compose-dev.yml up -d && DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres REDIS=localhost:6379 air
test:
	go test ./...
build:
	docker-compose build
run:
	docker-compose  up --build -d
db-clear:
	docker-compose exec db sh -c "./scripts/db/delete-db.sh"
db-migrate:
	docker-compose exec db sh -c "./scripts/db/create-db.sh"
bench:
	hey -c 500 -n 100000 -m POST -T 'application/json' -d '{"key": "dddddddd-1234-5678-1234-567812345678","value": 1}' http://localhost:3333/increment

all: test run db-migrate