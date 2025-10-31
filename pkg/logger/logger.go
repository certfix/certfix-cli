package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// InitLogger initializes the logger
func InitLogger(verbose bool) {
	log = logrus.New()

	// Set output to stdout
	log.SetOutput(os.Stdout)

	// Set log level
	if verbose {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

// GetLogger returns the logger instance
func GetLogger() *logrus.Logger {
	if log == nil {
		InitLogger(false)
	}
	return log
}
