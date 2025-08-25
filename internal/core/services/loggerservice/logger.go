package loggerservice

import (
	"log/slog"
	"os"
)

type LoggerService struct {
	logger *slog.Logger
}

func NewLoggerService() *LoggerService {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	return &LoggerService{logger: logger}
}

func (s *LoggerService) Info(msg string, args ...any) {
	s.logger.Info(msg, args...)
}

func (s *LoggerService) Error(msg string, args ...any) {
	s.logger.Error(msg, args...)
}

func (s *LoggerService) Debug(msg string, args ...any) {
	s.logger.Debug(msg, args...)
}

func (s *LoggerService) Warn(msg string, args ...any) {
	s.logger.Warn(msg, args...)
}
