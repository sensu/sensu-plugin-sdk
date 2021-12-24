package metric

import (
	"io"
	"math"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	sensuPromHelpTagName = "prom_help"
	sensuPromTypeTagName = "prom_type"
)

// Points
type Points []*corev2.MetricPoint

// ToProm writes Points to a buffer using the prometheus text exposition format
func (m Points) ToProm(writer io.Writer) error {
	metricFamilies := make(map[string]*dto.MetricFamily)
	for _, point := range m {
		var family *dto.MetricFamily
		family, ok := metricFamilies[point.Name]
		if !ok {
			name := point.Name
			var help string
			metricType := dto.MetricType_UNTYPED
			for _, tag := range point.Tags {
				if tag.Name == sensuPromHelpTagName {
					help = tag.Value
					continue
				}
				if tag.Name == sensuPromTypeTagName {
					val := strings.ToLower(tag.Value)
					switch val {
					case "counter":
						metricType = dto.MetricType_COUNTER
					case "gauge":
						metricType = dto.MetricType_GAUGE
					}
				}
			}

			family = &dto.MetricFamily{
				Name: &name,
				Help: &help,
				Type: &metricType,
			}
			metricFamilies[point.Name] = family
		}

		timestampMS := msTimestamp(point.Timestamp)
		metricType := family.GetType()
		value := point.Value
		metric := &dto.Metric{
			TimestampMs: &timestampMS,
		}
		switch metricType {
		case dto.MetricType_COUNTER:
			metric.Counter = &dto.Counter{
				Value: &value,
			}
		case dto.MetricType_GAUGE:
			metric.Gauge = &dto.Gauge{
				Value: &value,
			}
		default:
			metric.Untyped = &dto.Untyped{
				Value: &value,
			}
		}

		for _, tag := range point.Tags {
			if tag.Name == sensuPromHelpTagName || tag.Name == sensuPromTypeTagName {
				continue
			}
			tagName := tag.Name
			tagVal := tag.Value
			metric.Label = append(metric.Label, &dto.LabelPair{Name: &tagName, Value: &tagVal})
		}
		family.Metric = append(family.Metric, metric)
	}

	enc := expfmt.NewEncoder(writer, expfmt.FmtText)
	for _, family := range metricFamilies {
		err := enc.Encode(family)
		if err != nil {
			return err
		}
	}
	return nil
}

// msTimestamp auto-detection of metric point timestamp precision using a heuristic with a 250-ish year cutoff
func msTimestamp(ts int64) int64 {
	timestamp := ts
	switch ts := math.Log10(float64(timestamp)); {
	case ts < 10:
		// assume timestamp is seconds convert to millisecond
		timestamp = time.Unix(timestamp, 0).UnixNano() / int64(time.Millisecond)
	case ts < 13:
		// assume timestamp is milliseconds
	case ts < 16:
		// assume timestamp is microseconds
		timestamp = (timestamp * 1000) / int64(time.Millisecond)
	default:
		// assume timestamp is nanoseconds
		timestamp = timestamp / int64(time.Millisecond)
	}

	return timestamp
}
