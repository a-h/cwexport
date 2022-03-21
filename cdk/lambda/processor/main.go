package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/a-h/cwexport/cw"
	"github.com/a-h/cwexport/db"
	"github.com/a-h/cwexport/firehose"
	"github.com/a-h/cwexport/processor"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"go.uber.org/zap"
)

var log *zap.Logger
var proc processor.Processor

func main() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to load aws config %w", err))
	}

	firehoseName := os.Getenv("METRIC_FIREHOSE_NAME")
	if firehoseName == "" {
		log.Fatal("Missing METRIC_FIREHOSE_NAME env variable")
		return
	}
	fh, err := firehose.New(cfg, firehoseName)
	if err != nil {
		log.Fatal("Cannot create firehose", zap.Error(err))
		return
	}

	if startTimeEnv := os.Getenv("METRIC_START_TIME"); startTimeEnv != "" {
		metricStartTime, err = time.Parse(time.RFC3339, startTimeEnv)
		if err != nil {
			log.Fatal("Unable to parse METRIC_START_TIME, using default time")
			return
		}
	}

	tableName := os.Getenv("METRIC_TABLE_NAME")
	if tableName == "" {
		log.Fatal("Missing METRIC_TABLE_NAME env variable")
		return
	}
	store, err := db.NewMetricStore(tableName, cfg.Region)
	if err != nil {
		log.Fatal("Cannot create store", zap.Error(err))
		return
	}

	proc, err = processor.New(log, store, fh.Put, cw.Cloudwatch{})
	if err != nil {
		log.Error("Failed to create new processor", zap.Error(err))
		return
	}

	lambda.Start(Handle)
}

var metricStartTime = time.Now().Add(time.Minute * -5)

func Handle(ctx context.Context, event types.MetricStat) (err error) {
	log.Info("Received event", zap.Any("event", event))

	err = proc.Process(ctx, metricStartTime, &event)
	if err != nil {
		log.Error("An error occured during processing", zap.Error(err))
		return
	}
	return nil
}
