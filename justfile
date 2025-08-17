#!/usr/bin/just

up-migrate:
	#!/bin/bash
	cd sql/schema/
	goose postgres "postgres://postgres:@localhost:5432/chirpy" up

drop-migrate:
	#!/bin/bash
	cd sql/schema/
	goose postgres "postgres://postgres:@localhost:5432/chirpy" down

drop-migrate-all:
	#!/bin/bash
	cd sql/schema/
	goose postgres "postgres://postgres:@localhost:5432/chirpy" down-to 0
	
