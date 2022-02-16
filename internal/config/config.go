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
)

var vip *viper.Viper
var defaultDataDir = btcutil.AppDataDir("tdexad", false)

func init() {
	vip = viper.New()
	vip.SetEnvPrefix("TDEXA")
	vip.AutomaticEnv()

	vip.SetDefault(GrpcServerPortKey, 9000)
	vip.SetDefault(InfluxDbOrg, "tdex-network")
	vip.SetDefault(InfluxDbUrl, "http://localhost:8086")
	vip.SetDefault(InfluxDbAnalyticsBucket, "analytics")

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
