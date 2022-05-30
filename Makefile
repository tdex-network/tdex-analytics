.PHONY: build-server build-cli build proto si iid fmt vet clean ci testall \
testinfluxdb testapp pg createdb dropdb createtestdb droptestdb pgcreatetestdb dev

## build the tdexa server
build-server:
	@export GO111MODULE=on; \
	env go build -tags netgo -ldflags="-s -w" -o bin/tdexad cmd/tdexad/main.go

## build only the tdexa CLI
build-cli:
	@export GO111MODULE=on; \
	env go build -tags netgo -ldflags="-s -w" -o bin/tdexa cmd/tdexa/*.go

## build the project (CLI + server)
build: build-server build-cli

## proto: compile proto files
proto:
	buf generate

## pit: provision influxdb used for testing test
pit:
	INFLUXDB_BUCKET=analytics INFLUXDB_ORG=tdex-network INFLUXDB_PASSWORD=admin123 INFLUXDB_USERNAME=admin ./script/provision_influxdb_test.sh

## tor: setup client tor proxy
tor:
	docker run -d -p 9050:9050 --name=tor -v `pwd`/tor-proxy-conf:/etc/tor connectical/tor

## fmt: Go Format
fmt:
	@echo "Gofmt..."
	@if [ -n "$(gofmt -l .)" ]; then echo "Go code is not formatted"; exit 1; fi

## vet: code analysis
vet:
	@echo "Vet..."
	@go vet ./...

## clean: cleans the binary
clean:
	@echo "Cleaning..."
	@go clean

## ci: continuous integration
ci: clean fmt vet testall

# testall: test all
testall: testpgdb testinfluxdb testapp

# testinfluxdb: test influxdb
testinfluxdb:
	@echo "Testing influxdb..."
	go test -v -count=1 -race ./test/influx-db/...

## testpgdb: tests only pg db
testpgdb:
	@echo "Testing database layer..."
	go test -v -count=1 -race -timeout 30s ./test/pg-db/...

# testapp: test application layer
testapp:
	@echo "Testing influxdb..."
	go test -v -count=1 -race ./test/application/...

## pg: starts postgres db inside docker container
pg:
	docker run --name tdexa-postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres

## createdb: create db inside docker container
createdb:
	docker exec tdexa-postgres createdb --username=root --owner=root tdexa

## dropdb: drops db inside docker container
dropdb:
	docker exec tdexa-postgres dropdb tdexa

## createtestdb: create test db inside docker container
createtestdb:
	docker exec tdexa-postgres createdb --username=root --owner=root tdexa-test

## droptestdb: drops test db inside docker container
droptestdb:
	docker exec tdexa-postgres dropdb tdexa-test

## pgcreatetestdb: starts docker container and creates test db, used in CI
pgcreatetestdb:
	chmod u+x ./script/create_testdb
	./script/create_testdb

## dev: create dev env
dev:
	 docker-compose --env-file .env.dev up -d --build --force-recreate 

## dev-down: stop dev env, remove volumes
dev-down:
	docker-compose --env-file .env.dev down -v