package logger

import (
	logrus "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

var log = logrus.New()
var Logging Logger

type PlatformLogger struct{}

func init() {
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.ErrorLevel
	Logging = &PlatformLogger{}
}

func (this *PlatformLogger) Debug(args ...interface{}) { log.Debug(args) }
func (this *PlatformLogger) Info(args ...interface{})  { log.Info(args) }
func (this *PlatformLogger) Warn(args ...interface{})  { log.Warn(args) }
func (this *PlatformLogger) Error(args ...interface{}) { log.Error(args) }
func (this *PlatformLogger) Fatal(args ...interface{}) { log.Fatal(args) }

func (this *PlatformLogger) Debugf(format string, args ...interface{}) { log.Debugf(format, args) }
func (this *PlatformLogger) Infof(format string, args ...interface{})  { log.Infof(format, args) }
func (this *PlatformLogger) Warnf(format string, args ...interface{})  { log.Warnf(format, args) }
func (this *PlatformLogger) Errorf(format string, args ...interface{}) { log.Errorf(format, args) }
func (this *PlatformLogger) Fatalf(format string, args ...interface{}) { log.Fatalf(format, args) }
