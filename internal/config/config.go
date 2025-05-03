package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

// Config defines the main storage of service's configuration settings.
type Config struct {
	// Socket defines the socket that will be used for starting the GRPC-server.
	Socket string
}

func New(opts ...ConfigOpt) (Config, error) {
	const op = "config.New"

	conf := Config{}
	godotenv.Load("../../.env")

	for _, opt := range opts {
		if err := opt(&conf); err != nil {
			return Config{}, fmt.Errorf("error of the %s: %s", op, err)
		}
	}

	return conf, nil
}
