package log

import (
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var defaultLogger = logrus.New()

// Format selects the output format used by the package logger.
type Format string

// Level is the minimum severity that the package logger writes.
type Level string

const (
	// FormatText writes human-readable text. It is the default format.
	FormatText Format = "text"
	// FormatJSON writes structured JSON.
	FormatJSON Format = "json"

	// PanicLevel writes panic-level messages only.
	PanicLevel Level = "panic"
	// FatalLevel writes fatal-level messages and panic-level messages.
	FatalLevel Level = "fatal"
	// ErrorLevel writes error-level messages and more severe messages.
	ErrorLevel Level = "error"
	// WarnLevel writes warning-level messages and more severe messages.
	WarnLevel Level = "warn"
	// InfoLevel writes info-level messages and more severe messages. It is the default level.
	InfoLevel Level = "info"
	// DebugLevel writes debug-level messages and more severe messages.
	DebugLevel Level = "debug"
	// TraceLevel writes all supported log levels, including trace.
	TraceLevel Level = "trace"
)

// LoggerOptions configures the package-wide logger.
// The zero value selects info-level text output to standard output.
type LoggerOptions struct {
	// Level is the minimum severity to write. Unknown values use InfoLevel.
	Level Level
	// Format selects FormatText or FormatJSON. Unknown values use FormatText.
	Format Format
	// TimestampFormat is a Go time layout. The default is time.RFC3339.
	TimestampFormat string
	// ForceColors enables ANSI colors even when output is not a terminal.
	ForceColors bool
	// EnableColors enables colors for text output when supported.
	EnableColors bool
	// ForceQuote quotes all values in text output.
	ForceQuote bool
	// DisableQuote disables quoting values in text output unless ForceQuote is set.
	DisableQuote bool
	// EnvironmentOverrideColors lets CLICOLOR and CLICOLOR_FORCE control text colors.
	EnvironmentOverrideColors bool
	// DisableTimestamp omits timestamps from log entries.
	DisableTimestamp bool
	// FullTimestamp prints a full timestamp in text output instead of elapsed time.
	FullTimestamp bool
	// DisableSorting disables sorting of fields in text output.
	DisableSorting bool
	// QuoteEmptyFields quotes empty values in text output.
	QuoteEmptyFields bool
	// PrettyPrint indents JSON output for readability.
	PrettyPrint bool
	// SortingFunc customizes field-key sorting in text output.
	SortingFunc func([]string)
}

func init() {
	Config(LoggerOptions{})
}

// Config replaces the package-wide logger configuration.
// Call it during application startup, before starting goroutines that log.
func Config(opts LoggerOptions) {
	defaultLogger.SetLevel(getLevel(opts.Level))
	defaultLogger.SetFormatter(getFormat(opts))
	defaultLogger.SetOutput(os.Stdout)
}

func getLevel(level Level) logrus.Level {
	switch strings.ToLower(string(level)) {
	case string(PanicLevel):
		return logrus.PanicLevel
	case string(FatalLevel):
		return logrus.FatalLevel
	case string(ErrorLevel):
		return logrus.ErrorLevel
	case string(WarnLevel):
		return logrus.WarnLevel
	case string(InfoLevel):
		return logrus.InfoLevel
	case string(DebugLevel):
		return logrus.DebugLevel
	case string(TraceLevel):
		return logrus.TraceLevel
	}
	return logrus.InfoLevel
}

func getFormat(opts LoggerOptions) logrus.Formatter {
	tf := opts.TimestampFormat
	if tf == "" {
		tf = time.RFC3339
	}

	switch strings.ToLower(string(opts.Format)) {
	case string(FormatJSON):
		return &logrus.JSONFormatter{
			TimestampFormat:  tf,
			DisableTimestamp: opts.DisableTimestamp,
			PrettyPrint:      opts.PrettyPrint,
		}
	default:
		return &logrus.TextFormatter{
			TimestampFormat:           tf,
			FullTimestamp:             opts.FullTimestamp,
			DisableColors:             !opts.EnableColors,
			ForceColors:               opts.ForceColors,
			DisableQuote:              opts.DisableQuote,
			ForceQuote:                opts.ForceQuote,
			DisableTimestamp:          opts.DisableTimestamp,
			EnvironmentOverrideColors: opts.EnvironmentOverrideColors,
			DisableSorting:            opts.DisableSorting,
			QuoteEmptyFields:          opts.QuoteEmptyFields,
			SortingFunc:               opts.SortingFunc,
		}
	}
}

// Attribute is a structured key-value pair attached to a log entry.
type Attribute struct {
	// Key is the field name included in the log entry.
	Key string
	// Value is the field value included in the log entry.
	Value any
}

// Err creates an "error" attribute from err.
func Err(err error) Attribute {
	return Attribute{Key: "error", Value: err}
}

// Field creates a structured attribute with the given key and value.
func Field(key string, value any) Attribute {
	return Attribute{Key: key, Value: value}
}

// Info logs msg and its structured fields at info level.
func Info(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Info(msg)
}

// Infof logs a formatted message at info level.
func Infof(format string, args ...any) {
	defaultLogger.Infof(format, args...)
}

// Infoln logs the arguments at info level, separated by spaces.
func Infoln(args ...any) {
	defaultLogger.Infoln(args...)
}

// Error logs msg and its structured fields at error level.
func Error(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Error(msg)
}

// Errorf logs a formatted message at error level.
func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

// Errorln logs the arguments at error level, separated by spaces.
func Errorln(args ...any) {
	defaultLogger.Errorln(args...)
}

// Trace logs msg and its structured fields at trace level.
func Trace(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Trace(msg)
}

// Tracef logs a formatted message at trace level.
func Tracef(format string, args ...any) {
	defaultLogger.Tracef(format, args...)
}

// Traceln logs the arguments at trace level, separated by spaces.
func Traceln(args ...any) {
	defaultLogger.Traceln(args...)
}

// Panic logs msg and its structured fields at panic level, then panics.
func Panic(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Panic(msg)
}

// Panicf logs a formatted message at panic level, then panics.
func Panicf(format string, args ...any) {
	defaultLogger.Panicf(format, args...)
}

// Panicln logs the arguments at panic level, separated by spaces, then panics.
func Panicln(args ...any) {
	defaultLogger.Panicln(args...)
}

// Debug logs msg and its structured fields at debug level.
func Debug(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Debug(msg)
}

// Debugf logs a formatted message at debug level.
func Debugf(format string, args ...any) {
	defaultLogger.Debugf(format, args...)
}

// Debugln logs the arguments at debug level, separated by spaces.
func Debugln(args ...any) {
	defaultLogger.Debugln(args...)
}

// Print logs msg and its structured fields at info level.
func Print(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Print(msg)
}

// Printf logs a formatted message at info level.
func Printf(format string, args ...any) {
	defaultLogger.Printf(format, args...)
}

// Println logs the arguments at info level, separated by spaces.
func Println(args ...any) {
	defaultLogger.Println(args...)
}

// Warn logs msg and its structured fields at warning level.
func Warn(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Warn(msg)
}

// Warnf logs a formatted message at warning level.
func Warnf(format string, args ...any) {
	defaultLogger.Warnf(format, args...)
}

// Warnln logs the arguments at warning level, separated by spaces.
func Warnln(args ...any) {
	defaultLogger.Warnln(args...)
}

// Fatal logs msg and its structured fields at fatal level, then exits the process.
func Fatal(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Fatal(msg)
}

// Fatalf logs a formatted message at fatal level, then exits the process.
func Fatalf(format string, args ...any) {
	defaultLogger.Fatalf(format, args...)
}

// Fatalln logs the arguments at fatal level, separated by spaces, then exits the process.
func Fatalln(args ...any) {
	defaultLogger.Fatalln(args...)
}

func toLogrusFields(fields []Attribute) logrus.Fields {
	f := make(logrus.Fields, len(fields))
	for _, field := range fields {
		if err, ok := field.Value.(error); ok && err != nil {
			f[field.Key] = err.Error()
		} else {
			f[field.Key] = field.Value
		}
	}
	return f
}
