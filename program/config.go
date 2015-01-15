package program

import (
	"github.com/BurntSushi/toml"
	"github.com/uswitch/dagr/schedule"
	"os"
)

type Config struct {
	Schedule  schedule.Schedule
	Immediate bool
}

var defaultConfig = Config{
	schedule.DefaultSchedule,
	false,
}

func readConfig(configPath string) (*Config, error) {
	var config = defaultConfig

	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, &config); err != nil {
			return nil, err
		}
	}

	return &config, nil
}
