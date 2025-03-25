package logger

import (
	"go.uber.org/zap"
)

// NewLogger создаёт и возвращает новый экземпляр логгера.
func NewLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return logger.Sugar(), nil
}
