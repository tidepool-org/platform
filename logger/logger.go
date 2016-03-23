package logger

import (
	"os"

	logrus "github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/tidepool-org/platform/Godeps/_workspace/src/github.com/satori/go.uuid"
)

//Logger interface
type Logger interface {
	GetNamed(name string) Logger
	WithField(key string, value interface{}) Logger
	AddTrace(id string)
	AddTraceUUID()
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

const traceID = "traceID"

//PlatformLog type
type PlatformLog struct {
	fields   map[string]interface{}
	internal *logrus.Logger
}

//Log is an initialised PlatformLog instance
var Log *PlatformLog

//init will create a std logger that implements the `Logger` interface with all methods exported below
//your other option is to roll your own logger again making sure that it implements the `Logger` interface
func init() {
	Log = setup()
}

func setup() *PlatformLog {

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Formatter = new(logrus.JSONFormatter)
	logger.Level = logrus.ErrorLevel

	return &PlatformLog{
		internal: logger,
		fields:   make(map[string]interface{}),
	}

}

//GetNamed return a named instance of the logger
func (log *PlatformLog) GetNamed(name string) Logger {
	named := setup()
	named.WithField("name", name)
	return named
}

//AddTrace will add a trace id to the logs
func (log *PlatformLog) AddTrace(id string) {
	Log.internal.WithField(traceID, id)
}

//AddTraceUUID will add a trace id to the logs
func (log *PlatformLog) AddTraceUUID() {
	id := uuid.NewV4().String()
	Log.internal.WithField(traceID, id)
}

//WithField will log the message with and extra attached details
func (log *PlatformLog) WithField(key string, value interface{}) Logger {
	Log.fields[key] = value
	return log
}

//Debug Level
func (log *PlatformLog) Debug(args ...interface{}) {
	Log.internal.WithFields(Log.fields).Debug(args)
}

//Info Level
func (log *PlatformLog) Info(args ...interface{}) {
	Log.internal.WithFields(Log.fields).Info(args)
}

//Warn Level
func (log *PlatformLog) Warn(args ...interface{}) {
	Log.internal.WithFields(Log.fields).Warn(args)
}

//Error Level
func (log *PlatformLog) Error(args ...interface{}) {
	Log.internal.WithFields(Log.fields).Error(args)
}

//Fatal Level
func (log *PlatformLog) Fatal(args ...interface{}) {
	Log.internal.WithFields(Log.fields).Fatal(args)
}
