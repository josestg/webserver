SHELL := /bin/bash

tidy:
	go mod tidy
	go mod vendor

run:
	go run app/api/main.go