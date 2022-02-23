package config

import (
	"github.com/btcsuite/btcutil"
	"github.com/spf13/viper"
	"log"
)

const (
	// GrpcServerPortKey is port on which grpc server is running
	GrpcServerPortKey       = "SERVER_PORT"
	InfluxDbOrg             = "INFLUXDB_INIT_ORG"
	InfluxDbAuthToken       = "INFLUXDB_TOKEN"
	InfluxDbUrl             = "INFLUXDB_URL"
	InfluxDbAnalyticsBucket = "INFLUXDB_INIT_BUCKET"
	DbUserKey               = "DB_USER"
	DbPassKey               = "DB_PASS"
	DbHostKey               = "DB_HOST"
	DbPortKey               = "DB_PORT"
	DbNameKey               = "DB_NAME"
	DbMigrationPath         = "DB_MIGRATION_PATH"
	DbInsecure              = "DB_INSECURE"
	AwsRegion               = "AWSREGION"
)

var vip *viper.Viper
var defaultDataDir = btcutil.AppDataDir("tdexad", false)

func init() {
	vip = viper.New()
	vip.SetEnvPrefix("TDEXA")
	vip.AutomaticEnv()

	//TODO update default config to prod
	vip.SetDefault(GrpcServerPortKey, 9000)
	vip.SetDefault(InfluxDbOrg, "tdex-network")
	vip.SetDefault(InfluxDbUrl, "http://localhost:8086")
	vip.SetDefault(InfluxDbAnalyticsBucket, "analytics")
	vip.SetDefault(DbUserKey, "root")
	vip.SetDefault(DbPassKey, "secret")
	vip.SetDefault(DbHostKey, "127.0.0.1")
	vip.SetDefault(DbPortKey, 5432)
	vip.SetDefault(DbNameKey, "tdexa-test")
	vip.SetDefault(DbMigrationPath, "file://internal/infrastructure/db/pg/migrations")
	vip.SetDefault(DbInsecure, true)
	vip.SetDefault(AwsRegion, "eu-central-1")

	if vip.GetString(InfluxDbAuthToken) == "" {
		log.Fatalln("influx_db auth token not provided")
	}

	//TODO influxDB health check
}

func GetBool(key string) bool {
	return vip.GetBool(key)
}

func GetString(key string) string {
	return vip.GetString(key)
}

func GetInt(key string) int {
	return vip.GetInt(key)
}
