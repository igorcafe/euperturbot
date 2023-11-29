package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	GodID     int64
	BotToken  string
	OpenAIKey string
}

func Load() (c Config, err error) {
	b, err := os.ReadFile("config.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &c)
	return
}
