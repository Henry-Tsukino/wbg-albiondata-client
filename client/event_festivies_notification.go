package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/ao-data/albiondata-client/log"
	uuid "github.com/nu7hatch/gouuid"
)

// входящий VECTOR EVENT
type eventFestivitiesNotification struct {
	Field0      []byte   `mapstructure:"0"`
	CatActivity []string `mapstructure:"1"`
	Activity    []string `mapstructure:"2"`
	StartTime   []int64  `mapstructure:"3"`
	EndTime     []int64  `mapstructure:"4"`
}

// нормальный батч-пейлоад (ОДИН запрос)
type FestivitiesPayload struct {
	Name        string   `json:"name"`
	CatActivity []string `json:"CatActivity"`
	Activity    []string `json:"Activity"`
	StartTime   []int64  `json:"TimeStart"`
	EndTime     []int64  `json:"TimeEnd"`
}

func (event eventFestivitiesNotification) Process(state *albionState) {

	if len(event.Activity) == 0 {
		return
	}

	log.Infof("daily activity send")

	identifier, _ := uuid.NewV4()

	n := len(event.Activity)

	cat := make([]string, 0, n)
	act := make([]string, 0, n)
	start := make([]int64, 0, n)
	end := make([]int64, 0, n)

	for i := 0; i < n; i++ {

		if i >= len(event.CatActivity) ||
			i >= len(event.StartTime) ||
			i >= len(event.EndTime) {
			continue
		}

		cat = append(cat, event.CatActivity[i])
		act = append(act, event.Activity[i])
		start = append(start, event.StartTime[i])
		end = append(end, event.EndTime[i])
	}

	payload := FestivitiesPayload{
		Name:        state.CharacterName,
		CatActivity: cat,
		Activity:    act,
		StartTime:   start,
		EndTime:     end,
	}

	sendFestivitiesToHTTP(payload, identifier.String())
}
func sendFestivitiesToHTTP(payload FestivitiesPayload, identifier string) {

	customURL := IPinok + "/festivities.ingest"

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Error marshaling Festivities payload: %v", err)
		return
	}

	req, err := http.NewRequest("POST", customURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("Error creating Festivities request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Identifier", identifier)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error sending Festivities to %s: %v", customURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Infof("Successfully sent Festivities event to %s (id=%s)", customURL, identifier)
	} else {
		log.Warnf("Festivities endpoint returned status: %d", resp.StatusCode)
	}
}
