package main

import (
	"log/slog"
	"os"
)

func main() {

	cl, err := ProcessFlags()
	if err != nil {
		slog.Error("Failed to process command line flags", "error", err.Error())
		os.Exit(ExitCmdLineArgsFailure)
	}

	//configure the logger
	err = loggerConfig(cl)
	if err != nil {
		slog.Error("Failed to configure the logger", "err", err.Error())
		os.Exit(ExitConfigFailure)
	}

	/*Source config*/
	cnf, err := GetConfig(cl.ConfigFilePath)
	if err != nil {
		slog.Error("Error getting config", "err", err.Error())
		os.Exit(ExitConfigFailure)
	}

	slog.Info(cnf.GetSecret())

	/* initialise the channel package */
	cp := ChanPkg{
		ByteSlice:  make(chan []byte, ByteSliceBufferSize),
		Cancel:     make(chan bool),
		ReqDidResp: make(chan bool),
		DIDResp:    make(chan DIDResponse),
	}

	/*Start the DID Response server*/
	go DIDResponseServer(cnf, cp)

	slog.Info("Starting the bot")
	err = Start(cnf, cp)
	if err != nil {
		os.Exit(ExitBotFailure)
	}

	slog.Info("Shutdown complete")

}

func loggerConfig(cl *ClArgs) error {
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
