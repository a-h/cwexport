package localcmd

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a-h/cwexport/cw"
	"github.com/a-h/cwexport/processor"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func TestOutput(t *testing.T) {
	samples := []processor.MetricSample{
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
	}
	testCases := []struct {
		desc           string
		samples        []processor.MetricSample
		format         Format
		expectedOutput string
	}{

		{
			desc:           "Verify CSV output",
			samples:        samples,
			format:         FormatCSV,
			expectedOutput: "namespace,dimension1/value1,metricsname,Sum,2022-01-01T09:00:00Z,5.000000",
		},
		{
			desc:           "Verify JSON output",
			samples:        samples,
			format:         FormatJSON,
			expectedOutput: `[{"src":"source","Metric":{"Dimensions":[{"Name":"dimension1","Value":"value1"}],"MetricName":"metricsname","Namespace":"namespace"},"Period":null,"Stat":"Sum","Unit":"","sample":{"time":"2022-01-01T09:00:00Z","value":5}}]`,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var w strings.Builder
			var putter processor.MetricPutter
			switch tC.format {
			case FormatCSV:
				putter = newCSVPutter(&w).Put
			case FormatJSON:
				putter = newJSONPutter(&w).Put
			}
			err := putter(context.TODO(), tC.samples)
			if err != nil {
				t.Errorf("Failed to generate output string")
			}
			result := w.String()

			if strings.Compare(tC.expectedOutput, strings.TrimSpace(result)) != 0 {
				t.Errorf("Expected %s, got %s", tC.expectedOutput, result)
			}
		})
	}
}
