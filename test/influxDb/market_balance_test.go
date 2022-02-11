package testinfluxDb

import (
	"context"
	"fmt"
	"os"
	"tdex-analytics/internal/core/domain"
	dbinflux "tdex-analytics/internal/infrastructure/db/influx"
	"testing"
	"time"
)

func TestInsertMarketBalance(t *testing.T) {
	ctx := context.Background()

	token := os.Getenv("TDEXA_INFLUXDB_TOKEN")

	db, err := dbinflux.NewInfluxDb(dbinflux.Config{
		Org:             "tdex-network",
		AuthToken:       token,
		DbUrl:           "http://localhost:8086",
		AnalyticsBucket: "analytics",
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := db.InsertBalance(ctx, domain.MarketBalance{
			MarketID:     "213",
			BaseBalance:  50 + i,
			BaseAsset:    "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225",
			QuoteBalance: 500 + i,
			QuoteAsset:   "6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d",
			Time:         time.Now(),
		}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGetMarketBalance(t *testing.T) {
	ctx := context.Background()

	token := os.Getenv("TDEXA_INFLUXDB_TOKEN")

	db, err := dbinflux.NewInfluxDb(dbinflux.Config{
		Org:             "tdex-network",
		AuthToken:       token,
		DbUrl:           "http://localhost:8086",
		AnalyticsBucket: "analytics",
	})
	if err != nil {
		t.Fatal(err)
	}

	market, err := db.GetBalancesForMarket(ctx, "233", time.Now().AddDate(0, 0, -1))
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(market)

}
