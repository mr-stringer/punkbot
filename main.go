package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
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

	/* configure the logger */
	err = loggerConfig(cl)
	if err != nil {
		slog.Error("Failed to configure the logger", "err", err.Error())
		os.Exit(ExitConfigFailure)
	}

	/* Log current version */
	slog.Info("Version information", "Version", ReleaseVersion, "BuildTime", BuildTime)

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

	/* This anonymous function will trigger the ctx's cancel function         */
	/* this instructs all active routines to close a soon as possible.        */
	/* Any committed work (such as token refreshes, likes and reposts, should */
	/* be finished.                                                           */
	go func() {
		slog.Info("Listening for cancel signals")
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		esig := <-sig
		slog.Warn("Received shutdown signal", "value", esig)
		slog.Warn("Shutdown started")
		cancel()
	}()

	/* initialise the channel package */
	cp := ChanPkg{
		ByteSlice:      make(chan []byte, ByteSliceBufferSize),
		ReqDidResp:     make(chan bool),
		Session:        make(chan DIDResponse),
		JetStreamError: make(chan bool),
		Exit:           make(chan int),
	}

	/* Each go routine in increment the wait group */
	var wg sync.WaitGroup

	/* Start the Session Server server */
	wg.Add(1)
	tm := &TokenManager{}
	go sessionServer(tm, ctx, &wg, cnf, cp, time.Second*60)

	/* Start the bot */
	wg.Add(1)
	slog.Info("Starting the bot")
	err = bot(ctx, &wg, cnf, cp)
	if err != nil {
		os.Exit(ExitBotFailure)
	}

	var i int = 0

	go func() {
		select {
		case <-cp.JetStreamError: //block until signal
			slog.Error("Jetstream Error, cannot continue")
			slog.Warn("Shutdown started")
			cancel()
			return
		case i = <-cp.Exit:
			slog.Error("Exit requested, shutting down")
			cancel()
			slog.Info("Shutdown complete")
		}

	}()

	wg.Wait()
	slog.Info("Shutdown complete")
	os.Exit(i)
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
		Level:     cl.LogLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				// Get the source information
				source, ok := a.Value.Any().(*slog.Source)
				if !ok {
					return a
				}

				// Extract function name from the source
				fullFunc := source.Function
				// Split and get the last part of the function name
				funcName := fullFunc[strings.LastIndexByte(fullFunc, '.')+1:]

				// Return a new attribute with just the function name
				return slog.Attr{Key: "function", Value: slog.StringValue(funcName)}
			}
			return a
		},
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
