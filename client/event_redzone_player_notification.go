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

Known values for map[0] (status0):
  0 = warning (attack announced ~15 min before)
  2 = attack is happening now

EventDataType: [474]evRedZonePlayerNotification - map[0:2 1:0 3:0 252:474]

map[0] - status0: current phase/status of the red zone event
map[1] - Unknown1
map[2] - Unknown2
map[3] - Unknown3
*/

type eventRedZonePlayerNotification struct {
	Status0  int `mapstructure:"0"`
	Unknown1 int `mapstructure:"1"`
	Unknown2 int `mapstructure:"2"`
	Unknown3 int `mapstructure:"3"`
}

type redZonePayload struct {
	Name    string `json:"name"`
	Status0 int    `json:"status0"`
}

func (event eventRedZonePlayerNotification) Process(state *albionState) {
	log.Debug("Got red zone player notification...")
	log.Infof("Red Zone Player Notification detected: status:%d unknown1:%d unknown2:%d unknown3:%d",
		event.Status0,
		event.Unknown1,
		event.Unknown2,
		event.Unknown3)

	if event.Status0 != 0 {
		log.Infof("Skipping red zone notification: status0=%d (only forwarding status=0 warnings)", event.Status0)
		return
	}

	identifier, _ := uuid.NewV4()
	payload := redZonePayload{
		Name:    state.CharacterName,
		Status0: event.Status0,
	}
	go sendRedZoneToHTTP(payload, identifier.String())
}

func sendRedZoneToHTTP(payload redZonePayload, identifier string) {
	customURL := IPinok + "/redzoneventclusterstatus.ingest"

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
