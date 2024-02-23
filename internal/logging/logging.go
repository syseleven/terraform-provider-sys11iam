package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
)

var (
	isDebug bool
)

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	DEBUG LogLevel = "DEBUG"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// required while traversing the call stack to find the position of the caller; e.g. we'll skip Debug and Debugf
const InternalFuncNameSuffix = "logging.Debug"

// required to trim the file path to contain only the file name
const InternalPackageName = "glue-worker"

type LogFields map[string]interface{}

func EnableDebug() {
	isDebug = true
}

func Debug(args ...interface{}) {
	if isDebug {
		callerInfo := getLineAndNumberFromCallerFunc()
		logMessage := fmt.Sprint(args...)

		WriteOutput(DEBUG, fmt.Sprintf("%s: %s", callerInfo, logMessage), map[string]interface{}{})
	}
}

func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}

func Info(message string) {
	InfoWithFields(message, map[string]interface{}{})
}

func InfoWithFields(message string, fields LogFields) {
	WriteOutput(INFO, message, fields)
}

func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

func Error(message string) {
	ErrorWithFields(message, map[string]interface{}{})
}

func ErrorWithFields(message string, fields LogFields) {
	WriteOutput(ERROR, message, fields)
}

func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

func Fatal(err error) {
	WriteOutput(FATAL, err.Error(), map[string]interface{}{})
	os.Exit(1)
}

func WriteOutput(logLevel LogLevel, message string, fields LogFields) {
	var outputStr string
	fields["message"] = message
	fields["log_level"] = logLevel
	if isDebug {
		outputStr = MapToPlainText(fields)
	} else {
		outputStr = MapToJSON(fields)
	}

	if logLevel == ERROR || logLevel == FATAL {
		_, _ = fmt.Fprintln(os.Stderr, outputStr)
	} else {
		_, _ = fmt.Fprintln(os.Stdout, outputStr)
	}
}

func MapToPlainText(fields map[string]interface{}) string {
	stringArray := make([]string, 0)
	for key, value := range fields {
		stringArray = append(stringArray, fmt.Sprintf("%s: `%v`", key, value))
	}
	return strings.Join(stringArray, " ")
}

func MapToJSON(fields map[string]interface{}) string {
	payload, err := json.Marshal(fields)
	if err != nil {
		return fmt.Sprintf("{\"message\": \"%v\"}", MapToPlainText(fields))
	}
	return string(payload)
}

// derived from: https://golang.org/pkg/runtime/#example_Frames
func getLineAndNumberFromCallerFunc() string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(0, pc)
	if n == 0 {
		// No pcs available. Stop now.
		// This can happen if the first argument to runtime.Callers is large.
		return ""
	}
	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)

	// Loop to get frames.
	// A fixed number of pcs can expand to an indefinite number of Frames.
	var (
		foundLoggingFrames bool
	)
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.Function, InternalFuncNameSuffix) {
			foundLoggingFrames = true
			continue
		}
		if foundLoggingFrames {
			fname := frame.File
			fnameSplit := strings.Split(fname, InternalPackageName)
			if len(fnameSplit) == 2 {
				fname = fnameSplit[1]
			}
			return fmt.Sprintf("%s:%d", fname, frame.Line)
		}
		if !more {
			break
		}
	}
	return ""
}
