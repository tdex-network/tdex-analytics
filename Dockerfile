# first image used to build the sources
FROM golang:1.17-buster AS builder

ENV GO111MODULE=on \
    GOOS=linux \
    CGO_ENABLED=1 \
    GOARCH=amd64

WORKDIR /tdexa

COPY . .

RUN go mod download

RUN go build -ldflags="-s -w " -o ./bin/tdexad cmd/tdexad/main.go
RUN go build -ldflags="-X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${DATE}'" -o ./bin/tdexa cmd/tdexa/*

# Second image, running the towerd executable
FROM debian:buster-slim

ENV TDEXA_DB_MIGRATION_PATH="file://"

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app

COPY --from=builder /tdexa/bin/tdexad .
COPY --from=builder /tdexa/bin/tdexa .
COPY --from=builder /tdexa/internal/infrastructure/db/pg/migrations/* ./

RUN install tdexa /bin

ENTRYPOINT ["./tdexad"]
