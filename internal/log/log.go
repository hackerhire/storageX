package log

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
)

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
	logger.Infof(msg, v...)
}

func Error(msg string, v ...interface{}) {
	logger.Errorf(msg, v...)
}

func Fatal(msg string, v ...interface{}) {
	logger.Fatalf(msg, v...)
}
