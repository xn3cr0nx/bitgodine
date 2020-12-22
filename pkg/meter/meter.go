package meter

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"go.opentelemetry.io/otel/metric"
)

var meter *metric.Meter

// Config tracer configuration
type Config struct {
	Name string
}

// NewMeter singleton implementation returns default meter
func NewMeter(conf *Config) (*metric.Meter, error) {
	if meter == nil {
		if err := configure(conf); err != nil {
			return nil, err
		}

		m := otel.Meter(conf.Name)
		meter = &m
	}
	return meter, nil
}

func configure(conf *Config) (err error) {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
	if err != nil {
		return
	}
	http.HandleFunc("/", exporter.ServeHTTP)
	go func() {
		_ = http.ListenAndServe(":9464", nil)
	}()

	fmt.Println("Prometheus server running on :9464")
	return
}
