package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"sync"

	"github.com/ao-data/albiondata-client/log"
)

type LocationBuffer struct {
	locations     []string
	characterName string
	mu            sync.RWMutex
}

func NewLocationBuffer() *LocationBuffer {
	log.Info("[LocationBuffer] инициализация нового буфера локаций")
	return &LocationBuffer{
		locations: make([]string, 0, 10),
	}
}

func (lb *LocationBuffer) SetCharacterName(name string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.characterName = name
}

func (lb *LocationBuffer) Add(location string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, loc := range lb.locations {
		if loc == location {
			lb.locations = lb.locations[:i+1]
			log.Infof("[LocationBuffer] Add: найдена %q, обрезано до %d", location, len(lb.locations))
			lb.checkAndIngest()
			return
		}
	}

	lb.locations = append(lb.locations, location)
	if len(lb.locations) > 10 {
		lb.locations = lb.locations[len(lb.locations)-10:]
	}
	log.Infof("[LocationBuffer] Add: записана %q, count=%d", location, len(lb.locations))
	log.Infof("[LocationBuffer] порядок локаций: %v", lb.locations)
	lb.checkAndIngest()
}

// checkAndIngest вызывается только внутри Add, мьютекс уже захвачен
func (lb *LocationBuffer) checkAndIngest() {
	var re = regexp.MustCompile(`^\d{4}$`)

	for {
		hasTNL := false
		fourDigitIdx := -1

		for i, loc := range lb.locations {
			if loc == "TNL-167" {
				hasTNL = true
			}
			if re.MatchString(loc) {
				fourDigitIdx = i
			}
		}

		if !hasTNL || fourDigitIdx == -1 {
			return
		}

		payload := map[string]interface{}{
			"user":      lb.characterName,
			"locations": append([]string{}, lb.locations...),
		}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Errorf("[LocationBuffer] JSON marshal error: %v", err)
			return
		}

		go func(b []byte) {
			resp, err := http.Post(IPinok+"/avaRout.Ingest", "application/json", bytes.NewReader(b))
			if err != nil {
				log.Errorf("[LocationBuffer] Ingest error: %v", err)
				return
			}
			defer resp.Body.Close()
			log.Infof("[LocationBuffer] Ingest отправлен, статус: %d", resp.StatusCode)
		}(body)

		lb.locations = append(lb.locations[:fourDigitIdx], lb.locations[fourDigitIdx+1:]...)
	}
}

func (lb *LocationBuffer) GetLast() string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.locations) == 0 {
		log.Info("[LocationBuffer] GetLast: буфер пуст")
		return ""
	}
	result := lb.locations[len(lb.locations)-1]
	log.Infof("[LocationBuffer] GetLast: %q", result)
	return result
}

func (lb *LocationBuffer) GetAll() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	result := make([]string, len(lb.locations))
	copy(result, lb.locations)
	log.Infof("[LocationBuffer] GetAll: %d локаций: %v", len(result), result)
	return result
}
