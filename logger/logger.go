package logger

import (
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type CTXLogger struct {
	Logger *zap.Logger
}

var Logger = &CTXLogger{
	Logger: nil,
}

func New(lumlog *lumberjack.Logger, isProduction bool, console bool) *zap.Logger {

	var config zapcore.EncoderConfig

	if isProduction {
		config = zap.NewProductionEncoderConfig()
	} else {
		config = zap.NewDevelopmentEncoderConfig()
	}
	config.EncodeTime = zapcore.RFC3339TimeEncoder

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

func Init(zapLogger *zap.Logger) {
	Logger.Logger = zapLogger
}

func FieldWithCTX(ctx *gin.Context) *[]zapcore.Field {
	return &[]zapcore.Field{
		zap.Int("status", ctx.Writer.Status()),
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.Request.URL.Path),
		zap.String("query", ctx.Request.URL.RawQuery),
		zap.String("ip", ctx.ClientIP()),
	}
}

func (c *CTXLogger) Info(msg string, fields ...zap.Field) {
	c.Logger.Info(msg, fields...)
}

func (c *CTXLogger) InfoWithCTX(ctx *gin.Context, msg string, fields ...zap.Field) {
	c.Logger.With(*FieldWithCTX(ctx)...).Info(msg, fields...)
}

func (c *CTXLogger) Error(msg string, fields ...zap.Field) {
	c.Logger.Error(msg, fields...)
}

func (c *CTXLogger) ErrorWithCTX(ctx *gin.Context, msg string, err error, fields ...zap.Field) {
	if err != nil {
		c.Logger.With(*FieldWithCTX(ctx)...).With(zap.String("error", err.Error())).Error(msg, fields...)
		return
	}
	c.Logger.With(*FieldWithCTX(ctx)...).Error(msg, fields...)
}

func (c *CTXLogger) Fatal(msg string, fields ...zap.Field) {
	c.Logger.Fatal(msg, fields...)
}

func (c *CTXLogger) FatalWithCTX(ctx *gin.Context, msg string, fields ...zap.Field) {
	c.Logger.With(*FieldWithCTX(ctx)...).Fatal(msg, fields...)
}

func (c *CTXLogger) DPanic(msg string, fields ...zap.Field) {
	c.Logger.DPanic(msg, fields...)
}

func (c *CTXLogger) DPanicWithCTX(ctx *gin.Context, msg string, fields ...zap.Field) {
	c.Logger.With(*FieldWithCTX(ctx)...).DPanic(msg, fields...)
}

func (c *CTXLogger) Debug(msg string, fields ...zap.Field) {
	c.Logger.DPanic(msg, fields...)
}

func (c *CTXLogger) DebugWithCTX(ctx *gin.Context, msg string, fields ...zap.Field) {
	c.Logger.With(*FieldWithCTX(ctx)...).Debug(msg, fields...)
}
