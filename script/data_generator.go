package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

//generates prices and balances data for 4 previous months,
func main() {
	filePrices, err := os.OpenFile("./script/prices.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	pricesWriter := bufio.NewWriter(filePrices)

	fileBalances, err := os.OpenFile("./script/balances.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	balancesWriter := bufio.NewWriter(fileBalances)

	//line protocol https://docs.influxdata.com/influxdb/v1.8/write_protocols/line_protocol_tutorial/
	pricesTemplate := "market_price,market_id=%v base_price=50,quote_price=500 %v\n"
	balancesTemplate := "market_balance,market_id=%v base_balance=50i,quote_balance=500i %v\n"

	start := time.Now()
	counter := time.Now()
	approxFourMonthsInHours := time.Duration(time.Hour) * 24 * 30 * 4
	//generate prices for 4 previous months
	for {
		if start.Sub(counter) > approxFourMonthsInHours {
			break
		}

		counter = counter.Add(-time.Minute * 5)

		_, err := pricesWriter.WriteString(fmt.Sprintf(pricesTemplate, 1, counter.UnixNano()))
		if err != nil {
			panic(err)
		}

		_, err = pricesWriter.WriteString(fmt.Sprintf(pricesTemplate, 2, counter.UnixNano()))
		if err != nil {
			panic(err)
		}

		_, err = balancesWriter.WriteString(fmt.Sprintf(balancesTemplate, 1, counter.UnixNano()))
		if err != nil {
			panic(err)
		}

		_, err = balancesWriter.WriteString(fmt.Sprintf(balancesTemplate, 2, counter.UnixNano()))
		if err != nil {
			panic(err)
		}
	}
	if err := pricesWriter.Flush(); err != nil {
		panic(err)
	}
	if err := filePrices.Close(); err != nil {
		panic(err)
	}

	if err := balancesWriter.Flush(); err != nil {
		panic(err)
	}
	if err := fileBalances.Close(); err != nil {
		panic(err)
	}

}
