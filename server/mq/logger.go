package mq

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"runtime"
	"strings"
	"time"
)

// CallerEncoder will add caller to log. format is "filename:lineNum:funcName", e.g:"zaplog/zaplog_test.go:15:zaplog.TestNewLogger"
func callerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(strings.Join([]string{caller.TrimmedPath(), runtime.FuncForPC(caller.PC).Name()}, ":"))
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func NewLogger(logpath string, maxSize int) *logger {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeCaller = callerEncoder
	encodeConfig.TimeKey = "time"
	encodeConfig.EncodeTime = timeEncoder

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logpath,
		MaxSize:    maxSize, // megabytes
		MaxBackups: 30,
		MaxAge:     10, // days
	})

	level := zap.NewAtomicLevelAt(zap.DebugLevel)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encodeConfig),
		w,
		level,
	)

	return &logger{zap.New(core)}
}

type logger struct {
	_log *zap.Logger
}

// Print logs a message at level Info on the compatibleLogger.
func (l *logger) Print(args ...interface{}) {
	l._log.Info(fmt.Sprint(args...))
}

// Println logs a message at level Info on the compatibleLogger.
func (l *logger) Println(args ...interface{}) {
	l._log.Info(fmt.Sprint(args...))
}

// Printf logs a message at level Info on the compatibleLogger.
func (l *logger) Printf(format string, args ...interface{}) {
	l._log.Info(fmt.Sprintf(format, args...))
}

func (l *logger) With(key string, value interface{}) *logger {
	return &logger{l._log.With(zap.Any(key, value))}
}

func (l *logger) WithFields(fields map[string]interface{}) *logger {
	i := 0
	var clog *logger
	for k, v := range fields {
		if i == 0 {
			clog = l.With(k, v)
		} else {
			clog = clog.With(k, v)
		}
		i++
	}
	return clog
}

// Debug logs a message at level Debug on the compatibleLogger.
func (l *logger) Debug(args ...interface{}) {
	l._log.Debug(fmt.Sprint(args...))
}

// Debugln logs a message at level Debug on the compatibleLogger.
func (l *logger) Debugln(args ...interface{}) {
	l._log.Debug(fmt.Sprint(args...))
}

// Debugf logs a message at level Debug on the compatibleLogger.
func (l *logger) Debugf(format string, args ...interface{}) {
	l._log.Debug(fmt.Sprintf(format, args...))
}

// Info logs a message at level Info on the compatibleLogger.
func (l *logger) Info(args ...interface{}) {
	l._log.Info(fmt.Sprint(args...))
}

// Infoln logs a message at level Info on the compatibleLogger.
func (l *logger) Infoln(args ...interface{}) {
	l._log.Info(fmt.Sprint(args...))
}

// Infof logs a message at level Info on the compatibleLogger.
func (l *logger) Infof(format string, args ...interface{}) {
	l._log.Info(fmt.Sprintf(format, args...))
}

// Warn logs a message at level Warn on the compatibleLogger.
func (l *logger) Warn(args ...interface{}) {
	l._log.Warn(fmt.Sprint(args...))
}

// Warnln logs a message at level Warn on the compatibleLogger.
func (l *logger) Warnln(args ...interface{}) {
	l._log.Warn(fmt.Sprint(args...))
}

// Warnf logs a message at level Warn on the compatibleLogger.
func (l *logger) Warnf(format string, args ...interface{}) {
	l._log.Warn(fmt.Sprintf(format, args...))
}

// Error logs a message at level Error on the compatibleLogger.
func (l *logger) Error(args ...interface{}) {
	l._log.Error(fmt.Sprint(args...))
}

// Errorln logs a message at level Error on the compatibleLogger.
func (l *logger) Errorln(args ...interface{}) {
	l._log.Error(fmt.Sprint(args...))
}

// Errorf logs a message at level Error on the compatibleLogger.
func (l *logger) Errorf(format string, args ...interface{}) {
	l._log.Error(fmt.Sprintf(format, args...))
}

// Fatal logs a message at level Fatal on the compatibleLogger.
func (l *logger) Fatal(args ...interface{}) {
	l._log.Fatal(fmt.Sprint(args...))
}

// Fatalln logs a message at level Fatal on the compatibleLogger.
func (l *logger) Fatalln(args ...interface{}) {
	l._log.Fatal(fmt.Sprint(args...))
}

// Fatalf logs a message at level Fatal on the compatibleLogger.
func (l *logger) Fatalf(format string, args ...interface{}) {
	l._log.Fatal(fmt.Sprintf(format, args...))
}

// Panic logs a message at level Painc on the compatibleLogger.
func (l *logger) Panic(args ...interface{}) {
	l._log.Panic(fmt.Sprint(args...))
}

// Panicln logs a message at level Painc on the compatibleLogger.
func (l *logger) Panicln(args ...interface{}) {
	l._log.Panic(fmt.Sprint(args...))
}

// Panicf logs a message at level Painc on the compatibleLogger.
func (l *logger) Panicf(format string, args ...interface{}) {
	l._log.Panic(fmt.Sprintf(format, args...))
}
