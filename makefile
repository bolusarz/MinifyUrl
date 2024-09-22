postgres:
	docker run --name postgres16 --network bank-network -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine -p 5432:5432

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root url_mini

dropdb:
	docker exec -it postgres16 dropdb url_mini

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/url_mini?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/url_mini?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/url_mini?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/url_mini?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/bolusarz/urlmini/db/sqlc Store

test:
	go test -v -cover -count 1 ./...

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test
