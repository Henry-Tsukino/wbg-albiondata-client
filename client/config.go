package client

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Henry-Tsukino/wbg-albiondata-client/log"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	logFileName = "wbg-client.log"
	maxLogFiles = 10
)

// ansiStripWriter wraps an io.Writer and strips ANSI escape codes before writing
type ansiStripWriter struct {
	writer io.Writer
	regex  *regexp.Regexp
}

// newAnsiStripWriter creates a writer that strips ANSI escape codes
func newAnsiStripWriter(w io.Writer) *ansiStripWriter {
	return &ansiStripWriter{
		writer: w,
		// Matches ANSI escape sequences like \x1b[0m, \x1b[36m, etc.
		regex: regexp.MustCompile(`\x1b\[[0-9;]*m`),
	}
}

func (w *ansiStripWriter) Write(p []byte) (n int, err error) {
	stripped := w.regex.ReplaceAll(p, []byte{})
	_, err = w.writer.Write(stripped)
	// Return original length to satisfy io.Writer contract
	return len(p), err
}

type config struct {
	AllowedWSHosts                 []string
	Debug                          bool
	Trace                          bool
	DebugEvents                    map[int]bool
	DebugEventsString              string
	DebugEventsBlacklistString     string
	DebugOperations                map[int]bool
	DebugOperationsString          string
	DebugOperationsBlacklistString string
	DebugIgnoreDecodingErrors      bool
	DisableUpload                  bool
	DisableMarket                  bool
	BanditWebhookURL               string
	EnableWebsockets               bool
	ListenDevices                  string
	LogLevel                       string
	Minimize                       bool
	Offline                        bool
	OfflinePath                    string
	RecordPath                     string
	PrivateIngestBaseUrls          string
	PublicIngestBaseUrls           string
	NoCPULimit                     bool
	PrintVersion                   bool
	UpdateGithubOwner              string
	UpdateGithubRepo               string
}

const (
	// DefaultWBGIngestURL receives every upload except market and gold
	// data (those still go to the public Albion Data Project servers).
	DefaultWBGIngestURL = "https://hermes-production-817b.up.railway.app"

	// Self-update source: the latest GitHub release of this repository.
	DefaultUpdateGithubOwner = "Henry-Tsukino"
	DefaultUpdateGithubRepo  = "wbg-albiondata-client"
)

// config global config data
var ConfigGlobal = &config{
	LogLevel:          "INFO",
	UpdateGithubOwner: DefaultUpdateGithubOwner,
	UpdateGithubRepo:  DefaultUpdateGithubRepo}

func (config *config) SetupFlags() {
	config.setupWebsocketFlags()
	config.setupDebugFlags()
	config.setupCommonFlags()

	flag.Parse()

	// Merge in the config-file value AFTER flag.Parse, since flag.Parse
	// resets a registered bool to its default (false) when the flag is
	// absent - combining earlier would be overwritten.
	if viper.GetBool("DisableMarket") {
		config.DisableMarket = true
	}

	if config.OfflinePath != "" {
		config.Offline = true
		config.DisableUpload = true

		if config.PublicIngestBaseUrls == "http+pow://west.aodp.local:3000" {
			config.DisableUpload = false
		}

		log.Infof("config.PublicIngestBaseUrls: %v", config.PublicIngestBaseUrls)
		log.Infof("config.DisableUpload: %v", config.DisableUpload)
	}

	if config.DisableUpload {
		log.Info("Upload is disabled.")
	}

	config.setupLogs()
}

func (config *config) setupWebsocketFlags() {
	// Setup the config file and parse values
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	// if we cannot find the configuration file, set Websockets to false
	if err != nil {
		viper.Set("EnableWebsockets", false)
	}

	config.EnableWebsockets = viper.GetBool("EnableWebsockets")
	config.AllowedWSHosts = viper.GetStringSlice("AllowedWebsocketHosts")

	// // Keeping for local development, but commenting out so it's not live.
	// // Read update configuration (use defaults if not specified)
	// if viper.IsSet("UpdateGithubOwner") {
	// 	config.UpdateGithubOwner = viper.GetString("UpdateGithubOwner")
	// }
	// if viper.IsSet("UpdateGithubRepo") {
	// 	config.UpdateGithubRepo = viper.GetString("UpdateGithubRepo")
	// }
}

