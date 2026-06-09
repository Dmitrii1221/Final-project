include .env
export

.PHONY: migrate-up migrate-down migrate-status

migrate-up:
	goose -dir migrations postgres "$(POSTGRES_DSN)" up

migrate-down:
	goose -dir migrations postgres "$(POSTGRES_DSN)" down

migrate-status:
	goose -dir migrations postgres "$(POSTGRES_DSN)" status