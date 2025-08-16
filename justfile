#!/usr/bin/just

up-migrate:
	#!/bin/bash
	cd sql/schema/
	goose postgres "postgres://postgres:@localhost:5432/chirpy" up
