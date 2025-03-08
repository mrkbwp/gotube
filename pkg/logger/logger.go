package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

// Logger представляет интерфейс для логгирования
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	With(key string, value interface{}) Logger
}

// loggerImpl реализация логгера на основе zap
type loggerImpl struct {
	logger *zap.SugaredLogger
}

var (
	instance Logger
	once     sync.Once
)

// NewLogger создает и настраивает логгер
func NewLogger(debug bool) Logger {
	once.Do(func() {
		// Определяем уровень логирования
		level := zapcore.InfoLevel
		if debug {
			level = zapcore.DebugLevel
		}

		// Настраиваем кодировщик
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// Создаем core для логгера
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)

		// Создаем логгер
		logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		instance = &loggerImpl{
			logger: logger.Sugar(),
		}
	})

	return instance
}

// Debug выводит отладочное сообщение
func (l *loggerImpl) Debug(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
}

// Info выводит информационное сообщение
func (l *loggerImpl) Info(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

// Warn выводит предупреждение
func (l *loggerImpl) Warn(msg string, args ...interface{}) {
	l.logger.Warnf(msg, args...)
}

// Error выводит сообщение об ошибке
func (l *loggerImpl) Error(msg string, args ...interface{}) {
	l.logger.Errorf(msg, args...)
}

// Fatal выводит критическую ошибку и завершает приложение
func (l *loggerImpl) Fatal(msg string, args ...interface{}) {
	l.logger.Fatalf(msg, args...)
}

// With добавляет контекстные поля к логгеру
func (l *loggerImpl) With(key string, value interface{}) Logger {
	return &loggerImpl{
		logger: l.logger.With(key, value),
	}
}

// Log возвращает глобальный экземпляр логгера
func Log() Logger {
	if instance == nil {
		return NewLogger(false)
	}
	return instance
}
