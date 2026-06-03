include .env

.PHONY: migrate-create migrate-up river-up migrate-tests-up migrate-tests-down

migrate-create:
	@migrate create -ext sql -dir db/migrations -seq $(f)

migrate-up:
	@migrate -database $(DATABASE_URL) -path db/migrations up

migrate-down:
	@migrate -database $(DATABASE_URL) -path db/migrations down

river-up:
	@river migrate-up --database-url "$(DATABASE_URL)"

migrate-tests-up:
	@migrate -database $(TEST_DATABASE_URL) -path db/migrations up

migrate-tests-down:
	@migrate -database $(TEST_DATABASE_URL) -path db/migrations down