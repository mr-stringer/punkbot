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
func ProcessFlags() (slog.Level, string, error) {
	sl := slog.LevelInfo
	var ver bool

	logLevelPtr := flag.String("l", "info", "used to set the logging level, may be err, warn, info or debug")
	configFilePtr := flag.String("f", "./botcnf.yml or ./botcnf.json", "specifies the location of the configuration file")
	flag.BoolVar(&ver, "v", false, "prints the version and quit")
	flag.Parse()

	switch *logLevelPtr {
	case "err":
		sl = slog.LevelError
	case "warn":
		sl = slog.LevelWarn
	case "info":
		sl = slog.LevelInfo
	case "debug":
		sl = slog.LevelDebug
	default:
		return 0, "", fmt.Errorf("log level '%s' not supported", *logLevelPtr)
	}

	if ver {
		fmt.Printf("Version:\t%s\n", global.ReleaseVersion)
		fmt.Printf("BuildTime:\t%s\n", global.BuildTime)
		os.Exit(0)
	}

	return sl, *configFilePtr, nil
}
