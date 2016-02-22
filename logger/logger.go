package logger

import (
	"os"

	logrus "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

type Logger interface {
	AddTrace(id string)
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

var log = logrus.New()
var Logging Logger

const trace_id = "traceid"

type PlatformLogger struct {
	field string
}

func init() {
	log.Out = os.Stdout
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.InfoLevel
	Logging = &PlatformLogger{}
}

func (this *PlatformLogger) AddTrace(id string) { this.field = id }

func (this *PlatformLogger) Debug(args ...interface{}) {
	if this.field != "" {
		log.WithField(trace_id, this.field).Debug(args)
		return
	}
	log.Debug(args)
}
func (this *PlatformLogger) Info(args ...interface{}) {
	if this.field != "" {
		log.WithField(trace_id, this.field).Info(args)
		return
	}
	log.Info(args)
}
func (this *PlatformLogger) Warn(args ...interface{}) {
	if this.field != "" {
		log.WithField(trace_id, this.field).Warn(args)
		return
	}
	log.Warn(args)
}
func (this *PlatformLogger) Error(args ...interface{}) {
	if this.field != "" {
		log.WithField(trace_id, this.field).Error(args)
		return
	}
	log.Error(args)
}
func (this *PlatformLogger) Fatal(args ...interface{}) {
	if this.field != "" {
		log.WithField(trace_id, this.field).Fatal(args)
		return
	}
	log.Fatal(args)
}
