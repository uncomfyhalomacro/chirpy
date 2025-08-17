#!/usr/bin/just

run:
	go build && ./chirpy

up-migrate:
	#!/bin/bash
	source .env
	cd sql/schema/
	goose postgres "$DB_URL" up

drop-migrate:
	#!/bin/bash
	source .env
	cd sql/schema/
	goose postgres "$DB_URL" down

drop-migrate-all:
	#!/bin/bash
	source .env
	cd sql/schema/
	goose postgres "$DB_URL" down-to 0

sqlc-generate:
	sqlc generate
