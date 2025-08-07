package telemetry

import (
	"errors"

	"go.opentelemetry.io/otel/metric"
)

var (
	PlayCommandExecutionsCounter metric.Int64Counter
	ActiveAudioGraphsGauge       metric.Int64Gauge
)

var _ error = (*MetricRegistrationError)(nil)

type MetricRegistrationError struct {
	metricName string
}

func NewMetricRegistrationError(metricName string) *MetricRegistrationError {
	return &MetricRegistrationError{metricName: metricName}
}

func (e *MetricRegistrationError) Error() string {
	return "failed to register metric %s"
}

func setupMetrics(meter metric.Meter) (err error) {
	PlayCommandExecutionsCounter, err = meter.Int64Counter("counter.commandExecutions.play")
	if err != nil {
		return errors.Join(NewMetricRegistrationError("counter.commandExecutions.play"), err)
	}

	ActiveAudioGraphsGauge, err = meter.Int64Gauge("gauge.audioGraphs.active")
	if err != nil {
		return errors.Join(NewMetricRegistrationError("gauge.audioGraphs.active"), err)
	}

	return nil
}
