compose-db-up:
	docker compose -f docker-compose.yml up -d postgres

compose-db-down:
	docker compose -f docker-compose.yml down postgres

swag:
	swag init -ot yaml --parseDependency --parseInternal -g cmd/app/main.go

migrate-up:
	migrate -database postgres://postgres:postgres@localhost:5432/filmlib?sslmode=disable -path db/migrations up

migrate-down:
	migrate -database postgres://postgres:postgres@localhost:5432/filmlib?sslmode=disable -path db/migrations down
	
create-migration:
	migrate create -ext sql -dir ./db/migrations $(name)