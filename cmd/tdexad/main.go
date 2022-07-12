package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"tdex-analytics/internal/config"
	"tdex-analytics/internal/core/application"
	dbinflux "tdex-analytics/internal/infrastructure/db/influx"
	dbpg "tdex-analytics/internal/infrastructure/db/pg"
	tdexagrpc "tdex-analytics/internal/interface/grpc"
	"tdex-analytics/pkg/rater"
	tdexmarketloader "tdex-analytics/pkg/tdex-market-loader"
)

func main() {
	influxDbSvc, err := dbinflux.New(dbinflux.Config{
		Org:             config.GetString(config.InfluxDbOrg),
		AuthToken:       config.GetString(config.InfluxDbAuthToken),
		DbUrl:           config.GetString(config.InfluxDbUrl),
		AnalyticsBucket: config.GetString(config.InfluxDbAnalyticsBucket),
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	marketRepository, err := dbpg.New(dbpg.DbConfig{
		DbUser:             config.GetString(config.DbUserKey),
		DbPassword:         config.GetString(config.DbPassKey),
		DbHost:             config.GetString(config.DbHostKey),
		DbPort:             config.GetInt(config.DbPortKey),
		DbName:             config.GetString(config.DbNameKey),
		MigrationSourceURL: config.GetString(config.DbMigrationPath),
		DbInsecure:         config.GetBool(config.DbInsecure),
		AwsRegion:          config.GetString(config.AwsRegion),
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	tdexMarketLoaderSvc := tdexmarketloader.NewService(
		config.GetString(config.TorProxyUrl),
		config.GetString(config.RegistryUrl),
		config.GetInt(config.PriceAmount),
	)

	marketLoaderSvc := application.NewMarketsLoaderService(
		marketRepository,
		tdexMarketLoaderSvc,
	)

	marketBalanceSvc := application.NewMarketBalanceService(
		influxDbSvc,
		marketRepository,
		tdexMarketLoaderSvc,
		config.GetString(config.JobPeriodInMinutes),
	)

	raterSvc := rater.NewExchangeRateClient(config.GetAssetCurrencyPair())

	marketPriceSvc := application.NewMarketPriceService(
		influxDbSvc,
		marketRepository,
		tdexMarketLoaderSvc,
		config.GetString(config.JobPeriodInMinutes),
		raterSvc,
	)

	opts := tdexagrpc.WithInsecureGrpcGateway()
	certFile := config.GetString(config.SSLCertPathKey)
	keyFile := config.GetString(config.SSLKeyPathKey)
	if certFile != "" && keyFile != "" {
		opts = tdexagrpc.WithTls(certFile, keyFile)
	}

	marketSvc := application.NewMarketService(marketRepository)

	tdexad, err := tdexagrpc.NewServer(
		strconv.Itoa(config.GetInt(config.GrpcServerPortKey)),
		marketBalanceSvc,
		marketPriceSvc,
		marketLoaderSvc,
		marketSvc,
		opts,
	)

	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	errC := tdexad.Start(ctx, stop)
	if err := <-errC; err != nil {
		log.Panicf("tdex-analytics daemon server noticed error while running: %s", err)
	}
}
