// internal/dashboard/service.go
package dashboard

import "github.com/Henry-Tsukino/wbg-albiondata-client/log"

// DashboardService is bound to the Wails frontend (see application.NewService
// in wbg-client.go). Its methods back-fill the dashboard window's
// state when it's shown; ongoing updates arrive via the "status:changed",
// "counters:snapshot", and "log:line" events instead.
type DashboardService struct{}

// GetStatus returns the current status snapshot.
func (s *DashboardService) GetStatus() Status {
	return GetStatus()
}

// GetUploadCounts returns the current upload counters by topic.
func (s *DashboardService) GetUploadCounts() map[string]int64 {
	return GetUploadCounts()
}

// GetRecentLogs returns the currently buffered recent log lines.
func (s *DashboardService) GetRecentLogs() []LogLine {
	return GetRecentLogs()
}

// IsMarketDisabled reports whether market data parsing/uploading is
// currently disabled.
func (s *DashboardService) IsMarketDisabled() bool {
	return GetStatus().MarketDisabled
}

// SetUploadDisabled is bound to the dashboard's "send data" toggle.
func (s *DashboardService) SetUploadDisabled(disabled bool) {
	log.Infof("UI CLICK -> SetUploadDisabled(%v)", disabled)
	SetUploadDisabled(disabled)
}

// SetMarketDisabled is bound to the dashboard's market toggle.
func (s *DashboardService) SetMarketDisabled(disabled bool) {
	log.Infof("UI CLICK -> SetMarketDisabled(%v)", disabled)
	SetMarketDisabled(disabled)
}
