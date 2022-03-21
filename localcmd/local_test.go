package localcmd

import (
	"context"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/a-h/cwexport/cw"
	"github.com/a-h/cwexport/processor"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func TestCSVOutput(t *testing.T) {
	testCases := []struct {
		desc           string
		samples        []processor.MetricSample
		expectedOutput string
	}{
		{
			desc: "Verify CSV output",
			samples: []processor.MetricSample{
				{
					Source: "source",
					MetricStat: &types.MetricStat{
						Metric: &types.Metric{
							MetricName: aws.String("metricsname"),
							Namespace:  aws.String("namespace"),
							Dimensions: []types.Dimension{
								{
									Name:  aws.String("dimension1"),
									Value: aws.String("value1"),
								},
							},
						},
						Stat: aws.String("Sum"),
					},
					Sample: cw.Sample{
						Time:  time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
						Value: 5,
					},
				},
			},
			expectedOutput: "namespace,dimension1/value1,metricsname,Sum,2022-01-01T09:00:00Z,5.000000\n",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var w strings.Builder
			cw := csv.NewWriter(&w)
			cp := csvPutter{writer: *cw}
			err := cp.Put(context.TODO(), tC.samples)
			cw.Flush()
			if err != nil {
				t.Errorf("Failed to generate csv")
			}

			result := w.String()

			if strings.Compare(tC.expectedOutput, result) != 0 {
				t.Errorf("Expected %s, got %s", tC.expectedOutput, result)
			}
		})
	}
}
