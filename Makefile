include .env

.PHONY: migrate-create

migrate-create:
	@migrate create -ext sql -dir db/migrations -seq $(f)

migrate-up:
	@migrate -database $(DATABASE_URL) -path db/migrations up

migrate-down:
	@migrate -database $(DATABASE_URL) -path db/migrations down
