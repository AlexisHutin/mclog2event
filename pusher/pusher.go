package pusher

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"mclog2event/types"
)

// Pusher sends events to a webhook.
type Pusher struct {
	webhookURL string
	client     *http.Client
}

// NewPusher creates a new Pusher.
func NewPusher(webhookURL string) *Pusher {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	return &Pusher{
		webhookURL: webhookURL,
		client:     client,
	}
}

// Push sends an event to the webhook with exponential backoff.
func (p *Pusher) Push(eventData types.EventPayload) error {
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	retries := 5
	backoff := 1 * time.Second

	for i := 0; i < retries; i++ {
		req, err := http.NewRequest("POST", p.webhookURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		log.Printf("Pushing...")
		resp, err := p.client.Do(req)

		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			log.Printf("Success! (code %v)", resp.StatusCode)
			return nil
		}
		if resp != nil {
			resp.Body.Close()
			log.Printf("Success!")
		}

		log.Println(err)
		log.Printf("failed to push event, retrying in %v...", backoff)
		time.Sleep(backoff)
		backoff *= 2
	}

	return fmt.Errorf("failed to push event after %d retries", retries)
}
