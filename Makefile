ifneq (,$(wildcard ./.env))
    include .env
    export
endif


.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]


.PHONY: run/api
run/api:
	@go run ./cmd/api



.PHONY: db/psql
db/psql:
	psql ${DB_DSN}



.PHONY: db/migrations/new
db/migrations/new:
	@echo "Creating migration files for ${name}..."
	migrate create -ext sql -dir ./db/migrations -seq ${name}


.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo "Running up migrations..."
	migrate -database ${DB_DSN} -path ./db/migrations up
	


.PHONY: audit
audit: vendor
	@echo "Formatting code..."
	go fmt ./...
	@echo "Vetting code..."
	go vet ./...
	staticcheck ./...
	@echo "Running tests..."
	go test -race -vet=off ./...


.PHONY: vendor
vendor:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy
	go mod verify
	@echo "Vendoring dependencies..."
	go mod vendor



.PHONY: sqlc
sqlc:
	sqlc generate


