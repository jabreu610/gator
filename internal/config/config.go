package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBURL       string `json:"db_url"`
	CurrentUser string `json:"current_user_name"`
}

func (c Config) SetUser(user string) error {
	c.CurrentUser = user
	if err := write(c); err != nil {
		return err
	}

	return nil
}

func Read() (Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func write(config Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	return nil
}

func getConfigFilePath() (string, error) {
	userDir, err := os.UserHomeDir()
	configPath := filepath.Join(userDir, configFileName)
	if err != nil {
		return "", err
	}

	return configPath, nil
}
