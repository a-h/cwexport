package db

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cw "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

func TestMetricStore(t *testing.T) {
	if testing.Short() {
		return
	}
	tableName := createLocalTable(t)
	defer deleteLocalTable(t, tableName)

	ms, err := NewMetricStore(tableName, region, WithClient(testClient))
	if err != nil {
		t.Fatalf("cannot create metric store: %v", err)
	}
	ctx := context.Background()
	t.Run("it cannot get a start time for a metric that doesn't exist", func(t *testing.T) {
		_, ok, err := ms.Get(ctx, &cw.MetricStat{
			Metric: &cw.Metric{
				Namespace:  aws.String("emptyNamespace"),
				MetricName: aws.String("missingMetric"),
				Dimensions: []cw.Dimension{
					{
						Name:  aws.String("ServiceName"),
						Value: aws.String("sn"),
					},
					{
						Name:  aws.String("ServiceType"),
						Value: aws.String("st"),
					},
				},
			},
			Period: aws.Int32(1),
			Stat:   aws.String("Sum"),
		})
		if err != nil {
			t.Fatalf("unexpected error getting metric: %v", err)
		}
		if ok {
			t.Fatal("expected ok=false, got ok=true")
		}
	})
	lastStart := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	t.Run("it can insert a metric", func(t *testing.T) {
		err := ms.Put(ctx, &cw.MetricStat{
			Metric: &cw.Metric{
				Namespace:  aws.String("ns"),
				MetricName: aws.String("metricA"),
				Dimensions: []cw.Dimension{
					{
						Name:  aws.String("ServiceName"),
						Value: aws.String("sn"),
					},
					{
						Name:  aws.String("ServiceType"),
						Value: aws.String("st"),
					},
				},
			},
			Period: aws.Int32(1),
			Stat:   aws.String("Sum"),
		}, lastStart)
		if err != nil {
			t.Fatalf("unexpected error putting metric: %v", err)
		}
	})
	t.Run("it can get a start time for a previously inserted metric", func(t *testing.T) {
		actualLastStart, ok, err := ms.Get(ctx, &cw.MetricStat{
			Metric: &cw.Metric{
				Namespace:  aws.String("ns"),
				MetricName: aws.String("metricA"),
				Dimensions: []cw.Dimension{
					{
						Name:  aws.String("ServiceName"),
						Value: aws.String("sn"),
					},
					{
						Name:  aws.String("ServiceType"),
						Value: aws.String("st"),
					},
				},
			},
			Period: aws.Int32(1),
			Stat:   aws.String("Sum"),
		})
		if err != nil {
			t.Fatalf("unexpected error getting metric: %v", err)
		}
		if !ok {
			t.Fatalf("expected ok=true, got ok=false")
		}
		if !actualLastStart.Equal(lastStart) {
			t.Fatalf("expected last start to be the one we previously set, but got %v", actualLastStart)
		}
	})
	lastStart = time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC)
	t.Run("it can update a previously inserted metric", func(t *testing.T) {
		err := ms.Put(ctx, &cw.MetricStat{
			Metric: &cw.Metric{
				Namespace:  aws.String("ns"),
				MetricName: aws.String("metricA"),
				Dimensions: []cw.Dimension{
					{
						Name:  aws.String("ServiceName"),
						Value: aws.String("sn"),
					},
					{
						Name:  aws.String("ServiceType"),
						Value: aws.String("st"),
					},
				},
			},
			Period: aws.Int32(1),
			Stat:   aws.String("Sum"),
		}, lastStart)
		if err != nil {
			t.Fatalf("unexpected error putting metric: %v", err)
		}
	})
	t.Run("it can get the start time for an updated metric", func(t *testing.T) {
		actualLastStart, ok, err := ms.Get(ctx, &cw.MetricStat{
			Metric: &cw.Metric{
				Namespace:  aws.String("ns"),
				MetricName: aws.String("metricA"),
				Dimensions: []cw.Dimension{
					{
						Name:  aws.String("ServiceName"),
						Value: aws.String("sn"),
					},
					{
						Name:  aws.String("ServiceType"),
						Value: aws.String("st"),
					},
				},
			},
			Period: aws.Int32(1),
			Stat:   aws.String("Sum"),
		})
		if err != nil {
			t.Fatalf("unexpected error getting metric: %v", err)
		}
		if !ok {
			t.Fatalf("expected ok=true, got ok=false")
		}
		if !actualLastStart.Equal(lastStart) {
			t.Fatalf("expected last start to be the one we just updated, but got %v", actualLastStart)
		}
	})
}
