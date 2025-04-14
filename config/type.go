package config

import (
	"log/slog"
	"os"

	"github.com/mr-stringer/punkbot/global"
	"github.com/spf13/viper"
)

// The config type is the what the yaml/json config file is marshalled into. It
// consists of the bluesky identity "Identifier", a slice of "Hashtags"
// that the bot will re-post and the "JetStreamServer" you want to connect to.
// If the "JetStreamServer" argument is not configured, the bot will test the
// latency of the four public servers and connect to the one with the lowest
// latency.

type ClArgs struct {
	LogLevel       slog.Level
	ConfigFilePath string
	LogPath        string
	JsonLog        bool
}
type Config struct {
	Identifier string
	Terms      []string
	//Add jetstream instance, allowing users to set their preferred instance
	//Also add the ability to auto select public instance automatically based on
	//latency
	JetStreamServer string
	password        string //unexported

}

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
		slog.Error("failed to unmarshal config", "error", err.Error())
		return nil, err
	}

	err = cnf.GetSecretFromEnv()
	if err != nil {
		slog.Error("Cannot source password from the environment, PUNKBOT_PASSWORD is probably not set")
		return nil, err
	}

	/*If no JetSteam configured, find the public server with the lowest latency*/
	if cnf.JetStreamServer == "" {
		slog.Info("No JetStream server provided. Finding public server with lowest latency")
		err = cnf.FindFastestServer(MeasureLatency)
		if err != nil {
			slog.Error("No usable server found, cannot continue")
			os.Exit(global.ExitJetStreamFailure)
		}
	}
	return &cnf, nil
}
