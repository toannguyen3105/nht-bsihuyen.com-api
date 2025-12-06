#!/bin/sh

set -e

echo "run db migration"
# Ensure DB_SOURCE is set
if [ -z "$DB_SOURCE" ]; then
  echo "DB_SOURCE is not set"
  exit 1
fi

/app/migrate -path /app/db/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"
