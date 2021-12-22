package metric_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-plugin-sdk/sensu/metric"
	"github.com/stretchr/testify/assert"
)

// TestToPromMetric tests attributes of single metric points from prom exposition
func TestToPromMetric(t *testing.T) {
	testCases := []struct {
		Name     string
		Metric   *corev2.MetricPoint
		Expected dto.MetricFamily
	}{
		{
			Name: "Basic Metric",
			Metric: &corev2.MetricPoint{
				Name:      "metric_point",
				Value:     22.234,
				Timestamp: 1e8,
			},
			Expected: dto.MetricFamily{
				Name:   sptr("metric_point"),
				Type:   dto.MetricType_UNTYPED.Enum(),
				Metric: []*dto.Metric{{Untyped: &dto.Untyped{Value: fptr(22.234)}, TimestampMs: iptr(1e8)}},
			},
		}, {
			Name: "Counter With Help Info",
			Metric: &corev2.MetricPoint{
				Name:      "gc_cycles",
				Value:     20.0,
				Timestamp: 1e8,
				Tags: []*corev2.MetricTag{
					{Name: "prom_type", Value: "counter"},
					{Name: "prom_help", Value: "halp"},
				},
			},
			Expected: dto.MetricFamily{
				Name:   sptr("gc_cycles"),
				Type:   dto.MetricType_COUNTER.Enum(),
				Help:   sptr("halp"),
				Metric: []*dto.Metric{{Counter: &dto.Counter{Value: fptr(20.0)}, TimestampMs: iptr(1e8)}},
			},
		}, {
			Name: "Gauge With Extra Tags",
			Metric: &corev2.MetricPoint{
				Name:      "goroutines",
				Value:     20.0,
				Timestamp: 1e8,
				Tags: []*corev2.MetricTag{
					{Name: "prom_type", Value: "Gauge"},
					{Name: "fizz", Value: "buzz"},
				},
			},
			Expected: dto.MetricFamily{
				Name: sptr("goroutines"),
				Type: dto.MetricType_GAUGE.Enum(),
				Metric: []*dto.Metric{
					{
						Label:       []*dto.LabelPair{{Name: sptr("fizz"), Value: sptr("buzz")}},
						Gauge:       &dto.Gauge{Value: fptr(20.0)},
						TimestampMs: iptr(1e8),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var buf bytes.Buffer
			underTest := metric.Points{tc.Metric}
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

	assert.Equal(t, 3, len(actualFamilies))

	actualFoo := actualFamilies["foo"]
	actualTest1 := actualFamilies["test1"]
	actualTest2 := actualFamilies["test2"]

	assert.Equal(t, "foo is a foo counter", actualFoo.GetHelp())
	assert.Equal(t, "", actualTest1.GetHelp())
	assert.Equal(t, "", actualTest2.GetHelp())

	assert.Equal(t, dto.MetricType_COUNTER.Enum(), actualFoo.Type)
	assert.Equal(t, dto.MetricType_GAUGE.Enum(), actualTest1.Type)
	assert.Equal(t, dto.MetricType_UNTYPED.Enum(), actualTest2.Type)

	assert.Equal(t, 2, len(actualFoo.Metric))
	assert.Equal(t, 1, len(actualTest1.Metric))
	assert.Equal(t, 2, len(actualTest2.Metric))
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
