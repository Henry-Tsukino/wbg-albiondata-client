package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ConsoleHidden bool `json:"console_hidden"`
	// новые настройки сюда
}

var Global = &Config{}

func configPath() (string, error) {
	// Windows — Local AppData
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		return filepath.Join(localAppData, "WBGAlbionClient", "config.json"), nil
	}
	// Linux/macOS — XDG (~/.config)
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config dir: %w", err)
	}
	return filepath.Join(configDir, "WBGAlbionClient", "config.json"), nil
}

func Load() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil // первый запуск — используем дефолты
	} else if err != nil {
		return err
	}

	return json.Unmarshal(data, Global)
}

func Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(Global, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}