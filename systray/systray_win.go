//go:build windows

package systray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ao-data/albiondata-client/client"
	"github.com/ao-data/albiondata-client/icon"
	"github.com/ao-data/albiondata-client/log"
	"github.com/getlantern/systray"
	"github.com/gonutz/w32"
)

var consoleHidden bool

func hideConsole() {
	console := w32.GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_HIDE)
		}
	}

	consoleHidden = true
}

func showConsole() {
	console := w32.GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_SHOW)
		}
	}

	consoleHidden = false
}

func GetActionTitle() string {
	if consoleHidden {
		return "Show Console"
	} else {
		return "Hide Console"
	}
}

func Run() {
	systray.Run(onReady, onExit)
}

func onExit() {

}

// compute the application display name. falls back to basename of the executable
func appName() string {
	if client.ConfigGlobal.AppName != "" {
		return client.ConfigGlobal.AppName
	}
	return filepath.Base(os.Args[0])
}

// registry key name should not contain spaces or punctuation
func regKeyName() string {
	name := appName()
	name = strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' || r == '.' {
			return -1
		}
		return r
	}, name)
	return name
}

// pick icon bytes, external file takes precedence
func trayIconData() []byte {
	if client.ConfigGlobal.TrayIconPath != "" {
		if data, err := os.ReadFile(client.ConfigGlobal.TrayIconPath); err == nil {
			return data
		} else {
			log.Errorf("Unable to load tray icon %s: %v", client.ConfigGlobal.TrayIconPath, err)
		}
	}
	return icon.Data
}

func onReady() {
	// Don't hide the console automatically
	// Unless started from the scheduled task or with the parameter
	// People think it is crashing
	if client.ConfigGlobal.Minimize {
		hideConsole()
	}
	systray.SetIcon(trayIconData())
	systray.SetTitle(appName())
	systray.SetTooltip(appName())
	mConHideShow := systray.AddMenuItem(GetActionTitle(), "Show/Hide Console")

	// Start on Windows checkbox - with safe initialization
	startChecked := false
	if isStartOnBootSafe() {
		startChecked = true
		log.Info("Auto-start is currently enabled")
	} else {
		log.Info("Auto-start is currently disabled")
	}
	mStartOnBoot := systray.AddMenuItemCheckbox("Start on Windows", "Launch on Windows logon", startChecked)
	mQuit := systray.AddMenuItem("Quit", "Close the Application")

	func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				log.Info("Quit requested")
				systray.Quit()
				os.Exit(0)

			case <-mConHideShow.ClickedCh:
				if consoleHidden == true {
					showConsole()
					mConHideShow.SetTitle(GetActionTitle())
				} else {
					hideConsole()
					mConHideShow.SetTitle(GetActionTitle())
				}
			case <-mStartOnBoot.ClickedCh:
				currentState := mStartOnBoot.Checked()
				newState := !currentState

				if newState {
					mStartOnBoot.Check()
					err := setStartOnBoot(true)
					if err != nil {
						log.Errorf("Failed to enable auto-start: %v", err)
						mStartOnBoot.Uncheck()
					} else {
						log.Info("Auto-start enabled successfully")
					}
				} else {
					mStartOnBoot.Uncheck()
					err := setStartOnBoot(false)
					if err != nil {
						log.Errorf("Failed to disable auto-start: %v", err)
						mStartOnBoot.Check()
					} else {
						log.Info("Auto-start disabled successfully")
					}
				}
			}
		}
	}()
}

func isStartOnBootSafe() bool {
	// Safe wrapper that never panics
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic in isStartOnBootSafe: %v", r)
		}
	}()
	return isStartOnBoot()
}

func isStartOnBoot() bool {
	exe, err := os.Executable()
	if err != nil {
		log.Debugf("Error getting executable path: %v", err)
		return false
	}

	// Use PowerShell to check registry
	key := regKeyName()
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		fmt.Sprintf(`(Get-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Run' -Name '%s' -ErrorAction SilentlyContinue).%s`, key, key))

	output, err := cmd.Output()
	if err != nil {
		log.Debugf("Registry check failed: %v", err)
		return false
	}

	value := strings.TrimSpace(string(output))
	if value == "" {
		return false
	}

	// Check if the stored value contains our exe path
	return strings.Contains(value, exe)
}

func setStartOnBoot(enable bool) error {
	exe, err := os.Executable()
	if err != nil {
		log.Errorf("Error getting executable path: %v", err)
		return err
	}

	if enable {
		// Store the path to the executable with -minimize flag
		key := regKeyName()
		cmdStr := fmt.Sprintf(`Set-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Run' -Name '%s' -Value '"%s" -minimize' -Force`, key, exe)
		cmd := exec.Command("powershell", "-NoProfile", "-Command", cmdStr)

		err := cmd.Run()
		if err != nil {
			log.Errorf("Error enabling auto-start: %v", err)
			return err
		}
		log.Infof("Auto-start enabled")
	} else {
		// Delete the registry value
		key := regKeyName()
		cmdStr := fmt.Sprintf(`Remove-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Run' -Name '%s' -ErrorAction SilentlyContinue`, key)
		cmd := exec.Command("powershell", "-NoProfile", "-Command", cmdStr)

		err := cmd.Run()
		if err != nil {
			log.Debugf("Error disabling auto-start (may already be disabled): %v", err)
		}
		log.Info("Auto-start disabled")
	}

	return nil
}
