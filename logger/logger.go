package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(lumlog *lumberjack.Logger, isProduction bool, console bool) *zap.Logger {

	var config zapcore.EncoderConfig

	if isProduction {
		config = zap.NewProductionEncoderConfig()
	} else {
		config = zap.NewDevelopmentEncoderConfig()
	}
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	var zapCores []zapcore.Core

	fileWriter := zapcore.AddSync(lumlog)
	fileEncoder := zapcore.NewJSONEncoder(config)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, zap.InfoLevel)
	zapCores = append(zapCores, fileCore)

	if console {
		consoleEncoder := zapcore.NewConsoleEncoder(config)
		consoleWriter := zapcore.AddSync(os.Stdout)
		consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, zap.InfoLevel)
		zapCores = append(zapCores, consoleCore)
	}

	core := zapcore.NewTee(zapCores...)

	return zap.New(core)
}
