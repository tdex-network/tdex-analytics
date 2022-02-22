package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

//generates prices and balances data for 5 months, going from beginning of 2022
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

	pricesTemplate := "market_price,market_id=%v base_asset=\"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225\",base_price=50i,quote_asset=\"6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d\",quote_price=500i %v\n"
	balancesTemplate := "market_balance,market_id=%v base_asset=\"5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225\",base_balance=50i,quote_asset=\"6f0279e9ed041c3d710a9f57d0c02928416460c4b722ae3457a11eec381c526d\",quote_balances=500i %v\n"
	t := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	for {
		if t.Month() == 5 {
			break
		}

		t = t.Add(time.Minute * 5)

		_, err := pricesWriter.WriteString(fmt.Sprintf(pricesTemplate, 1, t.UnixNano()))
		if err != nil {
			panic(err)
		}

		_, err = pricesWriter.WriteString(fmt.Sprintf(pricesTemplate, 2, t.UnixNano()))
		if err != nil {
			panic(err)
		}

		_, err = balancesWriter.WriteString(fmt.Sprintf(balancesTemplate, 1, t.UnixNano()))
		if err != nil {
			panic(err)
		}

		_, err = balancesWriter.WriteString(fmt.Sprintf(balancesTemplate, 2, t.UnixNano()))
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
