package zlog

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	ConsoleMode = "console"
	FileMode    = "file"
)

// Config describes the logging options used by the services in this project.
type Config struct {
	Name          string `json:",optional"`
	Level         string `json:",default=info"`
	Stacktrace    bool   `json:",default=true"`
	AddCaller     bool   `json:",default=true"`
	CallerShip    int    `json:",default=0"`
	Mode          string `json:",default=console,options=console|file"`
	FileName      string `json:",optional"`
	ErrorFileName string `json:",optional"`
	MaxSize       int    `json:",optional"`
	MaxAge        int    `json:",optional"`
	MaxBackup     int    `json:",optional"`
	Async         bool   `json:",optional"`
	JSON          bool   `json:"json,optional"`
	Compress      bool   `json:",optional"`
	Console       bool   `json:",optional"`
	Color         bool   `json:",optional"`
	Port          int32  `json:",optional"`
}

var (
	loggerMu      sync.RWMutex
	defaultLogger = zap.NewNop()
	serverOnce    sync.Once
)

// Build creates a zap logger from Config.
func (c Config) Build() *zap.Logger {
	level := parseLevel(c.Level)
	encoder := newEncoder(c.JSON, c.Color)

	mode := strings.ToLower(strings.TrimSpace(c.Mode))
	if mode == "" {
		mode = ConsoleMode
	}

	var cores []zapcore.Core
	switch mode {
	case ConsoleMode:
		cores = append(cores, zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level))
	case FileMode:
		if strings.TrimSpace(c.FileName) == "" {
			panic("zlog: fileName is required when mode is file")
		}
		cores = append(cores, zapcore.NewCore(encoder, c.fileWriter(c.FileName), level))
		if c.Console {
			cores = append(cores, zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level))
		}
	default:
		panic("zlog: mode must be console or file")
	}

	if strings.TrimSpace(c.ErrorFileName) != "" {
		cores = append(cores, zapcore.NewCore(
			encoder,
			c.fileWriter(c.ErrorFileName),
			zap.LevelEnablerFunc(func(l zapcore.Level) bool { return l >= zapcore.ErrorLevel }),
		))
	}

	options := make([]zap.Option, 0, 3)
	if c.AddCaller {
		options = append(options, zap.AddCaller())
		if c.CallerShip > 0 {
			options = append(options, zap.AddCallerSkip(c.CallerShip))
		}
	}
	if c.Stacktrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	result := zap.New(zapcore.NewTee(cores...), options...)
	if c.Name != "" {
		result = result.With(zap.String("service", c.Name))
	}
	if c.Port > 0 {
		startLevelServer(c.Port, level, result)
	}
	return result
}

func (c Config) fileWriter(filename string) zapcore.WriteSyncer {
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    c.MaxSize,
		MaxAge:     c.MaxAge,
		MaxBackups: c.MaxBackup,
		LocalTime:  true,
		Compress:   c.Compress,
	})
	if !c.Async {
		return zapcore.Lock(writer)
	}
	return &zapcore.BufferedWriteSyncer{
		WS:            zapcore.Lock(writer),
		Size:          256 * 1024,
		FlushInterval: 30 * time.Second,
	}
}

func parseLevel(value string) zap.AtomicLevel {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		value = zapcore.InfoLevel.String()
	}
	level, err := zap.ParseAtomicLevel(value)
	if err != nil {
		panic("zlog: invalid level " + value)
	}
	return level
}

func newEncoder(asJSON, color bool) zapcore.Encoder {
	config := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if color && !asJSON {
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	if asJSON {
		return zapcore.NewJSONEncoder(config)
	}
	return zapcore.NewConsoleEncoder(config)
}

func startLevelServer(port int32, level zap.AtomicLevel, logger *zap.Logger) {
	serverOnce.Do(func() {
		server := &http.Server{
			Addr:              "127.0.0.1:" + strconv.FormatInt(int64(port), 10),
			Handler:           level,
			ReadHeaderTimeout: 5 * time.Second,
		}
		go func() {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("log level server stopped", zap.Error(err))
			}
		}()
	})
}

// InitDefaultLogger replaces the process-wide logger used by service adapters.
func InitDefaultLogger(config *Config) {
	if config == nil {
		config = &Config{}
	}
	logger := config.Build()
	loggerMu.Lock()
	previous := defaultLogger
	defaultLogger = logger
	loggerMu.Unlock()
	_ = previous.Sync()
}

func GetZapLogger() *zap.Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return defaultLogger
}
