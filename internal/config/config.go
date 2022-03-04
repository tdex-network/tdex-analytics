package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	TorProxyUrl             = "TOR_PROXY_URL"
	RegistryUrl             = "REGISTRY_URL"
	LogLevelKey             = "LOG_LEVEL"
	PriceAmount             = "PRICE_AMOUNT"
	JobPeriodInMinutes      = "JOB_PERIOD_IN_MINUTES"
)

var vip *viper.Viper

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
	vip.SetDefault(DbNameKey, "tdexa")
	vip.SetDefault(DbMigrationPath, "file://internal/infrastructure/db/pg/migrations")
	vip.SetDefault(DbInsecure, true)
	vip.SetDefault(AwsRegion, "eu-central-1")
	vip.SetDefault(TorProxyUrl, "127.0.0.1:9050")
	vip.SetDefault(RegistryUrl, "https://raw.githubusercontent.com/tdex-network/tdex-registry/master/registry.json")
	vip.SetDefault(LogLevelKey, int(log.DebugLevel))
	vip.SetDefault(PriceAmount, 100)
	vip.SetDefault(JobPeriodInMinutes, "5")

	if vip.GetString(InfluxDbAuthToken) == "" {
		log.Fatalln("influx_db auth token not provided")
	}

	log.SetLevel(log.Level(GetInt(LogLevelKey)))
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
