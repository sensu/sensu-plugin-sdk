package metric_test

import (
	"bytes"
	"errors"
	"io"
	"math"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu/metric"
	"github.com/stretchr/testify/assert"
)

// TestToPromMetric tests attributes of single metric points from prom exposition
func TestToPromMetric(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339Nano, "2022-01-01T15:04:05.25Z")
	tsUnixNano := ts.UnixNano()
	tsUnixMicro := int64(tsUnixNano / 1e3)
	tsUnixMilli := int64(tsUnixNano / 1e6)
	tsUnix := ts.Unix()
	testCases := []struct {
		Name     string
		Metric   []*corev2.MetricPoint
		Expected dto.MetricFamily
	}{
		{
			Name: "Basic Metric",
			Metric: []*corev2.MetricPoint{{
				Name:      "metric_point",
				Value:     22.234,
				Timestamp: tsUnixNano,
			}},
			Expected: dto.MetricFamily{
				Name:   sptr("metric_point"),
				Type:   dto.MetricType_UNTYPED.Enum(),
				Metric: []*dto.Metric{{Untyped: &dto.Untyped{Value: fptr(22.234)}, TimestampMs: iptr(tsUnixMilli)}},
			},
		}, {
			Name: "Counter With Help Info",
			Metric: []*corev2.MetricPoint{{
				Name:      "gc_cycles",
				Value:     20.0,
				Timestamp: tsUnixMilli,
				Tags: []*corev2.MetricTag{
					{Name: "prom_type", Value: "counter"},
					{Name: "prom_help", Value: "halp"},
				},
			}},
			Expected: dto.MetricFamily{
				Name:   sptr("gc_cycles"),
				Type:   dto.MetricType_COUNTER.Enum(),
				Help:   sptr("halp"),
				Metric: []*dto.Metric{{Counter: &dto.Counter{Value: fptr(20.0)}, TimestampMs: iptr(tsUnixMilli)}},
			},
		}, {
			Name: "Gauge With Extra Tags",
			Metric: []*corev2.MetricPoint{{
				Name:      "goroutines",
				Value:     20.0,
				Timestamp: tsUnix,
				Tags: []*corev2.MetricTag{
					{Name: "prom_type", Value: "Gauge"},
					{Name: "fizz", Value: "buzz"},
				},
			}},
			Expected: dto.MetricFamily{
				Name: sptr("goroutines"),
				Type: dto.MetricType_GAUGE.Enum(),
				Metric: []*dto.Metric{
					{
						Label:       []*dto.LabelPair{{Name: sptr("fizz"), Value: sptr("buzz")}},
						Gauge:       &dto.Gauge{Value: fptr(20.0)},
						TimestampMs: iptr(tsUnixMilli - (tsUnixMilli % 1000)), // drop 3 lsd
					},
				},
			},
		}, {
			Name: "Metric with Microsecond Timestamp",
			Metric: []*corev2.MetricPoint{{
				Name:      "gc_cycles",
				Value:     20.0,
				Timestamp: tsUnixMicro,
				Tags: []*corev2.MetricTag{
					{Name: "prom_type", Value: "counter"},
					{Name: "prom_help", Value: "halp"},
				},
			}},
			Expected: dto.MetricFamily{
				Name:   sptr("gc_cycles"),
				Type:   dto.MetricType_COUNTER.Enum(),
				Help:   sptr("halp"),
				Metric: []*dto.Metric{{Counter: &dto.Counter{Value: fptr(20.0)}, TimestampMs: iptr(tsUnixMilli)}},
			},
		}, {
			Name: "Histogram",
			Metric: []*corev2.MetricPoint{
				{
					Name:      "http_request_duration_seconds_bucket",
					Value:     1000,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "histogram"},
						{Name: "prom_help", Value: "distribution of http request duration"},
						{Name: "le", Value: "0.1"},
					},
				}, {
					Name:      "http_request_duration_seconds_bucket",
					Value:     2000,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "histogram"},
						{Name: "prom_help", Value: "distribution of http request duration"},
						{Name: "le", Value: "0.5"},
					},
				}, {
					Name:      "http_request_duration_seconds_bucket",
					Value:     3000,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "histogram"},
						{Name: "prom_help", Value: "distribution of http request duration"},
						{Name: "le", Value: "+Inf"},
					},
				}, {
					Name:      "http_request_duration_seconds_count",
					Value:     6000,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "histogram"},
						{Name: "prom_help", Value: "distribution of http request duration"},
					},
				}, {
					Name:      "http_request_duration_seconds_sum",
					Value:     54321.01,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "histogram"},
						{Name: "prom_help", Value: "distribution of http request duration"},
					},
				},
			},
			Expected: dto.MetricFamily{
				Name: sptr("http_request_duration_seconds"),
				Type: dto.MetricType_HISTOGRAM.Enum(),
				Help: sptr("distribution of http request duration"),
				Metric: []*dto.Metric{
					{
						Histogram: &dto.Histogram{
							SampleCount: uiptr(6000),
							SampleSum:   fptr(54321.01),
							Bucket: []*dto.Bucket{
								{CumulativeCount: uiptr(1000), UpperBound: fptr(0.1)},
								{CumulativeCount: uiptr(2000), UpperBound: fptr(0.5)},
								{CumulativeCount: uiptr(3000), UpperBound: fptr(math.Inf(1))},
							},
						},
						TimestampMs: iptr(tsUnixMilli - (tsUnixMilli % 1000)), // drop 3 lsd
					},
				},
			},
		}, {
			Name: "Summary",
			Metric: []*corev2.MetricPoint{
				{
					Name:      "telemetry_requests_metrics_latency_microseconds_sum",
					Value:     1.7560473e7,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "summary"},
						{Name: "prom_help", Value: "A summary of the response latency."},
					},
				}, {
					Name:      "telemetry_requests_metrics_latency_microseconds_count",
					Value:     2693,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "summary"},
						{Name: "prom_help", Value: "A summary of the response latency."},
					},
				}, {
					Name:      "telemetry_requests_metrics_latency_microseconds",
					Value:     3102,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "summary"},
						{Name: "prom_help", Value: "A summary of the response latency."},
						{Name: "quantile", Value: "0.01"},
					},
				}, {
					Name:      "telemetry_requests_metrics_latency_microseconds",
					Value:     3272,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "summary"},
						{Name: "prom_help", Value: "A summary of the response latency."},
						{Name: "quantile", Value: "0.05"},
					},
				}, {
					Name:      "telemetry_requests_metrics_latency_microseconds",
					Value:     4773,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "summary"},
						{Name: "prom_help", Value: "A summary of the response latency."},
						{Name: "quantile", Value: "0.5"},
					},
				}, {
					Name:      "telemetry_requests_metrics_latency_microseconds",
					Value:     9001,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "summary"},
						{Name: "prom_help", Value: "A summary of the response latency."},
						{Name: "quantile", Value: "0.9"},
					},
				}, {
					Name:      "telemetry_requests_metrics_latency_microseconds",
					Value:     76656,
					Timestamp: tsUnix,
					Tags: []*corev2.MetricTag{
						{Name: "prom_type", Value: "summary"},
						{Name: "prom_help", Value: "A summary of the response latency."},
						{Name: "quantile", Value: "0.99"},
					},
				},
			},
			Expected: dto.MetricFamily{
				Name: sptr("telemetry_requests_metrics_latency_microseconds"),
				Type: dto.MetricType_SUMMARY.Enum(),
				Help: sptr("A summary of the response latency."),
				Metric: []*dto.Metric{
					{
						TimestampMs: iptr(tsUnixMilli - (tsUnixMilli % 1000)), // drop 3 lsd
						Summary: &dto.Summary{
							SampleCount: uiptr(2693),
							SampleSum:   fptr(1.7560473e7),
							Quantile: []*dto.Quantile{
								{Quantile: fptr(0.01), Value: fptr(3102)},
								{Quantile: fptr(0.05), Value: fptr(3272)},
								{Quantile: fptr(0.5), Value: fptr(4773)},
								{Quantile: fptr(0.9), Value: fptr(9001)},
								{Quantile: fptr(0.99), Value: fptr(76656)},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buf bytes.Buffer
			underTest := metric.Points(tc.Metric)
			underTest.ToProm(&buf)

			var family dto.MetricFamily
			err := expfmt.NewDecoder(&buf, expfmt.FmtText).Decode(&family)
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, family)

		})
	}
}

