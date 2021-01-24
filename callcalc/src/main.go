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
	// "github.com/olekukonko/tablewriter"
	static "github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"

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

func fetchConfig() Config {
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
	return cfg
}

func getChains(c *robinhood.Client, tickerSymbol string) []*robinhood.OptionChain {
	// fmt.Println(tickerSymbol, balance, stockPrice)
	i, err := c.GetInstrumentForSymbol(tickerSymbol)
	if err != nil {
		spew.Dump(err)
	}
	ch, err := c.GetOptionChains(i)
	if err != nil {
		spew.Dump(err)
	}
	return ch
}


func findIntersection(x1 float64, y1 float64, x2 float64, y2 float64, x3 float64, y3 float64, x4 float64, y4 float64) (float64, float64) {
	px := ((x1*y2-y1*x2)*(x3-x4) - (x1-x2)*(x3*y4-y3*x4)) / ((x1-x2)*(y3-y4) - (y1-y2)*(x3-x4))
	py := ((x1*y2-y1*x2)*(y3-y4) - (y1-y2)*(x3*y4-y3*x4)) / ((x1-x2)*(y3-y4) - (y1-y2)*(x3-x4))
	return px, py
}


func getMarketData(
	mel melody.Melody,
	c *robinhood.Client,
	date string,
	ch []*robinhood.OptionChain,
	ctx context.Context,
	balance float64,
	stockPrice float64,
	thread int,
) {

	dateInts := strings.Split(date, "-")

	y, _ := strconv.Atoi(dateInts[0])
	m, _ := strconv.Atoi(dateInts[1])
	d, _ := strconv.Atoi(dateInts[2])

	// insts, _ := ch[0].GetInstrument(ctx, "put", robinhood.NewDate(y, m, d))
	insts, _ := ch[0].GetInstrument(ctx, "call", robinhood.NewDate(y, m, d))

	for _, strikeOpt := range insts {
		is, _ := c.MarketData(strikeOpt)
		market := is[0]

		optPriceUsed := market.AskPrice

		wholeContracts := math.Floor(float64(balance) / (optPriceUsed * 100))
		wholeShares := math.Floor(float64(balance) / stockPrice)

		optBreakEvenUsed := optPriceUsed + strikeOpt.StrikePrice
		costOfStockPosition := wholeShares*stockPrice

		x, y := findIntersection(
			optBreakEvenUsed, 0.0, float64(1000), ((float64(1000)-optBreakEvenUsed)*(wholeContracts*100)),
			stockPrice, 0.0, float64(1000), (float64(1000)*wholeShares) - costOfStockPosition,
		)

		// mel.Broadcast([]byte("msg"))
		mel.Broadcast([]byte(fmt.Sprintf(
			"{\"optBreakEvenUsed\":\"%.2f\", \"stockPrice\":\"%.2f\", \"date\":\"%s\", \"x\":\"%.2f\", \"y\":\"%.2f\", \"optionprice\":\"%.2f\", \"strike\":\"%.2f\",\"contracts\":\"%.2f\", \"shares\":\"%.2f\"}",
			optBreakEvenUsed, stockPrice, date, x, y, optPriceUsed, strikeOpt.StrikePrice, wholeContracts, wholeShares)))
	}
}

func main() {

	// argsWithoutProg := os.Args[1:]
	cfg := fetchConfig()

	o := &robinhood.OAuth{
		Username: cfg.Account.Email,
		Password: cfg.Account.Password,
	}
	c, err := robinhood.Dial(o)
	if err != nil {
		spew.Dump(err)
	}
	tickerSymbol := "EOSE" // argsWithoutProg[0]
	balance := makeFloat("5000")
	quote, _ := c.GetQuote(tickerSymbol)
	stockPrice := quote[0].LastTradePrice
	// spew.Dump(quote)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ch := getChains(c, tickerSymbol)
	fmt.Println(ch[0].ExpirationDates)

	r := gin.Default()
	m := melody.New()

	r.Use(static.Serve("/", static.LocalFile("./public", true)))

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		m.Broadcast(msg)
	})

	for i, date := range ch[0].ExpirationDates {
		go heartBeat(*m, c, date, ch, ctx, balance, stockPrice, i)
	}

	r.Run(":5000")

}

func heartBeat(
	m melody.Melody,
	c *robinhood.Client,
	date string,
	ch []*robinhood.OptionChain,
	ctx context.Context,
	balance float64,
	stockPrice float64,
	thread int,
) {
	for range time.Tick(time.Second * 5) {
		getMarketData(m, c, date, ch, ctx, balance, stockPrice, thread)
	}
}