func (config *config) setupDebugFlags() {
	flag.BoolVar(
		&config.PrintVersion,
		"version",
		false,
		"Print version, then close.",
	)

	flag.BoolVar(
		&config.Debug,
		"debug",
		false,
		"Enable debug logging.",
	)

	flag.BoolVar(
		&config.Trace,
		"trace",
		false,
		"Enable trace logging. Even more verbose than debug.",
	)

	flag.StringVar(
		&config.DebugEventsString,
		"events",
		"",
		"Whitelist of event IDs to output messages when debugging. Comma separated.",
	)

	flag.StringVar(
		&config.DebugEventsBlacklistString,
		"events-ignore",
		"",
		"Blacklist of event IDs to hide messages when debugging. Comma separated.",
	)

	flag.StringVar(
		&config.DebugOperationsString,
		"operations",
		"",
		"Whitelist of operation IDs to output messages when debugging. Comma separated.",
	)

	flag.StringVar(
		&config.DebugOperationsBlacklistString,
		"operations-ignore",
		"",
		"Blacklist of operation IDs to hide messages when debugging. Comma separated.",
	)
	flag.StringVar(
		&config.PrivateIngestBaseUrls,
		"p",
		DefaultWBGIngestURL,
		"Base URL of the WBG server that receives everything except market and gold data. Comma separated.",
	)

	flag.BoolVar(
		&config.DebugIgnoreDecodingErrors,
		"ignore-decode-errors",
		false,
		"Ignore the decoding errors when debugging",
	)

	flag.BoolVar(
		&config.NoCPULimit,
		"no-limit",
		false,
		"Use all available CPU cores",
	)

}

func (config *config) setupCommonFlags() {
	flag.BoolVar(
		&config.DisableUpload,
		"d",
		false,
		"If specified no attempts will be made to upload data to remote server.",
	)

	flag.BoolVar(
		&config.DisableMarket,
		"no-market",
		false,
		"Skip decoding and uploading market data (orders/history/notifications) to reduce CPU usage on weak machines.",
	)

	flag.StringVar(
		&config.BanditWebhookURL,
		"bandit-webhook",
		DefaultWBGIngestURL,
		"Full URL to POST bandit event JSON to.",
	)

	flag.StringVar(
		&config.ListenDevices,
		"l",
		"",
		"Listen on this comma separated devices instead of all available. (Windows: Use MAC-Address, Linux: Use interface name)",
	)

	flag.StringVar(
		&config.OfflinePath,
		"o",
		"",
		"Parses a local file instead of checking albion ports.",
	)

	flag.BoolVar(
		&config.Minimize,
		"minimize",
		false,
		"Automatically minimize the window.",
	)

	flag.StringVar(
		&config.PublicIngestBaseUrls,
		"i",
		"https+pow://albion-online-data.com",
		"Base URL to send PUBLIC data to, can be 'nats://', 'http://', 'https://' or 'noop' and can have multiple uploaders. Comma separated.",
	)

	flag.StringVar(
		&config.RecordPath,
		"record",
		"",
		"Enable recording commands to a file for debugging later.",
	)
}

func (config *config) setupLogs() {
	if config.Debug {
		config.LogLevel = "DEBUG"
	}
	if config.Trace {
		config.LogLevel = "TRACE"
	}

	level, err := logrus.ParseLevel(strings.ToLower(config.LogLevel))
	if err != nil {
		log.Errorf("Error getting level: %v", err)
	}

	log.SetLevel(level)

	// Rotate existing log files before creating new one
	rotateLogFiles()

	// Always log to both file and terminal
	// Use colors for terminal, strip ANSI codes for file
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, DisableSorting: true, ForceColors: true})
	f, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err == nil {
		// Wrap file writer to strip ANSI codes
		strippedFileWriter := newAnsiStripWriter(f)
		multiWriter := io.MultiWriter(colorable.NewColorableStdout(), strippedFileWriter)
		log.SetOutput(multiWriter)
	} else {
		log.SetOutput(colorable.NewColorableStdout())
		log.Warnf("Could not create log file: %v", err)
	}
}

