package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/drbh/go-robinhood" // until changes get merged
	// "github.com/andrewstuart/go-robinhood"
	"github.com/davecgh/go-spew/spew"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

type BatchPrice struct {
	Price    float64
	Quantity float64
	Date     time.Time
}

type Config struct {
	Account struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"account"`
}

type Results struct {
	TotalGains  float64
	TotalLosses float64
}

type GainsLosses map[int]Results

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

	allGainsLosses := make([]GainsLosses, 1, 1)

	for i := range instrumentMap {
		instum, _ := c.GetInstrument(i)

		// if instum.Symbol != "GILD" {
		// 	continue
		// }

		fmt.Printf("\n(%s)\t%s\n", instum.Symbol, instum.Name)

		// spew.Dump(instum)

		diff := 0.0
		s := instrumentMap[i]
		var batchPricing struct {
			EndQuantity   float64
			Executions    []BatchPrice
			TotalGainLoss float64
		}

		mapOfGainsLosses := make(GainsLosses)

		data := [][]string{}

		for i := range s {
			// fmt.Println("\n-")
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

			avgCostPerShare := 0.0

			if x.Side == "buy" {

				// if a wash sale
				// was the last sale a loss within 30 days?

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

			if x.Side == "buy" {
				avgCostPerShare = batchPricing.TotalGainLoss / batchPricing.EndQuantity
			}
			if x.Side == "sell" {
				avgCostPerShare = batchPricing.TotalGainLoss / makeFloat(x.CumulativeQuantity)
			}

			dateString := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d-00:00\n",
				t.Year(), t.Month(), t.Day(),
				t.Hour(), t.Minute(), t.Second())

			isCapitalLoss := false
			if batchPricing.TotalGainLoss < 0 {
				if batchPricing.EndQuantity < makeFloat(x.CumulativeQuantity) {
					isCapitalLoss = true
				}
			}

			isCapitalGain := false
			if batchPricing.TotalGainLoss > 0 {
				if batchPricing.EndQuantity < makeFloat(x.CumulativeQuantity) {
					isCapitalGain = true
				}
			}

			_, ok := mapOfGainsLosses[t.Year()]
			if ok {
			} else {
				var val = Results{
					TotalGains:  0,
					TotalLosses: 0,
				}
				mapOfGainsLosses[t.Year()] = val
			}

			value, _ := mapOfGainsLosses[t.Year()]

			if isCapitalGain {
				value.TotalGains += batchPricing.TotalGainLoss
			}
			if isCapitalLoss {
				value.TotalLosses -= batchPricing.TotalGainLoss
			}
			mapOfGainsLosses[t.Year()] = value

			// spew.Dump(value)

			data = append(data, []string{
				// fmt.Sprintf("%d", optionIndex),
				// fmt.Sprintf(date),
				fmt.Sprintf("%.2f", x.AveragePrice),
				fmt.Sprintf("%.2f", makeFloat(x.CumulativeQuantity)),
				fmt.Sprintf(x.Side),
				fmt.Sprintf(x.State),
				fmt.Sprintf("%d", t.Year()),
				fmt.Sprintf(dateString),
				fmt.Sprintf("%.2f", profitGain),
				fmt.Sprintf("%.2f", diff),
				fmt.Sprintf("%.2f", batchPricing.TotalGainLoss),
				fmt.Sprintf("%.2f", batchPricing.EndQuantity),
				fmt.Sprintf("%d", len(batchPricing.Executions)),
				fmt.Sprintf("%.2f", avgCostPerShare),

				// fmt.Sprintf("%t", false),
				fmt.Sprintf("%t", isCapitalGain),
				fmt.Sprintf("%t", isCapitalLoss),
			})

			// reset if we cleared our shares
			if batchPricing.EndQuantity == 0 {
				batchPricing.TotalGainLoss = 0.0
			}

		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{
			"Price",
			"Quantity",
			"Side",
			"State",
			"Year",
			"Timestamp",
			"ProfitGain",
			"DaysHeld",
			"TotalGainLoss",
			"EndQuantity",
			"NumberOrders",
			"AvgCostPerShare",

			// "IsWash",
			"IsGain",
			"IsLoss",
		})

		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
		table.AppendBulk(data)
		table.Render()

		allGainsLosses = append(allGainsLosses, mapOfGainsLosses)
		// spew.Dump(mapOfGainsLosses)
	}

	fmt.Println("\n\nTOTALS FOR 2020")

	gains := 0.0
	losses := 0.0
	for _, v := range allGainsLosses {
		gains += v[2020].TotalGains
		losses += v[2020].TotalLosses
	}

	mydata := [][]string{}
	mydata = append(mydata, []string{
		fmt.Sprintf("%.2f", gains),
		fmt.Sprintf("%.2f", losses),
		fmt.Sprintf("%.2f", gains-losses),
		fmt.Sprintf("%.2f", (gains-losses)*0.3),
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Gains",
		"Losses",
		"Net",
		"Tax",
	})

	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(mydata)
	table.Render()

	fmt.Println("\n\nTOTALS FOR 2021")

	gains = 0.0
	losses = 0.0
	for _, v := range allGainsLosses {
		gains += v[2021].TotalGains
		losses += v[2021].TotalLosses
	}

	mydata = [][]string{}
	mydata = append(mydata, []string{
		fmt.Sprintf("%.2f", gains),
		fmt.Sprintf("%.2f", losses),
		fmt.Sprintf("%.2f", gains-losses),
		fmt.Sprintf("%.2f", (gains-losses)*0.3),
	})

	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Gains",
		"Losses",
		"Net",
		"Tax",
	})

	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(mydata)
	table.Render()

}

func makeFloat(val string) float64 {
	b2, _ := strconv.ParseFloat(val, 64)
	return b2
}
