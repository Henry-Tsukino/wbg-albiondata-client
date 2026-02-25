package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ao-data/albiondata-client/log"
)

type httpUploader struct {
	baseURL   string
	transport *http.Transport
}

// newHTTPUploader creates a new HTTP uploader
func newHTTPUploader(url string) uploader {
	return &httpUploader{
		baseURL:   url,
		transport: &http.Transport{},
	}
}

func (u *httpUploader) sendToIngest(body []byte, topic string, state *albionState, identifier string) {
	// not handling sending identifier since the official usage is with http_pow

	client := &http.Client{Transport: u.transport}

	fullURL := u.baseURL + "/" + topic

	log.Debugf("Sending to HTTP uploader: %v (topic: %v)", u.baseURL, topic)

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer([]byte(body)))
	if err != nil {
		log.Errorf("Error while create new request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Error while sending ingest with data to %v: %v", u.baseURL, err)
		return
	}

	if resp.StatusCode != 200 {
		log.Errorf("Got bad response code from %v: %v", u.baseURL, resp.StatusCode)
		return
	}

	// See: https://stackoverflow.com/questions/17948827/reusing-http-connections-in-golang
	io.Copy(ioutil.Discard, resp.Body)

	log.Infof("✓ Successfully sent data to %v (%v)", u.baseURL, topic)

	defer resp.Body.Close()
}
