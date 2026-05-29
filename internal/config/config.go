package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/edu292/gator/internal/database"

	"github.com/google/uuid"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DBUrl         string    `json:"db_url"`
	CurrentUserID uuid.UUID `json:"current_user_id,omitempty"`
}

type State struct {
	Cfg *Config
	DB  *database.Queries
}

func NewState(c *Config, db *database.Queries) *State {
	return &State{Cfg: c, DB: db}
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

func (c *Config) write() error {
	configFile, err := getConfigFile(true)
	if err != nil {
		return err
	}

	defer configFile.Close()

	json.NewEncoder(configFile).Encode(c)

	return nil
}

func (c *Config) SetUser(userID uuid.UUID) error {
	c.CurrentUserID = userID

	return c.write()
}

func (c *Config) Reset() error {
	c.CurrentUserID = uuid.Nil

	return c.write()
}
