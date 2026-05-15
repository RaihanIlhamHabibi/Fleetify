package service

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"fleetify/models"
)

type WebhookPayload struct {
	ReportID     uint64               `json:"report_id"`
	Status       models.ReportStatus  `json:"status"`
	LicensePlate string               `json:"license_plate,omitempty"`
	Timestamp    string               `json:"timestamp"`
}

type WebhookService struct {
	url string
}

func NewWebhookService(url string) *WebhookService {
	return &WebhookService{url: url}
}

func (s *WebhookService) NotifyStatusChange(reportID uint64, status models.ReportStatus, licensePlate string) {
	if s.url == "" {
		return
	}
	if status != models.StatusApproved && status != models.StatusCompleted {
		return
	}

	go func() {
		payload := WebhookPayload{
			ReportID:     reportID,
			Status:       status,
			LicensePlate: licensePlate,
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
		}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("webhook marshal error: %v", err)
			return
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(s.url, "application/json", bytes.NewReader(body))
		if err != nil {
			log.Printf("webhook POST error: %v", err)
			return
		}
		defer resp.Body.Close()
		log.Printf("webhook sent for report %d status %s -> %d", reportID, status, resp.StatusCode)
	}()
}
