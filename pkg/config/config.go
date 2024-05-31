package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

const (
	configDir      = ".ksbuilder"
	configFilename = "config.json"
)

type Config struct {
	Token  []byte `json:"token"`
	Server string `json:"server"`
}

func Write(token []byte, server string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	root := filepath.Join(home, configDir)
	if _, err = os.Stat(root); os.IsNotExist(err) {
		if err = os.Mkdir(root, 0755); err != nil {
			return err
		}
	}
	config := &Config{
		Token:  token,
		Server: server,
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, configFilename), data, 0644)
}

func Read() (*Config, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filepath.Join(home, configDir, configFilename))
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	if err = json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func Remove() error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(home, configDir, configFilename))
}
