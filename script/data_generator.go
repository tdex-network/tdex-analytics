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

	pricesTemplate := "market_price,market_id=%v base_asset=\"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225\",base_price=50,quote_asset=\"6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d\",quote_price=500 %v\n"
	balancesTemplate := "market_balance,market_id=%v base_asset=\"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225\",base_balance=50i,quote_asset=\"6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d\",quote_balances=500i %v\n"

	start := time.Now()
	counter := time.Now()
	approxFourMonthsInHours := time.Duration(time.Hour) * 24 * 30 * 4
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
