#!/bin/bash
cd "$(dirname "$0")"
PGPASSWORD=$POSTGRES_PASSWORD psql -h localhost -U $POSTGRES_USER -f ./migrations/migration_1_up.sql
