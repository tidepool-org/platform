package logger

import (
	"os"

	logrus "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

//Logger interface
type Logger interface {
	AddTrace(id string)
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

var log = logrus.New()

//Logging variable
var Logging Logger

const traceID = "traceID"

//PlatformLogger type
type PlatformLogger struct {
	field string
}

func init() {
	log.Out = os.Stdout
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.InfoLevel
	Logging = &PlatformLogger{}
}

//AddTrace will log add a trace id to the logs
func (platformLogger *PlatformLogger) AddTrace(id string) { platformLogger.field = id }

//Debug will log at the Debug Level
func (platformLogger *PlatformLogger) Debug(args ...interface{}) {
	if platformLogger.field != "" {
		log.WithField(traceID, platformLogger.field).Debug(args)
		return
	}
	log.Debug(args)
}

//Info will log at the Info Level
func (platformLogger *PlatformLogger) Info(args ...interface{}) {
	if platformLogger.field != "" {
		log.WithField(traceID, platformLogger.field).Info(args)
		return
	}
	log.Info(args)
}

//Warn will log at the Warn Level
func (platformLogger *PlatformLogger) Warn(args ...interface{}) {
	if platformLogger.field != "" {
		log.WithField(traceID, platformLogger.field).Warn(args)
		return
	}
	log.Warn(args)
}

//Error will log at the Error Level
func (platformLogger *PlatformLogger) Error(args ...interface{}) {
	if platformLogger.field != "" {
		log.WithField(traceID, platformLogger.field).Error(args)
		return
	}
	log.Error(args)
}

//Fatal will log at the Fatal Level
func (platformLogger *PlatformLogger) Fatal(args ...interface{}) {
	if platformLogger.field != "" {
		log.WithField(traceID, platformLogger.field).Fatal(args)
		return
	}
	log.Fatal(args)
}