// rotateLogFiles moves the current log file to a numbered backup and removes old backups
func rotateLogFiles() {
	// Check if current log file exists
	if _, err := os.Stat(logFileName); os.IsNotExist(err) {
		return // No log file to rotate
	}

	// Remove the oldest log file if we're at the limit
	oldestLog := fmt.Sprintf("%s.%d", logFileName, maxLogFiles)
	_ = os.Remove(oldestLog)

	// Shift all existing log files up by one number
	for i := maxLogFiles - 1; i >= 1; i-- {
		oldName := fmt.Sprintf("%s.%d", logFileName, i)
		newName := fmt.Sprintf("%s.%d", logFileName, i+1)
		_ = os.Rename(oldName, newName)
	}

	// Rename current log file to .1
	_ = os.Rename(logFileName, fmt.Sprintf("%s.1", logFileName))
}

// GetLogFilePath returns the full path to the current log file
func GetLogFilePath() string {
	absPath, err := filepath.Abs(logFileName)
	if err != nil {
		return logFileName
	}
	return absPath
}

func (config *config) setupDebugEvents() {
	config.DebugEvents = make(map[int]bool)
	if config.DebugEventsString != "" {
		for _, event := range strings.Split(config.DebugEventsString, ",") {
			number, err := strconv.Atoi(event)
			if err == nil {
				config.DebugEvents[number] = true
			}
		}
	}
	if config.DebugEventsBlacklistString != "" {
		for _, event := range strings.Split(config.DebugEventsBlacklistString, ",") {
			number, err := strconv.Atoi(event)
			if err == nil {
				config.DebugEvents[number] = false
			}
		}
	}

	// Looping through map keys is purposefully random by design in Go
	for number, shouldDebug := range config.DebugEvents {
		verb := "Ignoring"
		if shouldDebug {
			verb = "Showing"
		}
		log.Debugf("[%v] event: [%v]%v", verb, number, EventType(number))
	}

}

func (config *config) setupDebugOperations() {
	config.DebugOperations = make(map[int]bool)
	if config.DebugOperationsString != "" {
		for _, operation := range strings.Split(config.DebugOperationsString, ",") {
			number, err := strconv.Atoi(operation)
			if err == nil {
				config.DebugOperations[number] = true
			}
		}
	}

	if config.DebugOperationsBlacklistString != "" {
		for _, operation := range strings.Split(config.DebugOperationsBlacklistString, ",") {
			number, err := strconv.Atoi(operation)
			if err == nil {
				config.DebugOperations[number] = false
			}
		}
	}

	// Looping through map keys is purposefully random by design in Go
	for number, shouldDebug := range config.DebugOperations {
		verb := "Ignoring"
		if shouldDebug {
			verb = "Showing"
		}
		log.Debugf("[%v] operation: [%v]%v", verb, number, OperationType(number))
	}

}

// PersistDisableMarket writes the DisableMarket setting to config.yaml so
// the dashboard's market toggle survives a restart. Creates the file if
// it does not exist yet.
func PersistDisableMarket(disabled bool) {
	viper.Set("DisableMarket", disabled)

	path := "config.yaml"
	line := fmt.Sprintf("DisableMarket: %v", disabled)

	data, err := os.ReadFile(path)
	if err != nil {
		if err := os.WriteFile(path, []byte(line+"\n"), 0o644); err != nil {
			log.Errorf("Could not persist DisableMarket: %v", err)
			return
		}
		log.Infof("Saved DisableMarket=%v to %s", disabled, path)
		return
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "DisableMarket:") {
			lines[i] = line
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, line)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		log.Errorf("Could not persist DisableMarket: %v", err)
		return
	}
	log.Infof("Saved DisableMarket=%v to %s", disabled, path)
}
