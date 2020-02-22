package zaplog

import (
	"github.com/pkg/errors"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger        *zap.Logger
	sugaredLogger *zap.SugaredLogger
	once          sync.Once
	atomicLevel   zap.AtomicLevel
)

const (
	timeFormatLayout = "2006-01-02T15:04:05.000Z07:00"
)

// CustomizedTimeEncoder new a zap customized time encoder
func CustomizedTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(timeFormatLayout))
}

// NewEncoderConfig new a zap encoder config
func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     CustomizedTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// NewLogger new a zap logger and return its atomic level
// If the logPath is empty (use WithLogPath) or the
// development (use Development) is set, the logger will
// write log to os.stderr. In production, you have better
// specify the serviceName(use WithServiceName).
func NewLogger(opts ...Option) (*zap.Logger, zap.AtomicLevel) {
	logOpts := defaultLogOptions
	for i := 0; i < len(opts); i++ {
		opts[i].apply(&logOpts)
	}
	atomicLevel := zap.NewAtomicLevelAt(logOpts.logLevel)
	encoderConfig := NewEncoderConfig()
	var encoder zapcore.Encoder
	var logCores []zapcore.Core
	if logOpts.logPath == "" || logOpts.development == true {
		encoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
		logCores = []zapcore.Core{zapcore.NewCore(encoder, os.Stderr, atomicLevel)}
	} else {
		// k8s pod name
		podName := os.Getenv("KUBE_PODNAME")
		encoder = zapcore.NewJSONEncoder(encoderConfig)
		levelEnablerFuncMap := levelAndAbove(atomicLevel.Level()).EnableLevels()
		logCores = make([]zapcore.Core, 0, len(levelEnablerFuncMap))
		var err error
		for level, levelEnablerFunc := range levelEnablerFuncMap {
			err = os.MkdirAll(filepath.Join(logOpts.logPath, podName, logOpts.serviceName), 0755)
			if err != nil {
				panic(errors.Wrap(err, "error make directory"))
			}
			var pattern string
			pattern = filepath.Join(logOpts.logPath, podName, logOpts.serviceName, level.String()+".%Y-%m-%d.log")
			fileWriter, err := rotatelogs.New(pattern,
				rotatelogs.WithMaxAge(logOpts.fileRotateMaxAge),
				rotatelogs.WithRotationTime(logOpts.fileRotationTime))
			if err != nil {
				panic(errors.Wrap(err, "error create file rotate writer"))
			}
			writeSyncer := zapcore.Lock(zapcore.AddSync(fileWriter))
			logCores = append(logCores, zapcore.NewCore(encoder, writeSyncer, levelEnablerFunc))
		}
	}

	var newLogger *zap.Logger
	if logOpts.development {
		newLogger = zap.New(zapcore.NewTee(logCores...), zap.Development())
	} else {
		newLogger = zap.New(zapcore.NewTee(logCores...))
	}

	if logOpts.addCaller {
		newLogger = newLogger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(logOpts.callerSkip))
	}
	if logOpts.addStacktrace != nil {
		newLogger = newLogger.WithOptions(zap.AddStacktrace(logOpts.addStacktrace))
	}
	if logOpts.wrapCoreFunc != nil {
		newLogger = newLogger.WithOptions(zap.WrapCore(logOpts.wrapCoreFunc))
	}
	if len(logOpts.fields) != 0 {
		newLogger = newLogger.WithOptions(zap.Fields(logOpts.fields...))
	}
	if len(logOpts.hooks) != 0 {
		newLogger = newLogger.WithOptions(zap.Hooks(logOpts.hooks...))
	}
	if logOpts.errorOutput != nil {
		newLogger = newLogger.WithOptions(zap.ErrorOutput(logOpts.errorOutput))
	}
	return newLogger, atomicLevel
}

// NewSugaredLogger new a sugar logger for using method such as
// log.Infof(template string, args ...interface{}) or
// log.Infow(msg string, keysAndValues ...interface{})
func NewSugaredLogger(opts ...Option) (*zap.SugaredLogger, zap.AtomicLevel) {
	newLogger, atomicLevel := NewLogger(opts...)
	return newLogger.Sugar(), atomicLevel
}

// InitLogger initialize a global logger and sugar logger to use
func InitLogger(opts ...Option) {
	once.Do(func() {
		logger, atomicLevel = NewLogger(opts...)
		sugaredLogger = logger.Sugar()
	})
}

// ChangeLoggerLevel will atomically change the log level
func AtomicSetLoggerLevel(lvl zapcore.Level) {
	atomicLevel.SetLevel(lvl)
}

func AtomicLevelHandler(w http.ResponseWriter, r *http.Request) {
	atomicLevel.ServeHTTP(w, r)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(msg string, fields ...zap.Field) {
	logger.DPanic(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

// Debugf uses fmt.Sprintf to log a templated message.
// When the template is empty string, the Debug method will be used
func Debugf(template string, args ...interface{}) {
	sugaredLogger.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
// When the template is empty string, the Info method will be used
func Infof(template string, args ...interface{}) {
	sugaredLogger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
// When the template is empty string, the Warn method will be used
func Warnf(template string, args ...interface{}) {
	sugaredLogger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
// When the template is empty string, the Error method will be used
func Errorf(template string, args ...interface{}) {
	sugaredLogger.Errorf(template, args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the
// logger then panics. (See DPanicLevel for details.)
// When the template is empty string, the DPanic method will be used
func DPanicf(template string, args ...interface{}) {
	sugaredLogger.DPanicf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
// When the template is empty string, the Panic method will be used
func Panicf(template string, args ...interface{}) {
	sugaredLogger.Panicf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
// When the template is empty string, the Fatal method will be used
func Fatalf(template string, args ...interface{}) {
	sugaredLogger.Fatalf(template, args...)
}

// Debugw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
//
// When debug-level logging is disabled, this is much faster than
//  s.With(keysAndValues).Debug(msg)
func Debugw(msg string, keysAndValues ...interface{}) {
	sugaredLogger.Debugw(msg, keysAndValues...)
}

// Infow logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Infow(msg string, keysAndValues ...interface{}) {
	sugaredLogger.Infow(msg, keysAndValues...)
}

// Warnw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Warnw(msg string, keysAndValues ...interface{}) {
	sugaredLogger.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func Errorw(msg string, keysAndValues ...interface{}) {
	sugaredLogger.Errorw(msg, keysAndValues...)
}

// DPanicw logs a message with some additional context. In development, the
// logger then panics. (See DPanicLevel for details.) The variadic key-value
// pairs are treated as they are in With.
func DPanicw(msg string, keysAndValues ...interface{}) {
	sugaredLogger.DPanicw(msg, keysAndValues...)
}

// Panicw logs a message with some additional context, then panics. The
// variadic key-value pairs are treated as they are in With.
func Panicw(msg string, keysAndValues ...interface{}) {
	sugaredLogger.Panicw(msg, keysAndValues...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The
// variadic key-value pairs are treated as they are in With.
func Fatalw(msg string, keysAndValues ...interface{}) {
	sugaredLogger.Fatalw(msg, keysAndValues...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func Sync() error {
	err := logger.Sync()
	if err != nil {
		sugaredLogger.Sync()
		return err
	}
	err = sugaredLogger.Sync()
	return err
}
