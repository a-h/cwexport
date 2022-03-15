package main

import (
	"time"

	"github.com/a-h/cwexport/db"
)

const interval = time.Minute * 5

func GetIntervalCount(startTime time.Time, endTime time.Time) int {
	duration := endTime.Sub(startTime)
	return int(duration / interval)
}

func main() {
	db.DescribeTable()

	/*
		config := zap.NewProductionConfig()                                         // or zap.NewDevelopmentConfig() or any other zap.Config
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339) // or time.RubyDate or "2006-01-02 15:04:05" or even freaking time.Kitchen

		logger, _ := config.Build()

		defer logger.Sync()

		m := cw.Metric{
			Namespace:   "authApi",
			Name:        "challengesStarted",
			ServiceName: "auth-api-challengePostHandler92AD93BF-UH40AniBZd25",
			ServiceType: "AWS::Lambda::Function",
			StartTime:   time.Date(2022, time.March, 15, 9, 00, 0, 0, time.UTC),
			EndTime:     time.Date(2022, time.March, 15, 10, 30, 0, 0, time.UTC),
		}

		ic := GetIntervalCount(m.StartTime, m.EndTime)
		for i := 0; i < ic; i++ {
			m.StartTime = m.StartTime.Add(time.Duration(i) * interval)
			m.EndTime = m.StartTime.Add(interval)
			logger := logger.With(
				zap.Any("startTime", m.StartTime),
				zap.Any("endTime", m.EndTime),
			)
			logger.Info("Getting metrics for period")
			metrics, err := cw.GetMetrics(m)
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
		}
	*/
}
