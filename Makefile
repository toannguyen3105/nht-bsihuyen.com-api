server:
	go run main.go

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

include app.env

migrateup:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose up 1

migratedown1:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose down 1

dropdb:
	migrate -path db/migration -database "$(DB_SOURCE)" -verbose drop

createdb:
	createdb --username=tony --owner=tony bsihuyen

mock:
	mockgen -package mock_sqlc -destination db/mock/store.go github.com/toannguyen3105/nht-bsihuyen.com-api/db/sqlc Store

.PHONY: server sqlc test mock migrateup migrateup1 migratedown1 dropdb createdb
