package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

// The config type is the what the yaml/json config file is marshalled into. It
// consists of the bluesky identity "Identifier", a slice of "Hashtags"
// that the bot will re-post and the "JetStreamServer" you want to connect to.
// If the "JetStreamServer" argument is not configured, the bot will test the
// latency of the four public servers and connect to the one with the lowest
// latency.

func GetConfig(path string) (*Config, error) {
	slog.Info("Attempting to read config")
	if path == "" {
		slog.Info("No config name specified, will attempt to load config from current directory")
		viper.SetConfigName("botcnf")
		viper.AddConfigPath("./")
	} else {
		slog.Info("Attempting to read config file", "file", path)
		viper.SetConfigFile(path)
	}
	err := viper.ReadInConfig()
	if err != nil {
		slog.Error("Failed to read in config")
		return nil, err
	}

	cnf := Config{}
	err = viper.Unmarshal(&cnf)
	if err != nil {
		slog.Error("Failed to unmarshal config", "error", err.Error())
		return nil, err
	}

	/* Ensure that the terms list is not empty */
	/* No more radiator mode */
	if len(cnf.Terms) < 1 {
		slog.Error("The configuration contains no terms. Ensure that 'Terms' is set in the config file")
		return nil, fmt.Errorf("noTermsInConfig")
	}

	err = cnf.GetSecretFromEnv()
	if err != nil {
		slog.Error("Cannot source password from the environment, PUNKBOT_PASSWORD is probably not set")
		return nil, err
	}

	/*If no JetSteam configured, find the public server with the lowest latency*/
	if cnf.JetStreamServer == "" {
		cnf.setAutoJetStream()
		slog.Info("No JetStream server provided. Will find public server with lowest latency")

	}
	return &cnf, nil
}
