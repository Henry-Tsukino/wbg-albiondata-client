package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"sync"

	"github.com/ao-data/albiondata-client/log"
)

var (
	fourDigitRe       = regexp.MustCompile(`^\d{4}$`)
	validLocFormat    = regexp.MustCompile(`^(\d{4}|TNL-\d{3})$`)
	excludedLocations = map[string]bool{
		// excludedLocations — конкретные ID, которые проходят по формату но типа не нада
		// Города (основные)
		"0000": true, // Thetford
		"1000": true, // Lymhurst
		"2000": true, // Bridgewatch
		"3003": true, // Caerleon
		"3004": true, // Martlock
		"4000": true, // Fort Sterling
		"5000": true, // Brecilien
		"5001": true, // Brecilien

		// Банки и рынки городов
		"0006": true, // Bank of Thetford
		"0007": true, // Thetford Market
		"1001": true, // Bank of Lymhurst
		"1002": true, // Lymhurst Market
		"2003": true, // Bank of Bridgewatch
		"2004": true, // Bridgewatch Market
		"3005": true, // Caerleon Market
		"3006": true, // Bank of Caerleon
		"3007": true, // Bank of Martlock
		"3008": true, // Martlock Market
		"4001": true, // Bank of Fort Sterling
		"4002": true, // Fort Sterling Market
		"5002": true, // Brecilien Bank
		"5003": true, // Brecilien Market

		// Порталы городов (переход город <-> открытый мир)
		"0301": true, // Thetford Portal
		"1301": true, // Lymhurst Portal
		"2301": true, // Bridgewatch Portal
		"3301": true, // Martlock Portal
		"4301": true, // Fort Sterling Portal

		// Ресты (хабы вход/выход из royal continent)
		"0008": true, // Morgana's Rest
		"1012": true, // Merlyn's Rest
		"4300": true, // Arthur's Rest
	}
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
	if !validLocFormat.MatchString(location) {
		log.Debugf("[LocationBuffer] Add: %q, пропущено", location)
		return
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	if excludedLocations[location] {
		log.Debugf("[LocationBuffer] Add: %q, буфер очищен", location)
		lb.locations = lb.locations[:0]
		return
	}

	for i, loc := range lb.locations {
		if loc == location {
			lb.locations = lb.locations[:i+1]
			log.Debugf("[LocationBuffer] Add: %q, обрезано до %d", location, len(lb.locations))
			lb.checkAndingest()
			return
		}
	}
	lb.locations = append(lb.locations, location)
	if len(lb.locations) > 10 {
		lb.locations = lb.locations[len(lb.locations)-10:]
	}
	log.Debugf("[LocationBuffer] Add:  %q, count=%d", location, len(lb.locations))
	log.Debugf("[LocationBuffer] порядок: %v", lb.locations)
	lb.checkAndingest()
}

// checkAndingest вызывается только внутри Add, мьютекс уже захвачен
func (lb *LocationBuffer) checkAndingest() {
	for {
		hasTNL := false
		fourDigitIdx := -1
		for i, loc := range lb.locations {
			if loc == "TNL-167" {
				hasTNL = true
			}
			if fourDigitRe.MatchString(loc) {
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
			resp, err := http.Post(IPinok+"/avaRout.ingest", "application/json", bytes.NewReader(b))
			if err != nil {
				log.Errorf("[LocationBuffer] ingest error: %v", err)
				return
			}
			defer resp.Body.Close()
			log.Debugf("[LocationBuffer] ingest отправлен, статус: %d", resp.StatusCode)
		}(body)
		lb.locations = append(lb.locations[:fourDigitIdx], lb.locations[fourDigitIdx+1:]...)
	}
}

func (lb *LocationBuffer) GetLast() string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	if len(lb.locations) == 0 {
		log.Debugf("[LocationBuffer] GetLast: буфер пуст")
		return ""
	}
	result := lb.locations[len(lb.locations)-1]
	log.Debugf("[LocationBuffer] GetLast: %q", result)
	return result
}

func (lb *LocationBuffer) GetAll() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]string, len(lb.locations))
	copy(result, lb.locations)
	log.Debugf("[LocationBuffer] GetAll: %d локаций: %v", len(result), result)
	return result
}
