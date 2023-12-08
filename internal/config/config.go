package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Apn struct {
		KeyId  string `toml:"key_id"`
		TeamId string `toml:"team_id"`
	} `toml:"apn"`

	Device struct {
		DeviceToken string `toml:"device_token"`
		Topic       string `toml:topic`
	} `toml:"device"`

	Lemmy struct {
		Server   string `toml:"server"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"lemmy"`
}

var config *Config

func LoadConfig(filepath string) error {
	_, err := toml.DecodeFile(filepath, &config)
	if err != nil {
		return fmt.Errorf("Could not decode settings file %s: %w", filepath, err)
	}

	return nil
}

func Get() *Config {
	return config
}
