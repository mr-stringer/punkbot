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

	cl, err := config.ProcessFlags()
	if err != nil {
		slog.Error("Failed to process command line flags", "error", err.Error())
		os.Exit(global.ExitCmdLineArgsFailure)
	}

	//configure the logger
	err = loggerConfig(cl)
	if err != nil {
		slog.Error("Failed to configure the logger", "err", err.Error())
		os.Exit(global.ExitConfigFailure)
	}

	/*Source config*/
	cnf, err := config.GetConfig(cl.ConfigFilePath)
	if err != nil {
		slog.Error("Error getting config", "err", err.Error())
		os.Exit(global.ExitConfigFailure)
	}

	slog.Info(cnf.GetSecret())

	/* initialise the channel package */
	cp := global.ChanPkg{
		ByteSlice:  make(chan []byte, global.ByteSliceBufferSize),
		Cancel:     make(chan bool),
		ReqDidResp: make(chan bool),
		DIDResp:    make(chan global.DIDResponse),
	}

	/*Start the DID Response server*/
	go postoffice.DIDResponseServer(cnf, cp)

	slog.Info("Starting the bot")
	err = bot.Start(cnf, cp)
	if err != nil {
		os.Exit(global.ExitBotFailure)
	}

	slog.Info("Shutdown complete")

}

func loggerConfig(cl *config.ClArgs) error {
	var lf *os.File = nil
	var err error

	/*By default, we log to the console, check to see if file is specified*/
	if cl.LogPath != "" {
		/*Open file for append*/
		lf, err = os.OpenFile(cl.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0640)
		if err != nil {
			//promote the error
			return err
		}
	}

	ho := slog.HandlerOptions{
		Level: cl.LogLevel,
	}

	if cl.JsonLog { /*If logging to JSON, log to JSON*/
		if lf == nil {
			jcl := slog.New(slog.NewJSONHandler(os.Stdout, &ho))
			slog.Info("Setting JSON logger stdout")
			slog.SetDefault(jcl)
			slog.Info("Setting JSON logger stdout")
		} else {
			jfl := slog.New(slog.NewJSONHandler(lf, &ho))
			slog.Info("Setting JSON logger to file")
			slog.SetDefault(jfl)
			slog.Info("Setting JSON logger to file")
		}
	} else { /*If not JSON, do text logging*/
		if lf == nil {
			tcl := slog.New(slog.NewTextHandler(os.Stdout, &ho))
			slog.Info("Setting text logger to stdout")
			slog.SetDefault(tcl)
			slog.Info("Setting text logger to stdout")
		} else {
			rfl := slog.New(slog.NewTextHandler(lf, &ho))
			slog.Info("Setting text logger to file")
			slog.SetDefault(rfl)
			slog.Info("Setting text logger to file")
		}
	}
	return nil
}
