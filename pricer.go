package stockermart

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	bq "cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

var (
	logger    = log.New(os.Stdout, "[PRICE] ", 0)
	projectID = mustEnvVar("PID", "")
	bqDataSet = mustEnvVar("BQ_DATSET", "stocker")

	once      sync.Once
	bqClient  *bq.Client
	companies []string
)

// PubSubMessage is the payload of a Pub/Sub event
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// GetStockPrices processes pubsub topic events
func GetStockPrices(ctx context.Context, m PubSubMessage) error {

	once.Do(func() {

		// bigquery
		bqc, err := bq.NewClient(ctx, projectID)
		if err != nil {
			logger.Fatalf("Error creating BQ client: %v", err)
		}
		bqClient = bqc

		cos, err := getCompanies(ctx)
		if err != nil {
			logger.Fatalf("Error getting companies: %v", err)
		}
		companies = cos
	})

	// TODO: this won't work for larger number of companies
	quotes := make([]*Quote, 0)
	for _, co := range companies {

		quote, err := getStockPrice(co)
		if err != nil {
			logger.Printf("Error pricing symbol: %v", err)
		}
		quotes = append(quotes, quote)

	}

	return saveQuotes(ctx, quotes)

}

func saveQuotes(ctx context.Context, quotes []*Quote) error {

	logger.Println("Getting companies...")

	u := bqClient.Dataset(bqDataSet).Table("price").Uploader()
	if err := u.Put(ctx, quotes); err != nil {
		logger.Printf("Error inserting quotes: %v", err)
		return err
	}

	return nil

}

func getCompanies(ctx context.Context) (symbols []string, err error) {

	logger.Println("Getting companies...")

	qSQL := fmt.Sprintf("SELECT symbol FROM %s.company", bqDataSet)
	q := bqClient.Query(qSQL)
	it, err := q.Read(ctx)
	if err != nil {
		logger.Printf("Error quering BQ: %v", err)
		return nil, err
	}

	list := make([]string, 0)

	for {
		var c string
		err := it.Next(&c)
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Printf("Error looping through BQ values: %v", err)
			return nil, err
		}
		list = append(list, c)
	}

	logger.Printf("Found %d companies", len(list))
	return list, nil

}

func mustEnvVar(key, fallbackValue string) string {

	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	if fallbackValue == "" {
		logger.Fatalf("Required envvar not set: %s", key)
	}

	return fallbackValue
}
