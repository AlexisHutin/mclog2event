package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mclog2event/config"
	"mclog2event/matcher"
	"mclog2event/pusher"
	"mclog2event/telemetry"
	"mclog2event/types"
	"mclog2event/watcher"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	log.Println("Starting mclog2event...")

	logFilePath := os.Getenv("LOG_FILE_PATH")
	webhookURL := os.Getenv("WEBHOOK_URL")
	eventConfigPath := os.Getenv("EVENT_CONFIG_PATH")
	ctx := context.Background()

	shutdown, err := telemetry.Init(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer shutdown(ctx)

	if logFilePath == "" || webhookURL == "" || eventConfigPath == "" {
		log.Fatal("LOG_FILE_PATH, WEBHOOK_URL, and EVENT_CONFIG_PATH must be set")
	}

	cfg, err := config.LoadConfig(eventConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	m, err := matcher.NewMatcher(cfg.Events)
	if err != nil {
		log.Fatalf("Failed to create matcher: %v", err)
	}

	p := pusher.NewPusher(webhookURL)
	lines := make(chan string)

	go watcher.Watch(logFilePath, lines)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case line := <-lines:
			parseStart := time.Now()
			var matched bool

			if data := m.Match(line); data != nil {
				matched = true
				telemetry.MatchCount.Add(ctx, 1, metric.WithAttributes(attribute.String("type", data.EventType)))
				telemetry.MatchDuration.Record(ctx, time.Since(parseStart).Milliseconds(), metric.WithAttributes(attribute.String("type", data.EventType)))

				go func(data types.EventPayload) {
					pushStart := time.Now()
					p.Push(data)
					telemetry.PushCount.Add(ctx, 1, metric.WithAttributes(attribute.String("type", data.EventType)))
					telemetry.PushDuration.Record(ctx, time.Since(pushStart).Milliseconds(), metric.WithAttributes(attribute.String("type", data.EventType)))
				}(*data)
			}

			telemetry.LogsParsedCount.Add(ctx, 1, metric.WithAttributes(attribute.Bool("matched", matched)))
			telemetry.LogsParsedDuration.Record(ctx, time.Since(parseStart).Milliseconds(), metric.WithAttributes(attribute.Bool("matched", matched)))

		case <-sigs:
			log.Println("Exiting...")
			return
		}
	}

}
