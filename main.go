package main

import (
	"context"
	"fmt"
	"time"

	"github.com/a-h/cwexport/db"
	"github.com/a-h/cwexport/firehose"
	"github.com/a-h/cwexport/processor"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	zc := zap.NewProductionConfig()
	zc.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logger, _ := zc.Build()
	defer logger.Sync()

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to load test aws config %w", err))
	}

	fh, err := firehose.New(cfg, "cwexport-MetricDeliveryStreamA287D02E-oK0s7aoP5Bi1")
	if err != nil {
		logger.Fatal("Cannot create firehose", zap.Error(err))
		return
	}

	store, err := db.NewMetricStore("cwexport-CWExportMetricTable116C3288-14FQXYEUTR25R", cfg.Region)
	if err != nil {
		logger.Fatal("Cannot create store", zap.Error(err))
		return
	}

	m := &types.MetricStat{
		Metric: &types.Metric{
			Namespace:  aws.String("authApi"),
			MetricName: aws.String("challengesStarted"),
			Dimensions: []types.Dimension{
				{
					Name:  aws.String("ServiceName"),
					Value: aws.String("auth-api-challengePostHandler92AD93BF-UH40AniBZd25"),
				},
				{
					Name:  aws.String("ServiceType"),
					Value: aws.String("AWS::Lambda::Function"),
				},
			},
		},
		Period: aws.Int32(5),
		Stat:   aws.String("Sum"),
	}

	start := time.Date(2022, time.March, 15, 9, 00, 0, 0, time.UTC)

	p, err := processor.New(logger, fh, store)
	if err != nil {
		logger.Error("Failed to create new processor", zap.Error(err))
		return
	}

	err = p.Process(ctx, start, m)
	if err != nil {
		logger.Error("An error occured during processing", zap.Error(err))
		return
	}
}
