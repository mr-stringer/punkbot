package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/mr-stringer/punkbot/global"
)

// Process is the function that handles the command line flags. The flags
// processed are: -l, log level and -f, config file location.
//
// log level may be set to err, warn, info or debug.
//
// the file is set to a string which must be the path to a valid config file
func ProcessFlags() (*ClArgs, error) {
	cl := ClArgs{}
	var ver bool
	var json bool
	var debugPost bool

	logLevelPtr := flag.String("l", "info", "used to set the logging level, may be err, warn, info or debug")
	configFilePathPtr := flag.String("f", "", "specifies the location of the configuration file")
	flag.BoolVar(&ver, "v", false, "prints the version and quits")
	logPathPtr := flag.String("o", "", "specifies the path of the log file, default logging happens in the console")
	flag.BoolVar(&json, "j", false, "sets json log format")
	flag.BoolVar(&debugPost, "debugPosts", false, "if enabled, debug logging will include all scanned posts - very noisy!")
	flag.Parse()

	cl.JsonLog = json
	cl.ConfigFilePath = *configFilePathPtr
	cl.LogPath = *logPathPtr

	switch *logLevelPtr {
	case "err":
		cl.LogLevel = slog.LevelError
		global.LogLevel = slog.LevelError
	case "warn":
		cl.LogLevel = slog.LevelWarn
		global.LogLevel = slog.LevelWarn
	case "info":
		cl.LogLevel = slog.LevelInfo
		global.LogLevel = slog.LevelInfo
	case "debug":
		cl.LogLevel = slog.LevelDebug
		global.LogLevel = slog.LevelDebug
	default:
		return nil, fmt.Errorf("log level '%s' not supported", *logLevelPtr)
	}

	if ver {
		fmt.Printf("Version:\t%s\n", global.ReleaseVersion)
		fmt.Printf("BuildTime:\t%s\n", global.BuildTime)
		os.Exit(0)
	}

	return &cl, nil
}
