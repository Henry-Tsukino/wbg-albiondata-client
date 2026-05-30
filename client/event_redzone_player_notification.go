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

Known values for map[0] (EventStatus):
  0 = warning (attack announced ~15 min before)
  2 = attack is happening now

EventDataType: [474]evRedZonePlayerNotification - map[0:2 1:0 3:0 252:474]

map[0] - EventStatus: current phase/status of the red zone event
map[1] - Unknown1
map[2] - Unknown2
map[3] - Unknown3
*/

type eventRedZonePlayerNotification struct {
	EventStatus int `mapstructure:"0"`
	Unknown1    int `mapstructure:"1"`
	Unknown2    int `mapstructure:"2"`
	Unknown3    int `mapstructure:"3"`
}

type redZonePayload struct {
	EventStatus int    `json:"status0"`
	Name        string `json:"name"`
}

func (event eventRedZonePlayerNotification) Process(state *albionState) {
	log.Debug("Got red zone player notification...")
	log.Infof("Red Zone Player Notification detected: status:%d unknown1:%d unknown2:%d unknown3:%d",
		event.EventStatus,
		event.Unknown1,
		event.Unknown2,
		event.Unknown3)

	if event.EventStatus != 0 {
		log.Infof("Skipping red zone notification: EventStatus=%d (only forwarding status=0 warnings)", event.EventStatus)
		return
	}

	identifier, _ := uuid.NewV4()
	payload := redZonePayload{
		EventStatus: event.EventStatus, // значение первого unknown
		Name:        state.CharacterName,
	}
	go sendRedZoneToHTTP(payload, identifier.String())
}

func sendRedZoneToHTTP(payload redZonePayload, identifier string) {
	customURL := "http://95.111.247.74:3008/redzoneventclusterstatus.ingest"

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
		log.Infof("Successfully sent red zone warning to %s (Identifier: %s)", customURL, identifier)
	} else {
		log.Warnf("Red zone endpoint returned status: %d", resp.StatusCode)
	}
}


