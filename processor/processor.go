package processor

import (
	"context"
	"time"

	"github.com/a-h/cwexport/cw"
	"github.com/a-h/cwexport/db"
	"github.com/a-h/cwexport/firehose"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"go.uber.org/zap"
)

const interval = time.Minute * 5

type Processor struct {
	logger   *zap.Logger
	firehose firehose.Firehose
	store    *db.MetricStore
}

type MetricSample struct {
	*types.MetricStat
	cw.Sample `json:"sample"`
}

// New creates a new Process with all required fields populated.
func New(logger *zap.Logger, firehose firehose.Firehose, store *db.MetricStore) (Processor, error) {
	return Processor{
		logger:   logger,
		firehose: firehose,
		store:    store,
	}, nil
}

func GetIntervalCount(startTime time.Time, endTime time.Time) int {
	duration := endTime.Sub(startTime)
	return int(duration / interval)
}

func (p Processor) Process(ctx context.Context, start time.Time, metric *types.MetricStat) error {
	lst, ok, err := p.store.Get(ctx, metric)
	if err != nil {
		p.logger.Error("Failed to get last start time from store", zap.Error(err))
		return err
	}
	if !ok {
		p.logger.Info("No start time found...")
	} else {
		p.logger.Info("Last start time found", zap.Time("startTime", lst))
		start = lst
	}

	ic := GetIntervalCount(start, time.Now())
	if ic > 12 {
		ic = 12
	}
	for i := 0; i < ic; i++ {
		start = start.Add(time.Duration(i) * interval)
		end := start.Add(interval)
		logger := p.logger.With(
			zap.Time("startTime", start),
			zap.Time("endTime", end),
		)
		logger.Info("Getting metrics for period")
		samples, err := cw.GetSamples(metric, start, end)
		if err != nil {
			logger.Error("Failed to get metrics for interval", zap.Error(err))
			return err
		}
		logger.Info("Got metrics for period", zap.Int("metricCount", len(samples)))

		logger.Info("Sending metrics to Firehose")
		var metricSamples []interface{}
		for _, s := range samples {
			metricSamples = append(metricSamples, MetricSample{
				MetricStat: metric,
				Sample:     s,
			})
		}

		err = p.firehose.Put(ctx, metricSamples)
		if err != nil {
			logger.Error("Failed to send data to firehose", zap.Error(err))
			return err
		}

		logger.Info("Saving the last runtime in the database")
		err = p.store.Put(ctx, metric, end)
		if err != nil {
			logger.Error("Failed to save last end time to table", zap.Error(err))
			return err
		}
		logger.Info("Successfully processed interval :)")
	}
	p.logger.Info("Successfully completed all intervals :)", zap.Int("intervalCount", ic))
	return nil
}
