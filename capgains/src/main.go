package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/drbh/go-robinhood" // until changes get merged
	// "github.com/andrewstuart/go-robinhood"
	"github.com/davecgh/go-spew/spew"
	"gopkg.in/yaml.v2"
)

type BatchPrice struct {
	Price    float64
	Quantity float64
	Date     time.Time
}

type CurrentOrderImpact struct {
	AveragePrice float64
	Quantity     float64
	GainLoss     float64
	Difference   float64
	Side         string
	State        string
	Time         time.Time
}

type Config struct {
	Account struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"account"`
}

func main() {

	f, err := os.Open("/Users/drbh/.robintools/config.yml")
	if err != nil {
		spew.Dump(err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		spew.Dump(err)
	}

	o := &robinhood.OAuth{
		Username: cfg.Account.Email,
		Password: cfg.Account.Password,
	}

	c, err := robinhood.Dial(o)
	if err != nil {
		spew.Dump(err)
	}

	instrumentMap := make(map[string][]robinhood.OrderOutput)
	res, err := c.AllOrders()

	for _, x := range res {

		// this means it was cancelled
		if x.AveragePrice == 0.0 {
			continue
		}

		instrumentMap[x.Instrument] = append(instrumentMap[x.Instrument], x)
	}

	for i := range instrumentMap {

		instum, _ := c.GetInstrument(i)

		fmt.Println("\n\n\n##########################")

		fmt.Println(instum.Name)
		fmt.Println(instum.Symbol)

		// spew.Dump(instum)

		diff := 0.0
		s := instrumentMap[i]
		var batchPricing struct {
			EndQuantity   float64
			Executions    []BatchPrice
			TotalGainLoss float64
		}
		for i := range s {
			fmt.Println("\n-")
			x := s[len(s)-1-i]
			t, err := time.Parse(time.RFC3339, x.LastTransactionAt)

			if err != nil {
				spew.Dump(err)
			}

			if len(batchPricing.Executions) > 0 {
				t2 := batchPricing.Executions[0].Date
				diff = t.Sub(t2).Hours() / 24
			}

			profitGain := 0.0
			if x.Side == "buy" {
				batchPricing.EndQuantity += makeFloat(x.CumulativeQuantity)
				batchPricing.Executions = append(batchPricing.Executions, BatchPrice{Date: t, Price: x.AveragePrice, Quantity: makeFloat(x.CumulativeQuantity)})
				profitGain -= x.AveragePrice * makeFloat(x.CumulativeQuantity)
			}
			if x.Side == "sell" {
				batchPricing.EndQuantity -= makeFloat(x.CumulativeQuantity)

				s := batchPricing.Executions
				for i := range s {
					value := s[len(s)-1-i]
					subtractedQty := value.Quantity - makeFloat(x.CumulativeQuantity)
					if subtractedQty <= 0 {
						// we need to use whole order
						cashIn := value.Quantity * x.AveragePrice //value.Price
						profitGain += cashIn
						_, batchPricing.Executions = batchPricing.Executions[0], batchPricing.Executions[1:]
					} else {
						// fmt.Println(value)
						cashIn := makeFloat(x.CumulativeQuantity) * x.AveragePrice
						profitGain += cashIn
						// we only use part of the order
						batchPricing.Executions[0] = BatchPrice{Price: value.Price, Quantity: subtractedQty}
					}
				}
			}

			batchPricing.TotalGainLoss += profitGain

			impact := CurrentOrderImpact{
				AveragePrice: x.AveragePrice,
				Quantity:     makeFloat(x.CumulativeQuantity),
				Side:         x.Side,
				State:        x.State,
				Time:         t,
				GainLoss:     profitGain,
				Difference:   diff,
			}

			d, err := yaml.Marshal(&batchPricing)
			if err != nil {
			}
			fmt.Printf("%s", string(d))

			m, err := yaml.Marshal(&impact)
			if err != nil {
			}
			fmt.Printf("%s", string(m))

		}
	}
}

func makeFloat(val string) float64 {
	b2, _ := strconv.ParseFloat(val, 64)
	return b2
}
