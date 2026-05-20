package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

/*
The event is received when player enters red zone.

EventDataType: [474]evRedZonePlayerNotification - map[0:2 1:0 3:0 252:474]
*/

type eventRedZonePlayerNotification struct {
	EventTime int64 `mapstructure:"0"`
	Unknown1  int   `mapstructure:"1"`
	Unknown2  int   `mapstructure:"2"`
	Unknown3  int   `mapstructure:"3"`
}

type redZonePayload struct {
	Status0   int   `json:"Status0"`
	EventTime int64 `json:"0"`
	Unknown1  int   `json:"1"`
	Unknown2  int   `json:"2"`
	Unknown3  int   `json:"3"`
}

func (event eventRedZonePlayerNotification) Process(state *albionState) {
	log.Debug("Got red zone player notification...")

	// Always send, no rate limiting for this event
	log.Infof("Red Zone Player Notification detected: 0:%d 1:%d 2:%d 3:%d",
		event.EventTime,
		event.Unknown1,
		event.Unknown2,
		event.Unknown3)

	identifier, _ := uuid.NewV4()

	// Create custom payload with key:value format
	payload := redZonePayload{
		Status0: event.Unknown1,
	}

	// Send to custom HTTP endpoint
	go sendRedZoneToHTTP(payload, identifier.String())
}

func sendRedZoneToHTTP(payload redZonePayload, identifier string) {
	// Custom endpoint
	customURL := "http://82.38.2.16:3002/redzoneventclusterstatus.ingest"

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Error marshaling red zone payload: %v", err)
		return
	}

	req, err := http.NewRequest("POST", customURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("Error creating red zone request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Identifier", identifier)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error sending red zone to %s: %v", customURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Infof("Successfully sent red zone notification to %s (Identifier: %s)", customURL, identifier)
	} else {
		log.Warnf("Red zone endpoint returned status: %d", resp.StatusCode)
	}
}
