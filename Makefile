.PHONY: up down run test test-api lint load-test bench start

up:
	docker-compose up -d --build

down:
	docker-compose down

start:
	./scripts/start.sh

run:
	go run cmd/api/main.go

test:
	go test -v ./...

test-api:
	./scripts/test_api.sh

lint:
	golangci-lint run

load-test:
	./scripts/load_test.sh

bench:
	go run cmd/bench/main.go -rate=100 -duration=60s
