package logger

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	// log rolling package
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger         *zap.Logger
	LoggerInitOnce sync.Once
)

type LoggerCfg struct {
	EnableFile bool
	LogLevel   string
	LogPath    string
}

var LogLevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"panic": zapcore.PanicLevel,
	"fatal": zapcore.FatalLevel,
}

func findLevel(level string) zapcore.Level {
	if l, ok := LogLevelMap[level]; ok {
		return l
	}

	Infof("Invalid log level %s, using default log level Debug")
	return zapcore.DebugLevel
}

var defaultLogger = NewZapLogger(false, zapcore.InfoLevel, "logs")

func NewZapLogger(enableFile bool, logLevel zapcore.Level, logPath string) *zap.Logger {
	// AddSync converts an io.Writer to a WriteSyncer.
	stdout := zapcore.AddSync(os.Stdout)

	// set log level
	level := zap.NewAtomicLevelAt(logLevel)

	// production encoder config
	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	productionCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	productionCfg.EncodeDuration = zapcore.SecondsDurationEncoder

	// development encoder config
	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	developmentCfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	developmentCfg.ConsoleSeparator = " "

	// encoder
	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	var zapcores []zapcore.Core

	if enableFile {
		filename := filepath.Join(logPath, "echo.log")

		// AddSync converts an io.Writer to a WriteSyncer.
		file := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    10, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
		})
		zapcores = append(zapcores, zapcore.NewCore(fileEncoder, file, level))
	}

	zapcores = append(zapcores, zapcore.NewCore(consoleEncoder, stdout, level))

	// combine cores
	core := zapcore.NewTee(zapcores...)

	logger := zap.New(core, zap.AddStacktrace(zap.PanicLevel), zap.AddCaller(), zap.AddCallerSkip(1))

	// call global logger
	// zap.L().Info("global logger test")
	zap.ReplaceGlobals(logger)

	return logger
}

// InitLogger initializes the logger, enableFile is optional, default is false
func InitLogger(cfg *LoggerCfg) {
	Debugf("Logger configuration: %+v", cfg)
	// only initialize once
	LoggerInitOnce.Do(func() {
		Logger = NewZapLogger(cfg.EnableFile, findLevel(cfg.LogLevel), cfg.LogPath)
		Logger.Debug("Logger initialized")
	})
}

// GetLogger returns the logger,
// if Logger is nil, return defaultLogger
func GetLogger() *zap.Logger {
	if Logger == nil {
		// fmt.Printf("Logger is nil, using default logger\n")
		return defaultLogger
	}

	return Logger
}

func SyncLogger() {
	Logger.Sync()
}

// Wrapper functions for logging

func Debugf(template string, args ...any) {
	GetLogger().Sugar().Debugf(template, args...) // Sugar() is used to log with field
}

func Infof(template string, args ...any) {
	GetLogger().Sugar().Infof(template, args...) // Sugar() is used to log with field
}

func Warnf(template string, args ...any) {

	GetLogger().Sugar().Warnf(template, args...) // Sugar() is used to log with field
}

func Errorf(template string, args ...any) {

	GetLogger().Sugar().Errorf(template, args...) // Sugar() is used to log with field
}

// call panic() after logging
func Panicf(template string, args ...any) {

	GetLogger().Sugar().Panicf(template, args...) // Sugar() is used to log with field
}

// call os.Exit(1) after logging
func Fatalf(template string, args ...any) {
	GetLogger().Sugar().Fatalf(template, args...) // Sugar() is used to log with field
}
