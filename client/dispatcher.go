package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"strings"

	"github.com/Henry-Tsukino/wbg-albiondata-client/internal/dashboard"
	"github.com/Henry-Tsukino/wbg-albiondata-client/lib"
	"github.com/Henry-Tsukino/wbg-albiondata-client/log"
)

type dispatcher struct{}

var (
	wsHub *WSHub
	dis   *dispatcher

	// uploaderCacheMu guards uploaderCache.
	uploaderCacheMu sync.Mutex
	// uploaderCache holds one uploader per ingest target, reused across
	// every upload. Every market order, gold price, or mail upload calls
	// createUploaders, and building a fresh uploader per call used to open
	// a brand new, never-closed NATS connection (with its own internal
	// reconnect/buffering goroutines) or a fresh http.Transport whose idle
	// keep-alive connections were never reclaimed - both leaked memory
	// continuously, and far more sharply during network instability, when
	// abandoned NATS connections would sit retrying with buffered data.
	uploaderCache = map[string]uploader{}
)

func createDispatcher() {
	dis = &dispatcher{}

	if ConfigGlobal.EnableWebsockets {
		wsHub = newHub()
		go wsHub.run()
		go runHTTPServer()
	}
}

func createUploaders(targets []string) []uploader {
	uploaderCacheMu.Lock()
	defer uploaderCacheMu.Unlock()

	var uploaders []uploader
	for _, target := range targets {
		if target == "" {
			continue
		}
		// "noop" is the documented "send nowhere" target: no uploader.
		if target == "noop" {
			continue
		}
		if len(target) < 4 {
			log.Infof("Got an ingest target that was less than 4 characters, not a valid ingest target: %v", target)
			continue
		}

		if cached, ok := uploaderCache[target]; ok {
			uploaders = append(uploaders, cached)
			continue
		}

		// strings.HasPrefix is length-safe; slicing target[0:8] panicked
		// for any target shorter than 8 characters (e.g. "noop").
		var u uploader
		if strings.HasPrefix(target, "http+pow") || strings.HasPrefix(target, "https+pow") {
			u = newHTTPUploaderPow(target)
		} else if strings.HasPrefix(target, "http") {
			u = newHTTPUploader(target)
		} else if strings.HasPrefix(target, "nats") {
			u = newNATSUploader(target)
		} else {
			log.Infof("An invalid ingest target was specified: %v", target)
			continue
		}

		uploaderCache[target] = u
		uploaders = append(uploaders, u)
	}

	return uploaders
}

// sendMsgToPublicUploaders sends MARKET data (orders, histories) and GOLD
// prices to the public Albion Data Project servers. Nothing else may use
// it: every other kind of upload goes to the WBG server through
// sendMsgToWBGUploaders / sendMsgToPrivateUploaders.
func sendMsgToPublicUploaders(upload interface{}, topic string, state *albionState, identifier string, recordCount int) {
	dashboard.IncrementCounterBy(topic, int64(recordCount))

	data, err := json.Marshal(upload)
	if err != nil {
		log.Errorf("Error while marshalling payload for %v: %v", topic, err)
		return
	}

	var PublicIngestBaseUrls = ConfigGlobal.PublicIngestBaseUrls
	// http+pow://albion-online-data.com is used as a magic placeholder for every realm there is
	if strings.Contains(ConfigGlobal.PublicIngestBaseUrls, "https+pow://albion-online-data.com") {
		// we replace the placeholder with the correct one based on the serverID from albionState
		PublicIngestBaseUrls = strings.Replace(PublicIngestBaseUrls, "https+pow://albion-online-data.com", state.AODataIngestBaseURL, -1)
	}

	var publicUploaders = createUploaders(strings.Split(PublicIngestBaseUrls, ","))

	sendMsgToUploaders(data, topic, publicUploaders, state, identifier)

	// If websockets are enabled, send the data there too
	if ConfigGlobal.EnableWebsockets {
		sendMsgToWebSockets(data, topic)
	}
}

// sendMsgToWBGUploaders sends non-market, non-gold data (map data, bandit
// events, festivities, ...) to the WBG server (ConfigGlobal.PrivateIngestBaseUrls,
// default DefaultWBGIngestURL). It never touches the public servers.
func sendMsgToWBGUploaders(upload interface{}, topic string, state *albionState, identifier string, recordCount int) {
	dashboard.IncrementCounterBy(topic, int64(recordCount))

	data, err := json.Marshal(upload)
	if err != nil {
		log.Errorf("Error while marshalling payload for %v: %v", topic, err)
		return
	}

	var wbgUploaders = createUploaders(strings.Split(ConfigGlobal.PrivateIngestBaseUrls, ","))
	sendMsgToUploaders(data, topic, wbgUploaders, state, identifier)

	// If websockets are enabled, send the data there too
	if ConfigGlobal.EnableWebsockets {
		sendMsgToWebSockets(data, topic)
	}
}

func sendMsgToPrivateUploaders(upload lib.PersonalizedUpload, topic string, state *albionState, identifier string) {
	if ConfigGlobal.DisableUpload {
		log.Info("Upload is disabled.")
		return
	}

	// TODO: Re-enable this when issue #14 is fixed
	// Will personalize with blanks for now in order to allow people to see the format
	// if state.CharacterName == "" || state.CharacterId == "" {
	// 	log.Error("The player name or id has not been set. Please restart the game and make sure the client is running.")
	// 	notification.Push("The player name or id has not been set. Please restart the game and make sure the client is running.")
	// 	return
	// }

	upload.Personalize(state.CharacterId, state.CharacterName)

	data, err := json.Marshal(upload)
	if err != nil {
		log.Errorf("Error while marshalling payload for %v: %v", topic, err)
		return
	}

	var privateUploaders = createUploaders(strings.Split(ConfigGlobal.PrivateIngestBaseUrls, ","))
	if len(privateUploaders) > 0 {
		sendMsgToUploaders(data, topic, privateUploaders, state, identifier)
	}

	// If websockets are enabled, send the data there too
	if ConfigGlobal.EnableWebsockets {
		sendMsgToWebSockets(data, topic)
	}
}

func sendMsgToUploaders(msg []byte, topic string, uploaders []uploader, state *albionState, identifier string) {
	if ConfigGlobal.DisableUpload {
		log.Info("Upload is disabled.")
		return
	}

	for _, u := range uploaders {
		u.sendToIngest(msg, topic, state, identifier)
	}
}

func runHTTPServer() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(wsHub, w, r)
	})

	err := http.ListenAndServe(":8099", nil)

	if err != nil {
		log.Panic("ListenAndServe: ", err)
	}
}

func sendMsgToWebSockets(msg []byte, topic string) {
	// TODO (gradius): send JSON data with topic string
	// TODO (gradius): this seems super hacky, and I'm sure there's a better way.
	var result string
	result = "{\"topic\": \"" + topic + "\", \"data\": " + string(msg) + "}"
	wsHub.broadcast <- []byte(result)
}

// postBanditEvent POSTs a bandit event as JSON to an arbitrary URL.
func postBanditEvent(upload interface{}, url string) error {
	data, err := json.Marshal(upload)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
