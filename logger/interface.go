package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

func Println(lvl logrus.Level, v ...any) {
	logger := getInstance()

	logger.fileLogger.Log(lvl, v...)
	logger.consoleLogger.Log(lvl, v...)
}

func Printf(lvl logrus.Level, format string, v ...any) {
	logger := getInstance()

	logger.fileLogger.Logf(lvl, format, v...)
	logger.consoleLogger.Logf(lvl, format, v...)
}

func PrintJson(lvl logrus.Level, data map[string]any, msg string) {
	logger := getInstance()

	logger.consoleLogger.WithFields(data).Info(msg)
	logger.fileLogger.WithFields(data).Info(msg)
}

// NoLogf will only print to the console
func NoLogf(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
}
