//go:build linux

package systray

import (
	"github.com/ao-data/albiondata-client/config"
	"github.com/ao-data/albiondata-client/log"
)

const CanHideConsole = false

func HideConsole() {}
func ShowConsole()  {}

func Run() {
	if err := config.Load(); err != nil {
		log.Errorf("Failed to load config: %v", err)
	}
}