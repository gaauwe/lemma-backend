package config

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Apn struct {
		KeyId  string `toml:"key_id"`
		TeamId string `toml:"team_id"`
		Topic  string `toml:"topic"`
	} `toml:"apn"`

	Server struct {
		PollRate  int64  `toml:"poll_rate"`
		Token     string `toml:"token"`
		EnableSSL bool   `toml:"enable_ssl"`
		Domain    string `toml:"domain"`
		Email     string `toml:"email"`
		StartedAt time.Time
	} `toml:"server"`
}

var config *Config

func LoadConfig(filepath string) error {
	_, err := toml.DecodeFile(filepath, &config)
	if err != nil {
		return fmt.Errorf("Could not decode settings file %s: %w", filepath, err)
	}

	// Store at what time the server is started, so we never send notifications for events that happened earlier.
	config.Server.StartedAt = time.Now()

	return nil
}

func Get() *Config {
	return config
}
