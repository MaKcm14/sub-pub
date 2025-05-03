package config

import (
	"fmt"
	"os"
)

// ConfigOpt defines the func of options' configuration.
type ConfigOpt func(conf *Config) error

// getEnv gets the configuration of .env vars.
func getEnv(name string) (string, error) {
	const op = "config.getEnv"

	res := os.Getenv(name)

	if len(res) == 0 {
		return "", fmt.Errorf("error of the %s: the ENV '%s' is empty or doesn't set", op, name)
	}
	return res, nil
}

// ConfigSocket defines the SOCKET var configuration.
func ConfigSocket(conf *Config) error {
	socket, err := getEnv("SOCKET")

	if err != nil {
		return err
	}
	conf.Socket = socket

	return nil
}
