package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/a-h/cwexport/cw"
	"github.com/a-h/cwexport/db"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const interval = time.Minute * 5

func GetIntervalCount(startTime time.Time, endTime time.Time) int {
	duration := endTime.Sub(startTime)
	return int(duration / interval)
}

func main() {
	zc := zap.NewProductionConfig()                                         // or zap.NewDevelopmentConfig() or any other zap.Config
	zc.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339) // or time.RubyDate or "2006-01-02 15:04:05" or even freaking time.Kitchen
	logger, _ := zc.Build()
	defer logger.Sync()

	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}
	creds := credentials.NewStaticCredentialsProvider("5cuuni", "yy3dbj", "asdf")
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("localhost"),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		panic(fmt.Errorf("failed to load test aws config %w", err))
	}
	client := dynamodb.NewFromConfig(cfg, dynamodb.WithEndpointResolver(dynamodb.EndpointResolverFromURL(endpoint)))

	store, err := db.NewMetricStore("MetricPosition", "localhost", db.WithClient(client))
	if err != nil {
		logger.Fatal("Cannot create store", zap.Error(err))
		return
	}

	m := cw.Metric{
		Namespace:   "authApi",
		Name:        "challengesStarted",
		ServiceName: "auth-api-challengePostHandler92AD93BF-UH40AniBZd25",
		ServiceType: "AWS::Lambda::Function",
	}

	start := time.Date(2022, time.March, 15, 9, 00, 0, 0, time.UTC)

	lst, ok, err := store.Get(context.Background(), m)
	if err != nil {
		logger.Error("Failed to get last start time from store", zap.Error(err))
		return
	}
	if !ok {
		logger.Info("No start time found...")
	} else {
		logger.Info("Last start time found", zap.Time("startTime", lst))
		start = lst
	}

	ic := GetIntervalCount(start, time.Now())
	if ic > 12 {
		ic = 12
	}
	for i := 0; i < ic; i++ {
		start = start.Add(time.Duration(i) * interval)
		end := start.Add(interval)
		logger := logger.With(
			zap.Time("startTime", start),
			zap.Time("endTime", end),
		)
		logger.Info("Getting metrics for period")
		metrics, err := cw.GetMetrics(m, start, end)
		if err != nil {
			logger.Error("Failed to get metrics for interval", zap.Error(err))
			return
		}
		e := json.NewEncoder(os.Stdout)
		e.SetIndent("", " ")
		err = e.Encode(metrics)
		if err != nil {
			logger.Error("Failed to get serialise metric interval data", zap.Error(err))
			return
		}
		logger.Info("Got metrics for period", zap.Int("metricCount", len(metrics)))
		err = store.Put(context.Background(), m, end)
		if err != nil {
			logger.Error("Failed to save last end time to table", zap.Error(err))
			return
		}
		logger.Info("Successfully processed interval :)")
	}
	logger.Info("Successfully completed all intervals :)", zap.Int("intervalCount", ic))
}
