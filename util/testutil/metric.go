package testutil

import (
	"testing"
	"time"
)

type MetricUnit string

const (
	MetricBytes    MetricUnit = "bytes"
	MetricDuration MetricUnit = "duration"
)

func ReportMetric(b *testing.B, value float64, unit MetricUnit) {
	b.ReportMetric(value, string(unit))
}

func ReportMetricDuration(b *testing.B, value time.Duration) {
	ReportMetric(b, float64(value.Milliseconds()), MetricDuration)
}
