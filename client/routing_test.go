package client

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/Henry-Tsukino/wbg-albiondata-client/lib"
)

// countingServer returns a test server and a pointer to the number of
// requests it has received.
func countingServer(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// Market and gold data must reach only the public servers; everything else
// must reach only the WBG server.
func TestUploadRouting_MarketAndGoldPublic_RestWBG(t *testing.T) {
	public, publicHits := countingServer(t)
	wbg, wbgHits := countingServer(t)

	oldDisable := ConfigGlobal.DisableUpload
	oldPublic := ConfigGlobal.PublicIngestBaseUrls
	oldPrivate := ConfigGlobal.PrivateIngestBaseUrls
	oldWS := ConfigGlobal.EnableWebsockets
	t.Cleanup(func() {
		ConfigGlobal.DisableUpload = oldDisable
		ConfigGlobal.PublicIngestBaseUrls = oldPublic
		ConfigGlobal.PrivateIngestBaseUrls = oldPrivate
		ConfigGlobal.EnableWebsockets = oldWS
	})
	ConfigGlobal.DisableUpload = false
	ConfigGlobal.PublicIngestBaseUrls = public.URL
	ConfigGlobal.PrivateIngestBaseUrls = wbg.URL
	ConfigGlobal.EnableWebsockets = false

	state := &albionState{}

	for _, topic := range []string{
		lib.NatsMarketOrdersIngest,
		lib.NatsMarketHistoriesIngest,
		lib.NatsGoldPricesIngest,
	} {
		sendMsgToPublicUploaders(struct{}{}, topic, state, "id", 1)
	}
	if got := atomic.LoadInt32(publicHits); got != 3 {
		t.Fatalf("public server got %d requests for market/gold, want 3", got)
	}
	if got := atomic.LoadInt32(wbgHits); got != 0 {
		t.Fatalf("WBG server got %d market/gold requests, want 0", got)
	}

	for _, topic := range []string{
		lib.NatsMapDataIngest,
		lib.NatsBanditEvent,
		lib.NatsFestivitiesIngest,
	} {
		sendMsgToWBGUploaders(struct{}{}, topic, state, "id", 1)
	}
	if got := atomic.LoadInt32(wbgHits); got != 3 {
		t.Fatalf("WBG server got %d requests for other data, want 3", got)
	}
	if got := atomic.LoadInt32(publicHits); got != 3 {
		t.Fatalf("public server received non-market data (now %d requests, want 3)", got)
	}
}
