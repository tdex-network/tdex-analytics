.PHONY: build-server build-cli build proto provision

## build the tower server
build-server:
	@export GO111MODULE=on; \
	env go build -tags netgo -ldflags="-s -w" -o bin/tdexad cmd/tdexad/main.go

## build only the tower CLI
build-cli:
	@export GO111MODULE=on; \
	env go build -tags netgo -ldflags="-s -w" -o bin/tdexa cmd/tdexa/*.go

## build the project (CLI + server)
build: build-server build-cli

## proto: compile proto files
proto:
	buf generate

## provision: start tdexa daemon and influxdb and insert rand balances and prices data
provision:
	INFLUXDB_BUCKET=analytics INFLUXDB_ORG=tdex-network INFLUXDB_PASSWORD=admin123 INFLUXDB_USERNAME=admin ./script/provision.sh