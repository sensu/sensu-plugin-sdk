package metric

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
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
		pointTags := make(map[string]string)
		for _, tag := range point.Tags {
			pointTags[tag.Name] = tag.Value
		}

		var family *dto.MetricFamily
		name := normalizedName(point)
		family, ok := metricFamilies[name]
		if !ok {
			family = createFamily(name, pointTags)
			metricFamilies[name] = family
		}

		timestampMS := msTimestamp(point.Timestamp)
		metricType := family.GetType()
		value := point.Value
		metric := &dto.Metric{
			TimestampMs: &timestampMS,
		}

		// histograms and symmaries are inferred from multiple metric points
		extendsComplexMetric := false

		filteredTags := make(map[string]string)
		for k, v := range pointTags {
			if k == sensuPromHelpTagName || k == sensuPromTypeTagName {
				continue
			}
			if metricType == dto.MetricType_HISTOGRAM && k == model.BucketLabel {
				continue
			}
			if metricType == dto.MetricType_SUMMARY && k == model.QuantileLabel {
				continue
			}
			filteredTags[k] = v
		}

		for k, v := range filteredTags {
			tagName := k
			tagVal := v
			metric.Label = append(metric.Label, &dto.LabelPair{Name: &tagName, Value: &tagVal})
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
		case dto.MetricType_HISTOGRAM:
			// see if metric point already exists
			if match := findMatchingLabels(filteredTags, family.Metric); match != nil {
				extendsComplexMetric = true
				metric = match
			}
			if metric.Histogram == nil {
				metric.Histogram = &dto.Histogram{}
			}
			switch {
			case strings.HasSuffix(point.Name, "_count"):
				sampleCount := uint64(value)
				metric.Histogram.SampleCount = &sampleCount
			case strings.HasSuffix(point.Name, "_sum"):
				metric.Histogram.SampleSum = &value
			case strings.HasSuffix(point.Name, "_bucket"):
				ct := uint64(value)
				val := pointTags[model.BucketLabel]
				le, err := strconv.ParseFloat(val, 64)
				if err != nil {
					return fmt.Errorf("could not map metric point to Prometheus histogram bucket. expected 'le' tag to be a floating point number: %v", err)
				}
				metric.Histogram.Bucket = append(metric.Histogram.Bucket, &dto.Bucket{CumulativeCount: &ct, UpperBound: &le})
			}
		case dto.MetricType_SUMMARY:
			// see if metric point already exists
			if match := findMatchingLabels(filteredTags, family.Metric); match != nil {
				extendsComplexMetric = true
				metric = match
			}
			if metric.Summary == nil {
				metric.Summary = &dto.Summary{}
			}
			switch {
			case strings.HasSuffix(point.Name, "_count"):
				sampleCount := uint64(value)
				metric.Summary.SampleCount = &sampleCount
			case strings.HasSuffix(point.Name, "_sum"):
				metric.Summary.SampleSum = &value
			case family.GetName() == point.Name:
				qLabel := pointTags[model.QuantileLabel]
				quant, err := strconv.ParseFloat(qLabel, 64)
				if err != nil {
					return fmt.Errorf("could not map metric point to Prometheus summary quantile. expected 'quantile' tag to be a floating point number: %v", err)
				}
				metric.Summary.Quantile = append(metric.Summary.Quantile, &dto.Quantile{Quantile: &quant, Value: &value})
			}

		default:
			metric.Untyped = &dto.Untyped{
				Value: &value,
			}
		}

		// this metric point mutated an existing hisogram or summary - skip appending to family
		if extendsComplexMetric {
			continue
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
		timestamp = timestamp * 1e3
	case ts < 13:
		// assume timestamp is milliseconds
	case ts < 16:
		// assume timestamp is microseconds
		timestamp = (timestamp / 1e3)
	default:
		// assume timestamp is nanoseconds
		timestamp = timestamp / int64(time.Millisecond)
	}

	return timestamp
}

// normalizedName for prometheus metric family
func normalizedName(point *corev2.MetricPoint) string {
	switch name := point.Name; {
	case strings.HasSuffix(name, "_bucket"):
		fallthrough
	case strings.HasSuffix(name, "_sum"):
		fallthrough
	case strings.HasSuffix(name, "_count"):
		// only truncate suffix when explicity marked as a prometheus hisogram or summary
		isComplexMetric := false
		for _, t := range point.Tags {
			if t.Name == sensuPromTypeTagName {
				val := strings.ToLower(t.Value)
				if val == "histogram" || val == "summary" {
					isComplexMetric = true
				}
				break
			}
		}
		if !isComplexMetric {
			return name
		}
		parts := strings.Split(name, "_")
		return strings.Join(parts[:len(parts)-1], "_")
	default:
		return name
	}
}

func createFamily(name string, tags map[string]string) *dto.MetricFamily {
	var help string
	metricType := dto.MetricType_UNTYPED
	if val, ok := tags[sensuPromHelpTagName]; ok {
		help = val
	}
	if val, ok := tags[sensuPromTypeTagName]; ok {
		val = strings.ToLower(val)
		switch val {
		case "counter":
			metricType = dto.MetricType_COUNTER
		case "gauge":
			metricType = dto.MetricType_GAUGE
		case "histogram":
			metricType = dto.MetricType_HISTOGRAM
		case "summary":
			metricType = dto.MetricType_SUMMARY
		}
	}
	return &dto.MetricFamily{
		Name: &name,
		Help: &help,
		Type: &metricType,
	}
}

// findMatchingLabels returns the first metric where all labels match the provided tags
func findMatchingLabels(tags map[string]string, metrics []*dto.Metric) *dto.Metric {
	for _, match := range metrics {
		if len(tags) != len(match.Label) {
			continue
		}
		hasMatch := true
		for _, tag := range match.Label {
			v, ok := tags[tag.GetName()]
			if !ok || v != tag.GetValue() {
				hasMatch = false
				break
			}
		}
		if hasMatch {
			return match
		}
	}
	return nil
}
