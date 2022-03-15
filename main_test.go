package main

import (
	"testing"
	"time"
)

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
			actual := GetIntervalCount(tC.startTime, tC.endTime)
			if actual != tC.expectedIntervals {
				t.Errorf("expected %d, got %d", tC.expectedIntervals, actual)
			}
		})
	}
}