// TestToPromFamily more broadly tests ToProm for cases with multiple metric points and families
func TestToPromFamily(t *testing.T) {
	inpt := []*corev2.MetricPoint{
		{
			Name:      "test1",
			Value:     1.23,
			Timestamp: 1e9,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "gauge"},
				{Name: "host", Value: "web-01"},
			},
		}, {
			Name:      "test2",
			Value:     2.34,
			Timestamp: 1e9 + 1,
			Tags: []*corev2.MetricTag{
				{Name: "host", Value: "web-01"},
			},
		}, {
			Name:      "foo",
			Value:     1.0,
			Timestamp: 200000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "counter"},
				{Name: "prom_help", Value: "foo is a foo counter"},
				{Name: "host", Value: "web-01"},
			},
		}, {
			Name:      "test2",
			Value:     3.45,
			Timestamp: 1e9 + 1,
			Tags: []*corev2.MetricTag{
				{Name: "host", Value: "web-02"},
			},
		}, {
			Name:      "foo",
			Value:     2.2,
			Timestamp: 200000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "counter"},
				{Name: "prom_help", Value: "foo is a foo counter"},
				{Name: "host", Value: "web-02"},
			},
		}, {
			Name:      "http_request_duration_seconds_bucket",
			Value:     1000,
			Timestamp: 200000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-01"},
				{Name: "le", Value: "0.1"},
			},
		}, {
			Name:      "http_request_duration_seconds_bucket",
			Value:     2000,
			Timestamp: 2000000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-01"},
				{Name: "le", Value: "0.5"},
			},
		}, {
			Name:      "http_request_duration_seconds_bucket",
			Value:     3000,
			Timestamp: 2000000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-01"},
				{Name: "le", Value: "+Inf"},
			},
		}, {
			Name:      "http_request_duration_seconds_count",
			Value:     6000,
			Timestamp: 2000000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-01"},
			},
		}, {
			Name:      "http_request_duration_seconds_sum",
			Value:     54321.01,
			Timestamp: 2000000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-01"},
			},
		}, {
			Name:      "http_request_duration_seconds_bucket",
			Value:     1000,
			Timestamp: 200000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-02"},
				{Name: "le", Value: "0.1"},
			},
		}, {
			Name:      "http_request_duration_seconds_bucket",
			Value:     2000,
			Timestamp: 2000000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-02"},
				{Name: "le", Value: "0.5"},
			},
		}, {
			Name:      "http_request_duration_seconds_bucket",
			Value:     3000,
			Timestamp: 2000000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-02"},
				{Name: "le", Value: "+Inf"},
			},
		}, {
			Name:      "http_request_duration_seconds_count",
			Value:     6000,
			Timestamp: 2000000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-02"},
			},
		}, {
			Name:      "http_request_duration_seconds_sum",
			Value:     54321.01,
			Timestamp: 2000000,
			Tags: []*corev2.MetricTag{
				{Name: "prom_type", Value: "histogram"},
				{Name: "prom_help", Value: "distribution of http request duration"},
				{Name: "host", Value: "web-02"},
			},
		},
	}
	var actualOutput bytes.Buffer
	underTest := metric.Points(inpt)
	underTest.ToProm(&actualOutput)
	decoder := expfmt.NewDecoder(&actualOutput, expfmt.FmtText)

	actualFamilies := make(map[string]*dto.MetricFamily)
	for {
		var family dto.MetricFamily
		err := decoder.Decode(&family)
		if errors.Is(err, io.EOF) {
			break
		}
		assert.NoError(t, err)
		actualFamilies[family.GetName()] = &family
	}

	assert.Equal(t, 4, len(actualFamilies))

	actualFoo := actualFamilies["foo"]
	actualTest1 := actualFamilies["test1"]
	actualTest2 := actualFamilies["test2"]
	actualHttpReqDuration := actualFamilies["http_request_duration_seconds"]

	assert.Equal(t, "foo is a foo counter", actualFoo.GetHelp())
	assert.Equal(t, "", actualTest1.GetHelp())
	assert.Equal(t, "", actualTest2.GetHelp())
	assert.Equal(t, "distribution of http request duration", actualHttpReqDuration.GetHelp())

	assert.Equal(t, dto.MetricType_COUNTER.Enum(), actualFoo.Type)
	assert.Equal(t, dto.MetricType_GAUGE.Enum(), actualTest1.Type)
	assert.Equal(t, dto.MetricType_UNTYPED.Enum(), actualTest2.Type)
	assert.Equal(t, dto.MetricType_HISTOGRAM.Enum(), actualHttpReqDuration.Type)

	assert.Equal(t, 2, len(actualFoo.Metric))
	assert.Equal(t, 1, len(actualTest1.Metric))
	assert.Equal(t, 2, len(actualTest2.Metric))
	assert.Equal(t, 2, len(actualHttpReqDuration.Metric))
}

func sptr(wrapped string) *string {
	return &wrapped
}
func fptr(wrapped float64) *float64 {
	return &wrapped
}
func iptr(wrapped int64) *int64 {
	return &wrapped
}
func uiptr(wrapped uint64) *uint64 {
	return &wrapped
}
