Create Migration
migrate create -ext sql -dir db/migration -seq [migration-name]

Install Sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest