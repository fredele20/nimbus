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

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: migrate-up
migrate-up:
	migrate -path migrations \
		-database "postgres://postgres:postgres@localhost:5432/nimbus?sslmode=disable" \
		up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations \
		-database "postgres://postgres:postgres@localhost:5432/nimbus?sslmode=disable" \
		down 1

.PHONY: migrate-version
migrate-version:
	migrate -path migrations \
		-database "postgres://postgres:postgres@localhost:5432/nimbus?sslmode=disable" \
		version
