package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strings"
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
	// ExplorerUrl is explorer url used by tdexa
	ExplorerUrl = "EXPLORER_URL"
	//AssetCurrencyPair is the asset currency pair used by tdexa,
	//format: asset_hash:currency, asset_hash/currency pairs should be delimited by comma
	//example: 0x0000000000000000000000000000000000000000:LBTC,0x0000000000000000000000000000000000000000:USDT
	AssetCurrencyPair = "ASSET_CURRENCY_PAIRS"
)

var (
	vip *viper.Viper

	assetCurrencyPair = map[string]string{
		"6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d": "bitcoin",
		"ce091c998b83c78bb71a632313ba3760f1763d9cfcffae02258ffa9865a37bd2": "usd",
		"0e99c1a6da379d1f4151fb9df90449d40d0608f6cb33a5bcbfc8c265f42bab0a": "cad",
	}
)

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
	vip.SetDefault(ExplorerUrl, "https://blockstream.info/liquid/api/")

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

func GetAssetCurrencyPair() map[string]string {
	response := assetCurrencyPair
	if vip.GetString(AssetCurrencyPair) != "" {
		response = make(map[string]string)
		for _, pair := range strings.Split(vip.GetString(AssetCurrencyPair), ",") {
			parts := strings.Split(pair, ":")
			if len(parts) != 2 {
				log.Fatalf("invalid asset currency pair: %s", pair)
			}
			response[parts[0]] = parts[1]
		}
	}
	return response
}
