package test

import "github.com/tidepool-org/platform/log"

func NewLogger() log.Logger {
	return &logger{}
}

type logger struct{}

func (l *logger) Debug(message string)                               {}
func (l *logger) Info(message string)                                {}
func (l *logger) Warn(message string)                                {}
func (l *logger) Error(message string)                               {}
func (l *logger) WithError(err error) log.Logger                     { return l }
func (l *logger) WithField(key string, value interface{}) log.Logger { return l }
func (l *logger) WithFields(fields log.Fields) log.Logger            { return l }
