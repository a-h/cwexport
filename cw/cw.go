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

type Sample struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

func GetSamples(metric *types.MetricStat, start time.Time, end time.Time) (samples []Sample, err error) {
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
				Id:         aws.String("a"),
				MetricStat: metric,
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
