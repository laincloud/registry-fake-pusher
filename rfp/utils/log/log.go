package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

const (
	kCallDepth = 2
)

type Logger struct {
	wrapped     *log.Logger
	debug       bool
	exitOnFatal bool
}

var l *Logger

func init() {
	l = &Logger{
		wrapped: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func UnwrappedLogger() *log.Logger {
	return l.wrapped
}

func EnableDebug() {
	l.debug = true
}

func EnableExitOnFatal() {
	l.exitOnFatal = true
}

func DebugV(v interface{}) {
	Debugf("%+v", v)
}

func Debug(v ...interface{}) {
	if l.debug {
		l.wrapped.Output(kCallDepth, header("DEBUG", fmt.Sprint(v...)))
	}
}

func Debugf(format string, v ...interface{}) {
	if l.debug {
		l.wrapped.Output(kCallDepth, header("DEBUG", fmt.Sprintf(format, v...)))
	}
}

func Info(v ...interface{}) {
	l.wrapped.Output(kCallDepth, header("INFO", fmt.Sprint(v...)))
}

func Infof(format string, v ...interface{}) {
	l.wrapped.Output(kCallDepth, header("INFO", fmt.Sprintf(format, v...)))
}

func Warn(v ...interface{}) {
	l.wrapped.Output(kCallDepth, header("WARN", fmt.Sprint(v...)))
}

func Warnf(format string, v ...interface{}) {
	l.wrapped.Output(kCallDepth, header("WARN", fmt.Sprintf(format, v...)))
}

func Error(v ...interface{}) {
	l.wrapped.Output(kCallDepth, header("ERROR", fmt.Sprint(v...)))
}

func Errorf(format string, v ...interface{}) {
	l.wrapped.Output(kCallDepth, header("ERROR", fmt.Sprintf(format, v...)))
}

func Fatal(v ...interface{}) {
	msg := header("FATAL", fmt.Sprint(v...))
	l.wrapped.Output(kCallDepth, msg)
	if l.exitOnFatal {
		os.Exit(1)
	} else {
		panic(msg)
	}
}

func Fatalf(format string, v ...interface{}) {
	msg := header("FATAL", fmt.Sprintf(format, v...))
	l.wrapped.Output(kCallDepth, msg)
	if l.exitOnFatal {
		os.Exit(1)
	} else {
		panic(msg)
	}
}

func header(level, msg string) string {
	_, file, line, ok := runtime.Caller(kCallDepth)
	if ok {
		file = filepath.Base(file)
	}
	if len(file) == 0 {
		file = "???"
	}
	if line < 0 {
		line = 0
	}

	return fmt.Sprintf("%s %s:%d: %s", level, file, line, msg)
}
