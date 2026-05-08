include .envrc
.PHONY: run ui migrate-up migrate-down migrate-force migrate-version

run:
	@echo  'Running application…'
	@go run ./cmd/api

ui:
	@echo 'Running UI server…'
	@go run ./ui.go

migrate-up:
	migrate -path=./migrations -database=$(DSN) up

migrate-down:
	migrate -path=./migrations -database=$(DSN)	 down

migrate-force:
	migrate -path=./migrations -database=$(DSN) force 2

migrate-version:
	migrate -path=./migrations -database=$(DSN) version
