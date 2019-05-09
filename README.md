# stockermart

Stock market price downloader

## Setup

Create pubsub topic

```shell
gcloud beta pubsub topics create stocker-schedule
```


Create BigQuery price table

```shell
bq query --use_legacy_sql=false "
  CREATE OR REPLACE TABLE stocker.price (
    symbol STRING NOT NULL,
    price FLOAT64 NOT NULL,
    closingPrice FLOAT64 NOT NULL,
    closingDate STRING NOT NULL,
    quotedAt TIMESTAMP NOT NULL
)"
```

## Deploy

Create a Cloud Function to download and save the stock prices

```shell
gcloud functions deploy stocker-price \
  --entry-point GetStockPrices \
  --set-env-vars "PID=${GCP_PROJECT},STOCK_PRICER_API_KEY=${STOCK_PRICER_API_KEY}" \
  --memory 512MB \
  --region us-central1 \
  --runtime go112 \
  --trigger-topic stocker-schedule \
  --timeout=540s
```

Then create a Cloud Scheduler job to execute the above function every 30 min by publishing to the `stocker-schedule` topic.

```shell
gcloud beta scheduler jobs create pubsub stocker-price-scheduler \
  --schedule "*/30 * * * *" \
  --topic stocker-schedule \
  --message-body "refresh"
```

## Cleanup

Deleting price function

```shell
gcloud functions delete stocker-price --region us-central1
```

Delete the source topics

```shell
gcloud pubsub topics delete stocker-schedule
```

Delete BigQuery content table

```shell
bq rm stocker.price
```