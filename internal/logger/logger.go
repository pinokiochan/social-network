package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Global logger instance
var Log *logrus.Logger

// Custom fields type for easier field handling
type Fields map[string]interface{}

// Initialize the logger with default configuration
func init() {
	Log = logrus.New()

	// Set custom formatter
	Log.SetFormatter(&PrettyFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LevelDesc: []string{
			logrus.PanicLevel: "PANIC",
			logrus.FatalLevel: "FATAL",
			logrus.ErrorLevel: "ERROR",
			logrus.WarnLevel:  "WARN",
			logrus.InfoLevel:  "INFO",
			logrus.DebugLevel: "DEBUG",
			logrus.TraceLevel: "TRACE",
		},
	})

	// Write to stdout by default
	Log.SetOutput(os.Stdout)

	// Set default level to Info, can be overridden by environment variable
	Log.SetLevel(getLogLevel())
}

// PrettyFormatter implements logrus.Formatter interface
type PrettyFormatter struct {
	TimestampFormat string
	LevelDesc       []string
}

// Format renders a single log entry
func (f *PrettyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := fmt.Sprintf(entry.Time.Format(f.TimestampFormat))
	level := f.LevelDesc[entry.Level]
	message := entry.Message

	// Get caller information
	var fileInfo string
	if entry.HasCaller() {
		fileInfo = fmt.Sprintf("%s:%d", filepath.Base(entry.Caller.File), entry.Caller.Line)
	}

	// Format log fields
	var fields string
	if len(entry.Data) > 0 {
		var fieldStrings []string
		for k, v := range entry.Data {
			fieldStrings = append(fieldStrings, fmt.Sprintf("%s=%v", k, v))
		}
		fields = " " + strings.Join(fieldStrings, " ")
	}

	// Construct the log line
	logLine := fmt.Sprintf("%s [%s] %-44s %s%s\n", timestamp, level, message, fileInfo, fields)

	return []byte(logLine), nil
}

// getLogLevel returns the log level based on environment variable or defaults to Info
func getLogLevel() logrus.Level {
	lvl, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		return logrus.InfoLevel
	}

	level, err := logrus.ParseLevel(lvl)
	if err != nil {
		return logrus.InfoLevel
	}

	return level
}

// RequestLogger logs HTTP request details
func RequestLogger(r *http.Request, fields Fields) {
	Log.WithFields(logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"ip":         r.RemoteAddr,
		"user_agent": r.UserAgent(),
		"request_id": r.Context().Value("request_id"),
	}).WithFields(logrus.Fields(fields)).Info("Incoming request")
}

// ErrorLogger logs error with context
func ErrorLogger(err error, fields Fields) {
	Log.WithFields(logrus.Fields(fields)).Error(err)
}

// InfoLogger logs information with context
func InfoLogger(message string, fields Fields) {
	Log.WithFields(logrus.Fields(fields)).Info(message)
}

// DebugLogger logs debug information
func DebugLogger(message string, fields Fields) {
	Log.WithFields(logrus.Fields(fields)).Debug(message)
}

// WarnLogger logs warnings
func WarnLogger(message string, fields Fields) {
	Log.WithFields(logrus.Fields(fields)).Warn(message)
}
