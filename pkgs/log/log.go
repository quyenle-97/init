package log

import (
	"context"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"os"
)

type MultiLogger struct {
	*logrus.Logger
	kafkaHook *logrus.Hook
}

func NewMultiLogger(LogLevel logrus.Level) (*MultiLogger, error) {
	logger := logrus.New()
	logger.SetLevel(LogLevel)

	multiLogger := &MultiLogger{
		Logger: logger,
	}

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	return multiLogger, nil
}

func (l *MultiLogger) Handle(_ context.Context, err error) {
	var fieldError validator.ValidationErrors
	switch {
	default:
		l.Error(err)
	case errors.As(err, &fieldError):
		l.Info(err.Error())
	}
}
