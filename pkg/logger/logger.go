package logger

import (
	"errors"
	"runtime"

	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields
type Params map[string]string

var Log *logrus.Logger

func init() {
	// This is mainly done to export the logger in test
	Log = logrus.New()
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	})

}

func withFields(fields Fields) *logrus.Entry {
	return Log.WithFields(logrus.Fields(fields))
}

func Info(action, message string, params Params) {
	_, filename, line, _ := runtime.Caller(1)
	fields := make(map[string]interface{})
	fields["service"] = "bitgodine"
	fields["action"] = action
	fields["filename"] = filename
	fields["line"] = line
	if params != nil {
		for i, v := range params {
			fields[i] = v
		}
	}

	withFields(fields).Info(message)
}

func Warn(action, message string, params Params) {
	_, filename, line, _ := runtime.Caller(1)
	fields := make(map[string]interface{})
	fields["service"] = "bitgodine"
	fields["action"] = action
	fields["filename"] = filename
	fields["line"] = line
	if params != nil {
		for i, v := range params {
			fields[i] = v
		}
	}

	withFields(fields).Warn(message)
}

func Debug(action, message string, params Params) {
	_, filename, line, _ := runtime.Caller(1)
	fields := make(map[string]interface{})
	fields["service"] = "bitgodine"
	fields["action"] = action
	fields["filename"] = filename
	fields["line"] = line
	if params != nil {
		for i, v := range params {
			fields[i] = v
		}
	}

	withFields(fields).Debug(message)
}

func Error(action string, err error, params Params) {
	_, filename, line, _ := runtime.Caller(1)
	if err == nil { // something's wrong. fix needed
		err = errors.New("ERROR NOT PROVIDED")
	}
	fields := make(map[string]interface{})
	fields["service"] = "bitgodine"
	fields["action"] = action
	fields["filename"] = filename
	fields["line"] = line
	if params != nil {
		for i, v := range params {
			fields[i] = v
		}
	}

	withFields(fields).Error(err.Error())
}

func Panic(action string, err error, params Params) {
	_, filename, line, _ := runtime.Caller(1)
	if err == nil { // something's wrong. fix needed
		err = errors.New("ERROR NOT PROVIDED")
	}
	fields := make(map[string]interface{})
	fields["service"] = "bitgodine"
	fields["action"] = action
	fields["filename"] = filename
	fields["line"] = line
	if params != nil {
		for i, v := range params {
			fields[i] = v
		}
	}

	withFields(fields).Panic(err.Error())
}
