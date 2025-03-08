#!/bin/sh

set -x
goose postgres "host=$DB_HOST port=$DB_PORT user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME sslmode=$DB_SSLMODE" up
migration_status=$?
set +x

if [ $migration_status -eq 0 ]; then
    echo "Migrations completed successfully"
    exit 0
else
    echo "Migrations failed with status $migration_status"
    exit 1
fi
