package log

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var defaultLogger = logrus.New()

const (
	// FormatTimestampDev is the default timestamp layout for the development formatter.
	FormatTimestampDev = "2006/01/02 - 15:04:05"
	// DefaultPrefix is the default signature used by the development formatter.
	DefaultPrefix = "APP"
	// FormatText writes human-readable text. It is the default format.
	FormatText Format = "text"
	// FormatJSON writes structured JSON.
	FormatJSON Format = "json"
	// FormatDev writes structured logs using this package's development layout.
	FormatDev Format = "dev"

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

// Format selects the output format used by the package logger.
type Format string

// Level is the minimum severity that the package logger writes.
type Level string

// Options configures the package-wide logger.
// The zero value selects info-level text output to standard output.
type Options struct {
	// Level is the minimum severity to write. Unknown values use InfoLevel.
	Level Level
	// Format selects FormatText, FormatJSON, or FormatDev. Unknown values use FormatText.
	Format Format
	// Formatter overrides Format and built-in formatter options when non-nil.
	Formatter logrus.Formatter
	// Text contains options used by the built-in text formatter.
	Text TextOptions
	// JSON contains options used by the built-in JSON formatter.
	JSON JSONOptions
	// HideKeys omits field names in development output. By default,
	// field names are shown.
	HideKeys bool
	// Prefix identifies entries formatted by the development formatter. The default is
	// DefaultPrefix.
	Prefix string
	// FieldSeparator separates fields in development output. Its default
	// is " | ".
	FieldSeparator string
	// KeyValueSeparator separates each field key from its value in development
	// text output. Its default is four spaces.
	KeyValueSeparator string
}

// TextOptions configures Logrus's built-in text formatter. TimestampFormat is
// also used by the development formatter when FormatDev is selected.
type TextOptions struct {
	// TimestampFormat is a Go time layout. When empty, FormatDev uses
	// FormatTimestampDev and the built-in text formatter uses time.RFC3339.
	TimestampFormat string
	// ForceColors enables ANSI colors even when output is not a terminal.
	ForceColors bool
	// EnableColors enables colors for text output when supported.
	EnableColors bool
	// ForceQuote quotes all values in text output.
	ForceQuote bool
	// DisableQuote disables quoting values unless ForceQuote is set.
	DisableQuote bool
	// EnvironmentOverrideColors lets CLICOLOR and CLICOLOR_FORCE control colors.
	EnvironmentOverrideColors bool
	// DisableTimestamp omits timestamps from log entries.
	DisableTimestamp bool
	// FullTimestamp prints a full timestamp instead of elapsed time.
	FullTimestamp bool
	// DisableSorting disables sorting of fields.
	DisableSorting bool
	// QuoteEmptyFields quotes empty values.
	QuoteEmptyFields bool
	// SortingFunc customizes field-key sorting.
	SortingFunc func([]string)
}

// JSONOptions configures the built-in JSON formatter.
type JSONOptions struct {
	// TimestampFormat is a Go time layout. The default is time.RFC3339.
	TimestampFormat string
	// DisableTimestamp omits timestamps from log entries.
	DisableTimestamp bool
	// PrettyPrint indents JSON output for readability.
	PrettyPrint bool
}

func init() {
	Config(Options{})
}

// Config replaces the package-wide logger configuration.
// Call it during application startup, before starting goroutines that log.
func Config(opts Options) {
	defaultLogger.SetLevel(withLevel(opts.Level))
	defaultLogger.SetFormatter(withFormatter(opts))
	defaultLogger.SetOutput(os.Stdout)
}

func withFormatter(opts Options) logrus.Formatter {
	if opts.Formatter != nil {
		return opts.Formatter
	}
	if strings.EqualFold(string(opts.Format), string(FormatDev)) {
		textOptions := opts.Text
		if textOptions.TimestampFormat == "" {
			textOptions.TimestampFormat = FormatTimestampDev
		}
		return &Formatter{
			TimestampFormat:   textOptions.TimestampFormat,
			Prefix:            opts.Prefix,
			HideKeys:          opts.HideKeys,
			FieldSeparator:    opts.FieldSeparator,
			KeyValueSeparator: opts.KeyValueSeparator,
		}
	}

	return withFormat(opts)
}

func withLevel(level Level) logrus.Level {
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

func withFormat(opts Options) logrus.Formatter {
	switch strings.ToLower(string(opts.Format)) {
	case string(FormatJSON):
		return &logrus.JSONFormatter{
			TimestampFormat:  timestampFormat(opts.JSON.TimestampFormat),
			DisableTimestamp: opts.JSON.DisableTimestamp,
			PrettyPrint:      opts.JSON.PrettyPrint,
		}
	default:
		return textFormatter(opts.Text)
	}
}

func textFormatter(opts TextOptions) *logrus.TextFormatter {
	return &logrus.TextFormatter{
		TimestampFormat:           timestampFormat(opts.TimestampFormat),
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

func timestampFormat(value string) string {
	if value == "" {
		return time.RFC3339
	}
	return value
}

// ToFormat converts a string to a supported Format.
// Matching is case-insensitive and ignores surrounding whitespace.
// If value is invalid, the first optional fallback is returned when valid;
// otherwise FormatText is returned.
func ToFormat(value string, fallback ...Format) Format {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(FormatText):
		return FormatText
	case string(FormatJSON):
		return FormatJSON
	case string(FormatDev):
		return FormatDev
	}

	if len(fallback) > 0 && isValidFormat(fallback[0]) {
		return fallback[0]
	}

	return FormatText
}

// ToLevel converts a string to a supported Level.
// Matching is case-insensitive and ignores surrounding whitespace.
// If value is invalid, the first optional fallback is returned when valid;
// otherwise InfoLevel is returned.
func ToLevel(value string, fallback ...Level) Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(PanicLevel):
		return PanicLevel
	case string(FatalLevel):
		return FatalLevel
	case string(ErrorLevel):
		return ErrorLevel
	case string(WarnLevel):
		return WarnLevel
	case string(InfoLevel):
		return InfoLevel
	case string(DebugLevel):
		return DebugLevel
	case string(TraceLevel):
		return TraceLevel
	}

	if len(fallback) > 0 && isValidLevel(fallback[0]) {
		return fallback[0]
	}

	return InfoLevel
}

func isValidLevel(level Level) bool {
	switch strings.ToLower(string(level)) {
	case string(PanicLevel), string(FatalLevel), string(ErrorLevel),
		string(WarnLevel), string(InfoLevel), string(DebugLevel), string(TraceLevel):
		return true
	default:
		return false
	}
}

func isValidFormat(format Format) bool {
	switch strings.ToLower(string(format)) {
	case string(FormatText), string(FormatJSON), string(FormatDev):
		return true
	default:
		return false
	}
}

// Formatter formats development entries as
// "[PREFIX] timestamp | LEVEL: message | fields". It is selected by FormatDev.
type Formatter struct {
	// Prefix identifies entries. The zero value uses DefaultPrefix.
	Prefix string
	// TimestampFormat is a Go time layout. The zero value uses
	// FormatTimestampDev.
	TimestampFormat string
	// HideKeys omits field names, leaving only field values.
	HideKeys bool
	// FieldSeparator separates formatted fields. The zero value selects a
	// default of " | ".
	FieldSeparator string
	// KeyValueSeparator separates keys from values when keys are shown.
	// The zero value is four spaces.
	KeyValueSeparator string
}

// Format formats one log entry using the development text layout.
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	prefix := f.Prefix
	if prefix == "" {
		prefix = DefaultPrefix
	}
	layout := f.TimestampFormat
	if layout == "" {
		layout = FormatTimestampDev
	}
	fmt.Fprintf(
		b,
		"[%s] %s | %s: %s",
		prefix,
		entry.Time.Format(layout),
		strings.ToUpper(entry.Level.String()),
		entry.Message,
	)

	if len(entry.Data) == 0 {
		b.WriteByte('\n')
		return b.Bytes(), nil
	}

	keys := make([]string, 0, len(entry.Data))
	for key := range entry.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fields := make([]string, 0, len(keys))
	for _, key := range keys {
		value := fmt.Sprint(entry.Data[key])
		if !f.HideKeys {
			fields = append(fields, key+f.keyValueSeparator()+value)
		} else {
			fields = append(fields, value)
		}
	}

	b.WriteString(" | ")
	b.WriteString(strings.Join(fields, f.fieldSeparator()))
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *Formatter) fieldSeparator() string {
	if f.FieldSeparator != "" {
		return f.FieldSeparator
	}
	return " | "
}

