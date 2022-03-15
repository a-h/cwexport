package cw

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type Metric struct {
	Namespace   string `json:"ns"`
	Name        string `json:"name"`
	ServiceName string `json:"serviceName"`
	ServiceType string `json:"serviceType"`
}

type Sample struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

func GetMetrics(metric Metric, start time.Time, end time.Time) (samples []Sample, err error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		err = fmt.Errorf("unable to load SDK config: %w", err)
		return
	}

	cw := cloudwatch.NewFromConfig(cfg)
	params := &cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(start),
		EndTime:   aws.Time(end),
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("a"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						MetricName: &metric.Name,
						Namespace:  &metric.Namespace,
						Dimensions: []types.Dimension{
							{
								Name:  aws.String("ServiceName"),
								Value: &metric.ServiceName,
							},
							{
								Name:  aws.String("ServiceType"),
								Value: &metric.ServiceType,
							},
						},
					},
					Period: aws.Int32(60 * 5), // Seconds
					Stat:   aws.String("Sum"),
				},
				ReturnData: aws.Bool(true),
			},
		},
		ScanBy: types.ScanByTimestampAscending,
	}

	paginator := cloudwatch.NewGetMetricDataPaginator(cw, params)
	for paginator.HasMorePages() {
		var md *cloudwatch.GetMetricDataOutput
		md, err = paginator.NextPage(ctx)
		if err != nil {
			err = fmt.Errorf("failed to get metrics: %w", err)
			return
		}
		for _, m := range md.MetricDataResults {
			for i := 0; i < len(m.Timestamps); i++ {
				samples = append(samples, Sample{
					Time:  m.Timestamps[i],
					Value: m.Values[i],
				})
			}
		}
	}
	return
}
