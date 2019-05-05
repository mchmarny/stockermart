package stockermart

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	apiURL = "https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s"
	apiKey = mustEnvVar("STOCK_PRICER_API_KEY", "")
)


type response struct {
	Quote responseContent `json:"Global Quote"`
}

type responseContent struct {
	Symbol       string  `json:"01. symbol"`
	CurrentPrice float64 `json:"05. price,string"`
	ClosedPrice  float64 `json:"08. previous close,string"`
	ClosedDate   string  `json:"07. latest trading day"`
}

// Quote represents externalize stock quote
type Quote struct {
	Symbol       string    `json:"symbol"`
	CurrentPrice float64   `json:"currentPrice"`
	ClosingPrice float64   `json:"closingPrice"`
	ClosingDate  string    `json:"closingDate"`
	QuotedAt     time.Time `json:"quotedAt"`
}

func getStockPrice(symbol string) (quote *Quote, err error) {

	d, err := getData(symbol)
	if err != nil {
		return nil, err
	}

	q := &Quote{
		Symbol:       d.Symbol,
		CurrentPrice: d.CurrentPrice,
		ClosingPrice: d.ClosedPrice,
		ClosingDate:  d.ClosedDate,
		QuotedAt:     time.Now(),
	}

	return q, nil

}

func getData(symbol string) (data *responseContent, err error) {

	if symbol == "" {
		return nil, errors.New("Nil symbol")
	}

	url := fmt.Sprintf(apiURL, symbol, apiKey)
	logger.Printf("Getting data from: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	logger.Printf("Response status: %s", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Invalid response type")
	}

	var r response
	err = json.NewDecoder(resp.Body).Decode(&r)

	if err != nil {
		return nil, err
	}

	return &r.Quote, nil

}
