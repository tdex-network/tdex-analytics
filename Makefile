.PHONY: build-server build-cli build proto si iid fmt vet clean ci testall influxdb

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

## pit: provision influxdb used for testing test
pit:
	INFLUXDB_BUCKET=analytics INFLUXDB_ORG=tdex-network INFLUXDB_PASSWORD=admin123 INFLUXDB_USERNAME=admin ./script/provision_influxdb_test.sh

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
testall: influxdb

# influxdb: test influxdb
influxdb:
	@echo "Testing influxdb..."
	go test -v -count=1 -race ./test/influx-db/...