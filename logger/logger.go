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
	WithField(key string, value interface{})
}

const traceID = "traceID"

//platformLogger type
type platformLogger struct {
	fields   map[string]interface{}
	internal *logrus.Logger
}

var std *platformLogger

//init will create a std logger that implements the `Logger` interface with all methods exported below
//your other option is to roll your own logger again making sure that it implements the `Logger` interface
func init() {

	std = &platformLogger{
		internal: logrus.New(),
	}

	std.internal.Out = os.Stdout
	std.internal.Formatter = new(logrus.JSONFormatter)
	std.internal.Level = logrus.InfoLevel

}

//AddTrace will add a trace id to the logs
func AddTrace(id string) {
	std.internal.WithField(traceID, id)
}

//WithField will log the message with and extra attached details
func WithField(key string, value interface{}) {
	std.fields[key] = value
}

//Debug will log at the Debug Level
func Debug(args ...interface{}) {
	std.internal.WithFields(std.fields).Debug(args)
}

//Info will log at the Info Level
func Info(args ...interface{}) {
	std.internal.WithFields(std.fields).Info(args)
}

//Warn will log at the Warn Level
func Warn(args ...interface{}) {
	std.internal.WithFields(std.fields).Warn(args)
}

//Error will log at the Error Level
func Error(args ...interface{}) {
	std.internal.WithFields(std.fields).Error(args)
}

//Fatal will log at the Fatal Level
func Fatal(args ...interface{}) {
	std.internal.WithFields(std.fields).Fatal(args)
}
