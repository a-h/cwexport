package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	cw := cloudwatch.NewFromConfig(cfg)
	md, err := cw.GetMetricData(ctx, &cloudwatch.GetMetricDataInput{
		StartTime: aws.Time(time.Date(2022, time.March, 14, 11, 40, 0, 0, time.UTC)),
		EndTime:   aws.Time(time.Date(2022, time.March, 14, 14, 00, 0, 0, time.UTC)),
		MetricDataQueries: []types.MetricDataQuery{
			{
				Id: aws.String("a"),
				MetricStat: &types.MetricStat{
					Metric: &types.Metric{
						MetricName: aws.String("challengesStarted"),
						Namespace:  aws.String("authApi"),
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
					Period: aws.Int32(60 * 5), // Seconds
					Stat:   aws.String("Sum"),
				},
				ReturnData: aws.Bool(true),
			},
		},
		ScanBy: types.ScanByTimestampAscending,
	})
	if err != nil {
		log.Fatalf("failed to get metrics: %v", err)
	}
	for _, m := range md.MetricDataResults {
		for i := 0; i < len(m.Timestamps); i++ {
			fmt.Printf("%v\t%v\n", m.Timestamps[i], m.Values[i])
		}
	}
}
