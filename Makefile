APP_NAME=nimbus

.PHONY: run
run:
	go run ./cmd/api

.PHONY: test
test:
	go test ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker compose down

.PHONY: logs
logs:
	docker compose logs -f

.PHONY: migrate-up
migrate-up:
	soda migrate up

.PHONY: migrate-down
migrate-down:
	soda migrate down
