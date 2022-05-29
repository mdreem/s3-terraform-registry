package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func init() {
	Logger, _ = zap.NewProduction()
	Sugar = Logger.Sugar()
}

func Info(message string, fields ...interface{}) {
	Sugar.Infow(message, fields...)
}

func Debug(message string, fields ...interface{}) {
	Sugar.Debugw(message, fields...)
}

func Error(message string, fields ...interface{}) {
	Sugar.Errorw(message, fields...)
}

func Fatal(message string, fields ...interface{}) {
	Sugar.Fatalw(message, fields...)
}
