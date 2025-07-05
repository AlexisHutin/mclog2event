package telemetry

import (
	"context"
	"log"
	"net/http"

	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	stdprom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Meter metric.Meter

	LogsParsedCount    metric.Int64Counter
	LogsParsedDuration metric.Int64Histogram
	MatchCount         metric.Int64Counter
	MatchDuration      metric.Int64Histogram
	PushCount          metric.Int64Counter
	PushDuration       metric.Int64Histogram
)

func Init(ctx context.Context) (func(context.Context) error, error) {
	reg := stdprom.NewRegistry()
	exporter, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return nil, err
	}

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go func() {
		log.Println("Prometheus metrics available at http://localhost:2222/metrics")
		if err := http.ListenAndServe(":2222", nil); err != nil {
			log.Fatalf("failed to start metrics server: %v", err)
		}
	}()

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	Meter = provider.Meter("mclog2event")

	LogsParsedCount, _ = Meter.Int64Counter("logs.parsed.count")
	LogsParsedDuration, _ = Meter.Int64Histogram("logs.parsed.duration.ms")
	MatchCount, _ = Meter.Int64Counter("logs.match.count")
	MatchDuration, _ = Meter.Int64Histogram("logs.match.duration.ms")
	PushCount, _ = Meter.Int64Counter("logs.push.count")
	PushDuration, _ = Meter.Int64Histogram("logs.push.duration.ms")

	return provider.Shutdown, nil
}
