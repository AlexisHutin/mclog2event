package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"mclog2event/config"
	"mclog2event/matcher"
	"mclog2event/pusher"
	"mclog2event/watcher"
)

func main() {
	log.Println("Starting mclog2event...")

	logFilePath := os.Getenv("LOG_FILE_PATH")
	webhookURL := os.Getenv("WEBHOOK_URL")
	eventConfigPath := os.Getenv("EVENT_CONFIG_PATH")

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
			if data := m.Match(line); data != nil {
				go p.Push(*data)
			}
		case <-sigs:
			log.Println("Exiting...")
			return
		}
	}
}
