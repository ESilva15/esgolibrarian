package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	filename      string
	consoleLogger *logrus.Logger
	fileLogger    *logrus.Logger
}

var (
	defaultLogPath = "/golibrarian.log"
)

var (
	logger *Logger
	once   sync.Once
)

func createLogger(fpath string) *Logger {
	file, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		panic(err)
	}

	consoleLogger := logrus.New()
	consoleLogger.SetOutput(os.Stdout)
	consoleLogger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	fileLogger := logrus.New()
	fileLogger.SetOutput(file)
	fileLogger.SetFormatter(&logrus.JSONFormatter{})

	return &Logger{
		filename:      fpath,
		consoleLogger: consoleLogger,
		fileLogger:    fileLogger,
	}
}

func getInstance() *Logger {
	once.Do(func() {
		logger = createLogger(defaultLogPath)
	})

	return logger
}

// SetOptions I will create a better way to define some options here
// for now this will definitely do tho
func SetOptions(consoleOut bool) {
	logger = getInstance()

	logger.consoleLogger.SetOutput(io.Discard)
}

func SetLogpath(dir string) {
	defaultLogPath = filepath.Join(dir, defaultLogPath)
}
