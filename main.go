package main

import (
	"log/slog"
	"os"

	"github.com/mr-stringer/punkbot/bot"
	"github.com/mr-stringer/punkbot/config"
	"github.com/mr-stringer/punkbot/global"
	"github.com/mr-stringer/punkbot/postoffice"
)

func main() {

	logLevel, configPath, err := config.ProcessFlags()
	if err != nil {
		slog.Error("Failed to process command line flags", "error", err.Error())
		os.Exit(global.ExitCmdLineArgsFailure)
	}

	/*Configure log level*/
	lvl := new(slog.LevelVar)
	lvl.Set(logLevel)

	//TODO support more log output options, right now only supports stdout
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	}))

	slog.SetDefault(logger)

	/*Source config*/
	cnf, err := config.GetConfig(configPath)
	if err != nil {
		slog.Error("Error getting config", "err", err.Error())
		os.Exit(global.ExitConfigFailure)
	}

	/*Test the post office */
	err = postoffice.PreFlightCheck(cnf)
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Info("Starting the bot")
	err = bot.Start(cnf)
	if err != nil {
		os.Exit(global.ExitBotFailure)
	}

	slog.Info("Shutdown complete")

}
