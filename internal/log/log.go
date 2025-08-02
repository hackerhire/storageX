package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sayuyere/storageX/internal/config"
)

var (
	logger   *zap.SugaredLogger
	initOnce sync.Once
)

// ensureLogger initializes the logger if it hasn't been initialized yet, using config.LogDebug
func ensureLogger() {
	initOnce.Do(func() {
		cfg := config.GetConfig()
		debug := false
		if cfg != nil {
			debug = cfg.Log.Debug // Use config value if available
		}
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "time"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(zapcore.Lock(os.Stdout)),
			ifLevel(debug),
		)
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	})
}

// InitLogger allows explicit initialization with a debug flag (overrides config)
func InitLogger(debug bool) {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		ifLevel(debug),
	)
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}

func ifLevel(debug bool) zapcore.Level {
	if debug {
		return zapcore.DebugLevel
	}
	return zapcore.InfoLevel
}

func Info(msg string, v ...interface{}) {
	ensureLogger()
	logger.Infof(msg, v...)
}

func Error(msg string, v ...interface{}) {
	ensureLogger()
	logger.Errorf(msg, v...)
}

func Fatal(msg string, v ...interface{}) {
	ensureLogger()
	logger.Fatalf(msg, v...)
}
