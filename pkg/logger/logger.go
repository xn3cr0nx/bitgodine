package logger

import (
	"runtime"

	"github.com/fatih/color"
	"github.com/xn3cr0nx/bitgodine/internal/errorx"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Fields logrus Fields type
type Fields logrus.Fields

// Params mapping of string to interface to print params in logs
type Params map[string]interface{}

// Log custom logger initialized with Setup function
var Log *logrus.Logger

// Setup creates the new logger with custom configuration
func Setup() {
	printTitle()

	// This is mainly done to export the logger in test
	Log = logrus.New()
	Log.Formatter = &logrus.TextFormatter{
		ForceColors:            true,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
	}

	if viper.GetBool("debug") {
		Log.SetLevel(logrus.DebugLevel)
	}
}

func printTitle() {
	c := color.New(color.FgHiCyan)
	ascii := `
 _           _                      _                  
| |     _  _| |_                   | | _               
| |__  |_||_   _| _____  _____   __| ||_| _____  _____ 
|  _ \ | |  | |  |  _  ||  _  | / _  || ||  _  ||  __ |
| |_) || |  | |  | |_| || |_| || (_| || || | | ||  ___/
|____/ |_|  |_|   \___ ||_____| \____||_||_| |_||_____|
                   __| |                               
                  |____|                               `
	c.Println(ascii + "\n")
}

func withFields(fields Fields) *logrus.Entry {
	return Log.WithFields(logrus.Fields(fields))
}

// Info level log message
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

// Warn level log message
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

// Debug level log message
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

// Error level log message
func Error(action string, err error, params Params) {
	_, filename, line, _ := runtime.Caller(1)
	if err == nil { // something's wrong. fix needed
		err = errorx.ErrUnknown
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

// Panic level log message
func Panic(action string, err error, params Params) {
	_, filename, line, _ := runtime.Caller(1)
	if err == nil { // something's wrong. fix needed
		err = errorx.ErrUnknown
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
