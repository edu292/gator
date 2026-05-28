package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func getConfigFile(write ...bool) (*os.File, error) {
	w := false
	if len(write) > 0 {
		w = write[0]
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.New("user home dir not found")
	}

	var flag int
	if w {
		flag = os.O_TRUNC | os.O_WRONLY
	} else {
		flag = os.O_RDONLY
	}

	filePath := path.Join(homeDir, configFileName)
	configFile, err := os.OpenFile(filePath, flag, 0o644)
	if err != nil {
		return nil, errors.New("config file not found")
	}

	return configFile, nil
}

func Read() (*Config, error) {
	configFile, err := getConfigFile()
	if err != nil {
		return nil, err
	}

	defer configFile.Close()

	config := new(Config)
	err = json.NewDecoder(configFile).Decode(config)
	if err != nil {
		return nil, errors.New("error while parsing config file")
	}

	return config, nil
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName

	configFile, err := getConfigFile(true)
	if err != nil {
		return err
	}

	defer configFile.Close()

	json.NewEncoder(configFile).Encode(c)

	return nil
}
