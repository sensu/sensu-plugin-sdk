package metric

import (
	"testing"
	"time"
)

func TestMSTimestamp(t *testing.T) {

	lowerBoundDate, _ := time.Parse(time.RFC3339, "1970-04-27T00:00:00Z")
	upperBoundDate, _ := time.Parse(time.RFC3339, "2262-01-01T00:00:00Z")
	testCases := []struct {
		Unit       string
		LowerBound int64
		UpperBound int64
	}{
		{"Seconds", 0, 1e10 - 1},
		{"Milliseconds", 1e10, 1e13 - 1},
		{"Microseconds", 1e13, 1e16 - 10},
		{"Nanoseconds", 1e16, 1<<63 - 1},
	}

	for _, tc := range testCases {
		t.Run(tc.Unit, func(t *testing.T) {
			lower := msTimestamp(tc.LowerBound)
			upper := msTimestamp(tc.UpperBound)
			lowerDate := time.Unix(lower/1e3, (lower%1e3)*1e6)
			upperDate := time.Unix(upper/1e3, (upper%1e3)*1e6)
			if lowerBoundDate.Before(lowerDate) {
				t.Errorf("Expected date before %v: %v", lowerBoundDate, lowerDate)
			}
			if upperBoundDate.After(upperDate) {
				t.Errorf("Expected date after %v: %v", upperBoundDate, upperDate)
			}
		})
	}
}
