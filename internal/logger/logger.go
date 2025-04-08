// Package logger provides utilities for creating and managing loggers.
package logger

import (
	"go.uber.org/zap"
)

// NewLogger creates and returns a new instance of a logger.
func NewLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
