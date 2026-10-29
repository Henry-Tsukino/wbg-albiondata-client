// internal/dashboard/status.go
package dashboard

import "sync"

// Status is a point-in-time snapshot of the client's observable state,
// pushed to the dashboard frontend whenever it changes.
type Status struct {
	Version            string
	UpdateAvailable    string
	CaptureRunning     bool
	CaptureError       bool
	ServerID           int
	IngestBaseURL      string
	CustomPublicIngest bool
	DriverWarning      string
	DriverHelpURL      string
	// EncryptionStatus is "" (unknown - nothing observed yet this
	// session), "encrypted" (a market data response came back
	// encrypted), or "clear" (a market data response came back and
	// decoded normally). See client/albion_state.go's
	// ShouldNotifyMarketDataEncrypted for how "encrypted" is decided.
	EncryptionStatus string
	// MarketDisabled reports whether market data parsing/uploading is
	// currently disabled (via -no-market or config.yaml DisableMarket).
	MarketDisabled bool
	// UploadDisabled reports whether ALL data sending to remote servers
	// is currently disabled (via -d or config.yaml DisableUpload).
	UploadDisabled bool
}

var (
	statusMu   sync.Mutex
	status     Status
	statusEmit func(Status)
	// marketHook is called when the market toggle changes, so the client
	// package (which owns ConfigGlobal) can update its runtime flag
	// without dashboard importing client.
	marketHook func(bool)
	// uploadHook is the same for the "send no data at all" toggle.
	uploadHook func(bool)
)

// OnStatusChange registers the callback invoked whenever the status
// changes. Only one callback is supported; a later call replaces the
// earlier one.
func OnStatusChange(fn func(Status)) {
	statusMu.Lock()
	statusEmit = fn
	statusMu.Unlock()
}

// GetStatus returns the current status snapshot.
func GetStatus() Status {
	statusMu.Lock()
	defer statusMu.Unlock()
	return status
}

// SetVersionInfo records the running client version and, if one is known,
// the version available to update to (empty string if none).
func SetVersionInfo(version, updateAvailable string) {
	statusMu.Lock()
	next := status
	next.Version = version
	next.UpdateAvailable = updateAvailable
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
}

// SetCaptureRunning records whether packet capture is currently active.
func SetCaptureRunning(running bool) {
	statusMu.Lock()
	next := status
	next.CaptureRunning = running
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
}

// SetCaptureError records that packet capture has stopped due to a fatal
// error and will not resume without a restart. Unlike CaptureRunning,
// this is not expected to be cleared during the process's lifetime once
// set - a fatal capture error is terminal for that run.
func SetCaptureError(failed bool) {
	statusMu.Lock()
	next := status
	next.CaptureError = failed
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
}

// SetServer records the Albion Online server detected from captured
// traffic and the ingest URL that maps to it.
func SetServer(serverID int, ingestBaseURL string) {
	statusMu.Lock()
	next := status
	next.ServerID = serverID
	next.IngestBaseURL = ingestBaseURL
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
}

// SetCustomPublicIngest records whether the user has configured a custom
// (non-default) public ingest URL via -i, meaning uploads go to a
// self-hosted server instead of the Albion Data Project's public
// endpoint - not to be confused with ConfigGlobal.PrivateIngestBaseUrls
// (-p), a separate, unrelated feature for uploading a private copy of
// data to a second destination.
func SetCustomPublicIngest(custom bool) {
	statusMu.Lock()
	next := status
	next.CustomPublicIngest = custom
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
}

// SetDriverWarning records a problem with the system's packet-capture
// driver that the user should act on (see internal/pcapdriver), or
// clears it by passing empty strings.
func SetDriverWarning(message, helpURL string) {
	statusMu.Lock()
	next := status
	next.DriverWarning = message
	next.DriverHelpURL = helpURL
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
}

// EncryptionStatus values for SetEncryptionStatus.
const (
	EncryptionUnknown  = ""
	EncryptionDetected = "encrypted"
	EncryptionClear    = "clear"
)

// SetEncryptionStatus records whether market data has most recently come
// back encrypted, come back clear, or is unknown (EncryptionUnknown,
// the zero value - nothing observed yet, or reset because capture went
// idle and any prior observation is stale).
func SetEncryptionStatus(encryptionStatus string) {
	statusMu.Lock()
	next := status
	next.EncryptionStatus = encryptionStatus
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
}

// OnMarketChange registers a callback invoked when the market
// enabled/disabled state is changed (e.g. from the dashboard toggle).
func OnMarketChange(fn func(bool)) {
	statusMu.Lock()
	marketHook = fn
	statusMu.Unlock()
}

// SetMarketDisabled records whether market data parsing/uploading is
// currently disabled and notifies the registered hook so the runtime
// flag is kept in sync.
func SetMarketDisabled(disabled bool) {
	statusMu.Lock()
	next := status
	next.MarketDisabled = disabled
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	hook := marketHook
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
	if hook != nil {
		hook(disabled)
	}
}

// OnUploadChange registers a callback invoked when the "send data"
// state is changed (e.g. from the dashboard toggle).
func OnUploadChange(fn func(bool)) {
	statusMu.Lock()
	uploadHook = fn
	statusMu.Unlock()
}

// SetUploadDisabled records whether all data sending is currently
// disabled and notifies the registered hook so the runtime flag is kept
// in sync.
func SetUploadDisabled(disabled bool) {
	statusMu.Lock()
	next := status
	next.UploadDisabled = disabled
	changed := next != status
	if changed {
		status = next
	}
	emit := statusEmit
	hook := uploadHook
	statusMu.Unlock()

	if changed && emit != nil {
		emit(next)
	}
	if hook != nil {
		hook(disabled)
	}
}

// resetStatusForTest clears all package state. Test-only.
func resetStatusForTest() {
	statusMu.Lock()
	status = Status{}
	statusEmit = nil
	statusMu.Unlock()
}
