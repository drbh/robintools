package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/drbh/go-robinhood"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Account struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	} `yaml:"account"`
}

func makeFloat(val string) float64 {
	b2, _ := strconv.ParseFloat(val, 64)
	return b2
}

func main() {

	argsWithoutProg := os.Args[1:]

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

	tickerSymbol := argsWithoutProg[0]
	quote, _ := c.GetQuote(tickerSymbol)

	balance := makeFloat(argsWithoutProg[1])
	stockPrice := quote[0].LastTradePrice

	fmt.Println(tickerSymbol, balance, stockPrice)

	i, err := c.GetInstrumentForSymbol(tickerSymbol)

	if err != nil {
		spew.Dump(err)
	}

	ch, err := c.GetOptionChains(i)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data := [][]string{}


	currentTime := time.Now()

	for optionIndex, date := range ch[0].ExpirationDates {

		dateInts := strings.Split(date, "-")

		y, err := strconv.Atoi(dateInts[0])
		m, err := strconv.Atoi(dateInts[1])
		d, err := strconv.Atoi(dateInts[2])

		insts, err := ch[0].GetInstrument(ctx, "call", robinhood.NewDate(y, m, d))
		if err != nil {
			spew.Dump(err)
		}
		is, err := c.MarketData(insts...)

		for _, s := range is {

			// we use ask since were pessimistic and think we can only buy at the lowest sell price
			optPriceUsed := s.AskPrice
			// optPriceUsed := s.AdjustedMarkPrice

			strike := s.BreakEvenPrice - s.AdjustedMarkPrice
			optBreakEvenUsed := strike + optPriceUsed
			wholeContracts := math.Floor(float64(balance) / (optPriceUsed * 100))
			wholeShares := math.Floor(float64(balance) / stockPrice)

			x, _ := findIntersection(
				optBreakEvenUsed, 0.0, float64(1000), (float64(1000)-optBreakEvenUsed)*(wholeContracts*100),
				stockPrice, 0.0, float64(1000), (float64(1000)*wholeShares)-(wholeShares*stockPrice),
			)



			distanceFromPrice := x - stockPrice

			if wholeContracts < 2 {
				continue
			}

			relDistance := distanceFromPrice / stockPrice
			leverage := (wholeContracts * 100) / wholeShares

			if len(argsWithoutProg) > 2 {

				filterRelDistanceThreshold := argsWithoutProg[2]
				if filterRelDistanceThreshold != "" {
					if relDistance > makeFloat(filterRelDistanceThreshold) {
						continue
					}
				}
			}
			t, _ := time.Parse("2006-01-02", date)

			daysUntilExpiration := t.Sub(currentTime).Hours() / 24

			data = append(data, []string{
				fmt.Sprintf("%d", optionIndex),
				fmt.Sprintf(date),
				fmt.Sprintf("%.2f", strike),
				fmt.Sprintf("%.2f", optPriceUsed),
				fmt.Sprintf("%.0f", wholeContracts*100),
				fmt.Sprintf("%.0f", wholeContracts),
				fmt.Sprintf("%.0f", wholeShares),
				fmt.Sprintf("%.2f", optBreakEvenUsed),
				fmt.Sprintf("%.2f", x),
				fmt.Sprintf("%.2f", distanceFromPrice),
				fmt.Sprintf("%.2f", relDistance),
				fmt.Sprintf("%.2f", leverage),
				fmt.Sprintf("%.2f", leverage/relDistance),
				fmt.Sprintf("%.4f", math.Atan(wholeShares)*180.0/math.Pi),
				fmt.Sprintf("%.4f", math.Atan(wholeContracts*100)*180.0/math.Pi),
				fmt.Sprintf("%.2f", daysUntilExpiration),
			})
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Selection",
		"Expiration",
		"Strike",
		"Price",
		"Control",
		"Contracts",
		"Shares",
		"Breakeven",
		"ContractBreak",
		"Distance",
		"RelativeDistance",
		"Leverage",
		"LeverageToDistance",
		"EquityDeg",
		"OptionDeg",
		"DaysUntilExpiration",
	})

	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()
}

func findIntersection(x1 float64, y1 float64, x2 float64, y2 float64, x3 float64, y3 float64, x4 float64, y4 float64) (float64, float64) {
	px := ((x1*y2-y1*x2)*(x3-x4) - (x1-x2)*(x3*y4-y3*x4)) / ((x1-x2)*(y3-y4) - (y1-y2)*(x3-x4))
	py := ((x1*y2-y1*x2)*(y3-y4) - (y1-y2)*(x3*y4-y3*x4)) / ((x1-x2)*(y3-y4) - (y1-y2)*(x3-x4))
	return px, py
}
