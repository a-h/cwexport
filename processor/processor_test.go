package processor

import (
	"context"
	"testing"
	"time"

	"github.com/a-h/cwexport/cw"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"go.uber.org/zap"
)

type mockMetricStore struct {
	endTime         time.Time
	lastStartDate   time.Time
	lastStartDateOk bool
}

type mockCloudwatch struct {
	samples []cw.Sample
}

func (store *mockMetricStore) Get(ctx context.Context, m *types.MetricStat) (lastStart time.Time, ok bool, err error) {
	return store.lastStartDate, store.lastStartDateOk, nil
}

func (store *mockMetricStore) Put(ctx context.Context, m *types.MetricStat, endTime time.Time) (err error) {
	store.endTime = endTime
	return
}

func (m *mockCloudwatch) GetSamples(metric *types.MetricStat, start time.Time, end time.Time) (samples []cw.Sample, err error) {
	return m.samples, nil
}

// TestGetIntervalCount tests that correct expectedIntervals are returned when given start and end times
func TestGetIntervalCount(t *testing.T) {
	testCases := []struct {
		desc              string
		startTime         time.Time
		endTime           time.Time
		expectedIntervals int
	}{
		{
			desc:              "When startTime and endTime are equal, then expectedIntervals is 0",
			startTime:         time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			endTime:           time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			expectedIntervals: 0,
		},
		{
			desc:              "When timeRange is less than 5 minutes, there are no intervals",
			startTime:         time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			endTime:           time.Date(2022, time.January, 1, 9, 3, 0, 0, time.UTC),
			expectedIntervals: 0,
		},
		{
			desc:              "When startTime is 10 minutes before endTime, then there are 2 intervals",
			startTime:         time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			endTime:           time.Date(2022, time.January, 1, 9, 10, 0, 0, time.UTC),
			expectedIntervals: 2,
		},
		{
			desc:              "When the timeRange is smaller than the interval time, there are no intervals",
			startTime:         time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			endTime:           time.Date(2022, time.January, 1, 9, 0, 2, 0, time.UTC),
			expectedIntervals: 0,
		},
		{
			desc:              "When endTime is before startTime, then expectedIntervals is 2",
			startTime:         time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			endTime:           time.Date(2022, time.January, 1, 9, 10, 0, 0, time.UTC),
			expectedIntervals: 2,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			actual := getIntervalCount(tC.startTime, tC.endTime)
			if actual != tC.expectedIntervals {
				t.Errorf("expected %d, got %d", tC.expectedIntervals, actual)
			}
		})
	}
}

func TestProcess(t *testing.T) {
	testCases := []struct {
		desc            string
		startTime       time.Time
		lastStartDate   time.Time
		lastStartDateOk bool
		expectedEndtime time.Time
		expectedSamples int
		samples         []cw.Sample
	}{
		{
			desc:            "After processing the end time should match",
			startTime:       time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			samples:         []cw.Sample{},
			expectedEndtime: time.Date(2022, time.January, 1, 10, 0, 0, 0, time.UTC),
			expectedSamples: 0,
			lastStartDate:   time.Date(2022, time.January, 1, 5, 0, 0, 0, time.UTC),
			lastStartDateOk: false,
		},
		{
			desc:            "After processing the last stored start date should be used",
			startTime:       time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			samples:         []cw.Sample{},
			expectedEndtime: time.Date(2022, time.January, 1, 11, 0, 0, 0, time.UTC),
			expectedSamples: 0,
			lastStartDate:   time.Date(2022, time.January, 1, 10, 0, 0, 0, time.UTC),
			lastStartDateOk: true,
		},
		{
			desc:      "After processing we expect one sample to be available",
			startTime: time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
			samples: []cw.Sample{
				{
					Time:  time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC),
					Value: 1,
				},
			},
			expectedEndtime: time.Date(2022, time.January, 1, 10, 0, 0, 0, time.UTC),
			expectedSamples: 1,
			lastStartDate:   time.Date(2022, time.January, 1, 5, 0, 0, 0, time.UTC),
			lastStartDateOk: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			var sample []MetricSample

			metricPutter := func(ctx context.Context, ms []MetricSample) error {
				sample = ms
				return nil
			}
			store := mockMetricStore{
				lastStartDate:   tC.lastStartDate,
				lastStartDateOk: tC.lastStartDateOk,
			}

			testProcessor, _ := New(logger, &store, metricPutter, &mockCloudwatch{samples: tC.samples})
			_ = testProcessor.Process(context.TODO(), tC.startTime, nil)

			if !store.endTime.Equal(tC.expectedEndtime) {
				t.Errorf("Expected end time does not match - got %v expected %v", store.endTime, tC.expectedEndtime)
			}

			if len(sample) != tC.expectedSamples {
				t.Error("Unexpected number of samples")
			}
		})
	}
}
