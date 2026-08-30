include app.env
export

postgres: 
	docker run -p 5432:5432 --name postgres18 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:18

createdb: 
	docker exec -it postgres18 createdb --username=root simple_bank

dropdb: 
	docker exec -it postgres18 dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup-rds: 
	migrate -path db/migration -database "$(RDS_DB_SOURCE)" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

sqlc: 
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/rdozzi/simple_bank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migrateup-rds migrateup1 migratedown migratedown1 sqlc test server mock 