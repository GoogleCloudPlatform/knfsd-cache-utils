package log

import (
	"io"
	"log"
)

var (
	debug = log.New(io.Discard, "DEBUG: ", 0)
	info  = log.New(log.Writer(), "INFO: ", 0)
	error = log.New(log.Writer(), "ERROR: ", 0)
	fatal = log.New(log.Writer(), "ERROR: ", 0)
)

func EnableDebug() {
	debug.SetOutput(log.Writer())
}

func Debug(v ...interface{}) {
	debug.Print(v...)
}

func Debugf(format string, v ...interface{}) {
	debug.Printf(format, v...)
}

func Info(v ...interface{}) {
	info.Print(v...)
}

func Infof(format string, v ...interface{}) {
	info.Printf(format, v...)
}

func Error(v ...interface{}) {
	error.Print(v...)
}

func Errorf(format string, v ...interface{}) {
	error.Printf(format, v...)
}

func Fatal(v ...interface{}) {
	fatal.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	fatal.Fatalf(format, v...)
}
