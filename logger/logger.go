package logger

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/threefoldtech/tf-pinning-service/config"
)

var Log *logrus.Logger

type Fields = logrus.Fields

func GetDefaultLogger() *logrus.Logger {
	// Log as JSON instead of the default ASCII formatter.
	if Log == nil {
		log := logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})

		// Output to stdout instead of the default stderr
		// Can be any io.Writer, see below for File example
		log.SetOutput(os.Stdout)

		// Only log the warning severity or above.
		log.SetLevel(logrus.Level(config.CFG.Server.LogLevel))
		Log = log
	}

	return Log
}
