#!/bin/bash
# generate prices and balances test data
go run ./script/data_generator.go
# create config file
docker run --rm influxdb:2.1.1 influxd print-config > ./influxdb-conf/config.yml
# create the container
docker run --name influxdb -d \
  -p 8086:8086 \
  --volume `pwd`/influxdb-data:/var/lib/influxdb2 \
  --volume `pwd`/influxdb-conf/config.yml:/etc/influxdb2/config.yml \
  --volume `pwd`/script/balances.txt:/balances.txt \
  --volume `pwd`/script/prices.txt:/prices.txt \
  influxdb:2.1.1
# wait until the database server is ready
until docker exec influxdb influx ping
do
  echo "Retrying..."
  sleep 5
done
# configure influxdb
docker exec influxdb influx setup \
  --bucket "$INFLUXDB_BUCKET" \
  --org "$INFLUXDB_ORG" \
  --password "$INFLUXDB_PASSWORD" \
  --username "$INFLUXDB_USERNAME" \
  --force
# get the token
export TDEXA_INFLUXDB_TOKEN=$(docker exec influxdb influx auth list | awk -v username="$INFLUXDB_USERNAME" '$5 ~ username {print $4 " "}')
echo "InfluxDB token: ${TDEXA_INFLUXDB_TOKEN}"
echo "TDEXA_INFLUXDB_TOKEN=${TDEXA_INFLUXDB_TOKEN}" >> $GITHUB_ENV
# insert data
docker exec influxdb influx write -b analytics -o tdex-network -f balances.txt
docker exec influxdb influx write -b analytics -o tdex-network -f prices.txt
