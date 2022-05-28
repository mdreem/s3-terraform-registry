package logger

import "go.uber.org/zap"

var Logger *zap.Logger
var sugar *zap.SugaredLogger

func init() {
	Logger, _ = zap.NewProduction()
	sugar = Logger.Sugar()
}

func Info(message string, fields ...interface{}) {
	sugar.Infow(message, fields...)
}

func Debug(message string, fields ...interface{}) {
	sugar.Debugw(message, fields...)
}

func Error(message string, fields ...interface{}) {
	sugar.Errorw(message, fields...)
}

func Fatal(message string, fields ...interface{}) {
	sugar.Fatalw(message, fields...)
}
