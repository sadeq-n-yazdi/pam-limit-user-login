package logger

import (
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger
var lastLogLevel logrus.Level = logrus.InfoLevel

func GetLogger() *logrus.Logger {
	if logger == nil {
		logger = logrus.New()

		// Set the desired log level, for example, Debug, Info, Warn, Error, Fatal)
		logger.SetLevel(lastLogLevel)

		// Define colours for log levels if you're using a terminal.
		formatter := &logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		}
		// Attach the formatter to the logger
		logger.SetFormatter(formatter)
	}

	return logger
}
