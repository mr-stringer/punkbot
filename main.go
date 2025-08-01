package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
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

	/* Create the master context                                              */
	/* This context will handle all cancelling of the bot and the DID server  */
	/* Go routines created by these functions will inherit the context and    */
	/* cleanup should be more straight forward.                               */

	ctx, cancel := context.WithCancel(context.Background())

	/* This anonymous function will trigger the ctx's cancel function in the. */
	/* a instruct all active routines to close a soon as possible.            */
	/* Any committed work (such as token refreshes, likes and reposts, should */
	/* be finished.                                                           */
	go func() {
		slog.Info("Listening for cancel signals")
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		esig := <-sig
		slog.Warn("received shutdown signal", "value", esig)
		slog.Warn("Shutdown started")
		cancel()
	}()

	/* initialise the channel package */
	cp := ChanPkg{
		ByteSlice:      make(chan []byte, ByteSliceBufferSize),
		ReqDidResp:     make(chan bool),
		Session:        make(chan DIDResponse),
		JetStreamError: make(chan error),
	}

	/* Each go routine in increment the wait group */
	var wg sync.WaitGroup

	/* Start the Session Server server */
	wg.Add(1)
	go sessionServer(ctx, &wg, cnf, cp)

	/* Start the bot */
	wg.Add(1)
	slog.Info("Starting the bot")
	err = Start(ctx, &wg, cnf, cp)
	if err != nil {
		os.Exit(ExitBotFailure)
	}

	go func() {
		var jetstreamErrors int = 0
		for {
			select {
			case err := <-cp.JetStreamError: //block until error
				jetstreamErrors++
				slog.Error("Jetstream Error", "err", err.Error())
				if jetstreamErrors >= 10 {
					slog.Error("Jetstream Error count too high, attempting shutdown")
					cancel()
					/* Sometime, a clean shutdown doesn't work. If the we're. */
					/* if it takes longer than 60 seconds, go nuclear.        */
					time.Sleep(time.Second * 10)
					os.Exit(ExitWebSocketFailure)
				}
			case <-time.After(10 * time.Minute):
				if jetstreamErrors > 0 {
					slog.Error("Decrementing Jetstream error count", "current", jetstreamErrors, "new", jetstreamErrors-1)
					jetstreamErrors--
				}
			}
		}
	}()

	wg.Wait()

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
