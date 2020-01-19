package zaplog

import (
	"go.uber.org/zap/zapcore"
	"time"
)

// if you want to write log to file,
// you must set the serviceName and the logPath
type logOptions struct {
	serviceName      string
	logPath          string
	fileRotateMaxAge time.Duration
	fileRotationTime time.Duration
	logLevel         zapcore.Level
	development      bool
}

var defaultLogOptions = logOptions{
	serviceName: "",
	logPath:     "",
	logLevel:    zapcore.DebugLevel,
	// 0 means the fileRotateMaxAge will default set to 7 * 24 * time.Hour
	fileRotateMaxAge: 0,
	// 0 means the fileRotationTime will default set to 24 * time.Hour
	fileRotationTime: 0,
}

type Option interface {
	apply(*logOptions)
}

type logOptionFunc func(*logOptions)

func (f logOptionFunc) apply(opt *logOptions) {
	f(opt)
}

// WithServiceName when in the production,
//the service name specify the the name of the log file dir name
func WithServiceName(serviceName string) Option {
	return logOptionFunc(func(o *logOptions) {
		o.serviceName = serviceName
	})
}

// WithLogPath will write log the the path
func WithLogPath(logPath string) Option {
	return logOptionFunc(func(o *logOptions) {
		o.logPath = logPath
	})
}

// WithLogLevel will set the log level to the given
// level and above
func WithLogLevel(logLevel zapcore.Level) Option {
	return logOptionFunc(func(o *logOptions) {
		o.logLevel = logLevel
	})
}

// WithFileRotateMaxAge creates a new Option that sets the
// max age of a log file before it gets purged from
// the file system.
func WithFileRotateMaxAge(maxAge time.Duration) Option {
	return logOptionFunc(func(o *logOptions) {
		o.fileRotateMaxAge = maxAge
	})
}

// WithRotationTime creates a new Option that sets the
// time between rotation.
func WithFileRotationTime(rotationTime time.Duration) Option {
	return logOptionFunc(func(o *logOptions) {
		o.fileRotationTime = rotationTime
	})
}

// Development set the logger to development
func Development() Option {
	return logOptionFunc(func(o *logOptions) {
		o.development = true
	})
}