func (f *Formatter) keyValueSeparator() string {
	if f.KeyValueSeparator != "" {
		return f.KeyValueSeparator
	}
	return "    "
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

// Info logs a message and its structured fields at info level.
func Info(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Info(msg)
}

// Infof logs a formatted message at info level.
func Infof(format string, args ...any) {
	defaultLogger.Infof(format, args...)
}

// Infoln logs arguments separated by spaces at info level.
func Infoln(args ...any) {
	defaultLogger.Infoln(args...)
}

// Error logs a message and its structured fields at error level.
func Error(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Error(msg)
}

// Errorf logs a formatted message at error level.
func Errorf(format string, args ...any) {
	defaultLogger.Errorf(format, args...)
}

// Errorln logs arguments separated by spaces at error level.
func Errorln(args ...any) {
	defaultLogger.Errorln(args...)
}

// Trace logs a message and its structured fields at trace level.
func Trace(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Trace(msg)
}

// Tracef logs a formatted message at trace level.
func Tracef(format string, args ...any) {
	defaultLogger.Tracef(format, args...)
}

// Traceln logs arguments separated by spaces at trace level.
func Traceln(args ...any) {
	defaultLogger.Traceln(args...)
}

// Panic logs a message and its structured fields at panic level, then panics.
func Panic(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Panic(msg)
}

// Panicf logs a formatted message at panic level, then panics.
func Panicf(format string, args ...any) {
	defaultLogger.Panicf(format, args...)
}

// Panicln logs arguments separated by spaces at panic level, then panics.
func Panicln(args ...any) {
	defaultLogger.Panicln(args...)
}

// Debug logs a message and its structured fields at debug level.
func Debug(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Debug(msg)
}

// Debugf logs a formatted message at debug level.
func Debugf(format string, args ...any) {
	defaultLogger.Debugf(format, args...)
}

// Debugln logs arguments separated by spaces at debug level.
func Debugln(args ...any) {
	defaultLogger.Debugln(args...)
}

// Print logs a message and its structured fields at info level.
func Print(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Print(msg)
}

// Printf logs a formatted message at info level.
func Printf(format string, args ...any) {
	defaultLogger.Printf(format, args...)
}

// Println logs arguments separated by spaces at info level.
func Println(args ...any) {
	defaultLogger.Println(args...)
}

// Warn logs a message and its structured fields at warning level.
func Warn(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Warn(msg)
}

// Warnf logs a formatted message at warning level.
func Warnf(format string, args ...any) {
	defaultLogger.Warnf(format, args...)
}

// Warnln logs arguments separated by spaces at warning level.
func Warnln(args ...any) {
	defaultLogger.Warnln(args...)
}

// Fatal logs a message and its structured fields at fatal level, then exits.
func Fatal(msg string, fields ...Attribute) {
	defaultLogger.WithFields(toLogrusFields(fields)).Fatal(msg)
}

// Fatalf logs a formatted message at fatal level, then exits.
func Fatalf(format string, args ...any) {
	defaultLogger.Fatalf(format, args...)
}

// Fatalln logs arguments separated by spaces at fatal level, then exits.
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
