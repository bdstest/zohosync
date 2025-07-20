// Package utils provides utility functions for ZohoSync
package utils

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// InitLogger initializes the application logger
func InitLogger(level string) *logrus.Logger {
	if log != nil {
		return log
	}

	log = logrus.New()
	
	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)
	
	// Set formatter
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	
	// Create log directory
	logDir := filepath.Join(os.Getenv("HOME"), ".config", "zohosync", "logs")
	if err := os.MkdirAll(logDir, 0755); err == nil {
		logFile := filepath.Join(logDir, "zohosync.log")
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		}
	}
	
	return log
}

// GetLogger returns the application logger
func GetLogger() *logrus.Logger {
	if log == nil {
		return InitLogger("info")
	}
	return log
}
