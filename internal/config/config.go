package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// GrpcServerPortKey is port on which grpc server is running
	GrpcServerPortKey = "SERVER_PORT"
	// InfluxDbOrg is organisation name
	InfluxDbOrg = "INFLUXDB_INIT_ORG"
	// InfluxDbAuthToken is token used as cred by tdexad
	InfluxDbAuthToken = "INFLUXDB_TOKEN"
	// InfluxDbUrl is db url
	InfluxDbUrl = "INFLUXDB_URL"
	// InfluxDbAnalyticsBucket is bucket that tdexd using to store data
	InfluxDbAnalyticsBucket = "INFLUXDB_INIT_BUCKET"
	// DbUserKey is postgres db user used by tdexd
	DbUserKey = "DB_USER"
	// DbPassKey is postgres db pass used by tdexd
	DbPassKey = "DB_PASS"
	// DbHostKey is postgres db host
	DbHostKey = "DB_HOST"
	// DbPortKey is postgres db port
	DbPortKey = "DB_PORT"
	// DbNameKey is postgres db name
	DbNameKey = "DB_NAME"
	// DbMigrationPath is postgres db migration path
	DbMigrationPath = "DB_MIGRATION_PATH"
	// DbInsecure is bool based on which postgres db conn will be created to AWS RDS or self-hosted PG DB
	DbInsecure = "DB_INSECURE"
	// AwsRegion region, used for connectiog to AWS RDS
	AwsRegion = "AWSREGION"
	// TorProxyUrl is tor client proxy url used to connect to liquidity providers using onion
	TorProxyUrl = "TOR_PROXY_URL"
	// RegistryUrl is url with info about available liquidity providers
	RegistryUrl = "REGISTRY_URL"
	// LogLevelKey is log level used by tdexa
	LogLevelKey = "LOG_LEVEL"
	// PriceAmount is amount used when invoking tdex-daemon MarketPrice RPC
	PriceAmount = "PRICE_AMOUNT"
	// JobPeriodInMinutes is recurring interval for running fetch balance/price jobs
	JobPeriodInMinutes = "JOB_PERIOD_IN_MINUTES"
	// SSLCertPathKey is the path to the SSL certificate
	SSLCertPathKey = "SSL_CERT"
	// SSLKeyPathKey is the path to the SSL private key
	SSLKeyPathKey = "SSL_KEY"
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

	if err := validateTlsKeys(); err != nil {
		log.WithError(err).Panic("invalid tls keys")
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

func validateTlsKeys() error {
	certPath, keyPath := vip.GetString(SSLCertPathKey), vip.GetString(SSLKeyPathKey)
	if (certPath != "" && keyPath == "") || (certPath == "" && keyPath != "") {
		return fmt.Errorf("tls requires both key and certificate when enabled")
	}

	return nil
}
