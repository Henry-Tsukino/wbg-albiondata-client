//go:build windows

package systray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ao-data/albiondata-client/client"
	"github.com/ao-data/albiondata-client/config"
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
			wnd := w32.GetConsoleWindow()
			if wnd != 0 {
				currentStyle := w32.GetWindowLong(wnd, w32.GWL_EXSTYLE)
				w32.SetWindowLong(wnd, w32.GWL_EXSTYLE, uint32(currentStyle)|w32.WS_EX_TOOLWINDOW)
			}
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
	}
	return "Hide Console"
}

func Run() {
	systray.Run(onReady, onExit)
}

func onExit() {}

func appName() string {
	if client.ConfigGlobal.AppName != "" {
		return client.ConfigGlobal.AppName
	}
	return filepath.Base(os.Args[0])
}

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
	if err := config.Load(); err != nil {
		log.Errorf("Failed to load config: %v", err)
	}

	// -minimize флаг имеет приоритет, иначе восстанавливаем из конфига
	if client.ConfigGlobal.Minimize || config.Global.ConsoleHidden {
		hideConsole()
	}

	if client.ConfigGlobal.StartOnBoot {
		if !isStartOnBootSafe() {
			if err := setStartOnBoot(true); err != nil {
				log.Errorf("Failed to enable auto-start on startup: %v", err)
			} else {
				log.Info("Auto-start enabled from configuration")
			}
		}
	}

	systray.SetIcon(trayIconData())
	systray.SetTitle(appName())
	systray.SetTooltip(appName())
	mConHideShow := systray.AddMenuItem(GetActionTitle(), "Show/Hide Console")

	startChecked := false
	if isStartOnBootSafe() {
		startChecked = true
		log.Info("Auto-start is currently enabled")
		client.UpdateAutoStartStatus("Enabled")
	} else {
		log.Info("Auto-start is currently disabled")
		client.UpdateAutoStartStatus("Disabled")
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
				if consoleHidden {
					showConsole()
				} else {
					hideConsole()
				}
				config.Global.ConsoleHidden = consoleHidden
				if err := config.Save(); err != nil {
					log.Errorf("Failed to save config: %v", err)
				}
				mConHideShow.SetTitle(GetActionTitle())

			case <-mStartOnBoot.ClickedCh:
				currentState := mStartOnBoot.Checked()
				newState := !currentState

				if newState {
					mStartOnBoot.Check()
					if err := setStartOnBoot(true); err != nil {
						log.Errorf("Failed to enable auto-start: %v", err)
						mStartOnBoot.Uncheck()
					} else {
						log.Info("Auto-start enabled successfully")
						client.UpdateAutoStartStatus("Enabled")
					}
				} else {
					mStartOnBoot.Uncheck()
					if err := setStartOnBoot(false); err != nil {
						log.Errorf("Failed to disable auto-start: %v", err)
						mStartOnBoot.Check()
					} else {
						log.Info("Auto-start disabled successfully")
						client.UpdateAutoStartStatus("Disabled")
					}
				}
			}
		}
	}()
}

func isStartOnBootSafe() bool {
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

	key := "WBGAlbionClient"
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

	return strings.Contains(value, exe)
}

func setStartOnBoot(enable bool) error {
	exe, err := os.Executable()
	if err != nil {
		log.Errorf("Error getting executable path: %v", err)
		return err
	}

	key := "WBGAlbionClient"

	if enable {
		cmdStr := fmt.Sprintf(`Set-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Run' -Name '%s' -Value '"%s" -minimize' -Force`, key, exe)
		if err := exec.Command("powershell", "-NoProfile", "-Command", cmdStr).Run(); err != nil {
			log.Errorf("Error enabling auto-start: %v", err)
			return err
		}
		log.Info("Auto-start enabled")
	} else {
		cmdStr := fmt.Sprintf(`Remove-ItemProperty -Path 'HKCU:\\Software\\Microsoft\\Windows\\CurrentVersion\\Run' -Name '%s' -ErrorAction SilentlyContinue`, key)
		if err := exec.Command("powershell", "-NoProfile", "-Command", cmdStr).Run(); err != nil {
			log.Debugf("Error disabling auto-start (may already be disabled): %v", err)
		}
		log.Info("Auto-start disabled")
	}

	return nil
}