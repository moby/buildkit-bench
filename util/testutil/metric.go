package testutil

import (
	"testing"
	"time"
)

type MetricUnit string

const (
	MetricBytes MetricUnit = "bytes"

	metricDuration     MetricUnit = "duration"
	metricRefTimestamp MetricUnit = "ref_timestamp"
)

func ReportMetric(b *testing.B, value float64, unit MetricUnit) {
	b.ReportMetric(value, string(unit))
}

func ReportMetricDuration(b *testing.B, value time.Duration) {
	ReportMetric(b, float64(value.Nanoseconds()), metricDuration)
}
