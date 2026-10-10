package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func captureLogs(t *testing.T, level Level) *bytes.Buffer {
	t.Helper()

	Config(Options{
		Level:  level,
		Format: FormatJSON,
	})

	var output bytes.Buffer
	defaultLogger.SetOutput(&output)
	t.Cleanup(func() {
		Config(Options{})
	})

	return &output
}

func TestToLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fallback []Level
		want     Level
	}{
		{"panic", "panic", nil, PanicLevel},
		{"fatal", "fatal", nil, FatalLevel},
		{"error", "error", nil, ErrorLevel},
		{"warn", "warn", nil, WarnLevel},
		{"info", "info", nil, InfoLevel},
		{"debug", "debug", nil, DebugLevel},
		{"trace", "trace", nil, TraceLevel},
		{"case and spaces", "  DeBuG  ", nil, DebugLevel},
		{"invalid uses default", "verbose", nil, InfoLevel},
		{"empty uses default", "", nil, InfoLevel},
		{"invalid uses fallback", "verbose", []Level{ErrorLevel}, ErrorLevel},
		{"valid value ignores fallback", "debug", []Level{ErrorLevel}, DebugLevel},
		{"invalid fallback uses default", "verbose", []Level{Level("verbose")}, InfoLevel},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ToLevel(test.input, test.fallback...); got != test.want {
				t.Errorf("ToLevel(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestToFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fallback []Format
		want     Format
	}{
		{"text", "text", nil, FormatText},
		{"json", "json", nil, FormatJSON},
		{"dev", "dev", nil, FormatDev},
		{"case and spaces", "  JsOn  ", nil, FormatJSON},
		{"dev case and spaces", "  DeV  ", nil, FormatDev},
		{"invalid uses default", "yaml", nil, FormatText},
		{"empty uses default", "", nil, FormatText},
		{"invalid uses fallback", "yaml", []Format{FormatJSON}, FormatJSON},
		{"dev fallback", "yaml", []Format{FormatDev}, FormatDev},
		{"valid value ignores fallback", "text", []Format{FormatJSON}, FormatText},
		{"invalid fallback uses default", "yaml", []Format{Format("yaml")}, FormatText},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ToFormat(test.input, test.fallback...); got != test.want {
				t.Errorf("ToFormat(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func TestWithLevel(t *testing.T) {
	tests := []struct {
		input Level
		want  logrus.Level
	}{
		{PanicLevel, logrus.PanicLevel},
		{FatalLevel, logrus.FatalLevel},
		{ErrorLevel, logrus.ErrorLevel},
		{WarnLevel, logrus.WarnLevel},
		{InfoLevel, logrus.InfoLevel},
		{DebugLevel, logrus.DebugLevel},
		{TraceLevel, logrus.TraceLevel},
		{Level("DEBUG"), logrus.DebugLevel},
		{Level("invalid"), logrus.InfoLevel},
		{"", logrus.InfoLevel},
	}

	for _, test := range tests {
		t.Run(string(test.input), func(t *testing.T) {
			if got := withLevel(test.input); got != test.want {
				t.Errorf("getLevel(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	Config(Options{
		Level:  DebugLevel,
		Format: FormatJSON,
	})
	if defaultLogger.GetLevel() != logrus.DebugLevel {
		t.Errorf("logger level = %v, want %v", defaultLogger.GetLevel(), logrus.DebugLevel)
	}
	if _, ok := defaultLogger.Formatter.(*logrus.JSONFormatter); !ok {
		t.Errorf("formatter = %T, want JSONFormatter", defaultLogger.Formatter)
	}
	if defaultLogger.Out != os.Stdout {
		t.Error("Config should set output to os.Stdout")
	}
}

func TestWithFormat(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		formatter, ok := withFormat(Options{
			Format: Format("JSON"),
			JSON: JSONOptions{
				TimestampFormat:  "2006",
				DisableTimestamp: true,
				PrettyPrint:      true,
			},
		}).(*logrus.JSONFormatter)
		if !ok {
			t.Fatalf("formatter type = %T, want JSONFormatter", withFormat(Options{Format: FormatJSON}))
		}
		if formatter.TimestampFormat != "2006" || !formatter.DisableTimestamp || !formatter.PrettyPrint {
			t.Errorf("JSON formatter options not applied: %+v", formatter)
		}
	})

	t.Run("text defaults and options", func(t *testing.T) {
		formatter, ok := withFormat(Options{}).(*logrus.TextFormatter)
		if !ok {
			t.Fatalf("formatter type = %T, want TextFormatter", withFormat(Options{}))
		}
		if formatter.TimestampFormat == "" || !formatter.DisableColors {
			t.Errorf("unexpected text formatter defaults: %+v", formatter)
		}

		sortKeys := func(keys []string) {}
		opts := Options{
			Format: FormatText,
			Text: TextOptions{
				TimestampFormat:           "2006",
				ForceColors:               true,
				EnableColors:              true,
				ForceQuote:                true,
				DisableQuote:              true,
				EnvironmentOverrideColors: true,
				DisableTimestamp:          true,
				FullTimestamp:             true,
				DisableSorting:            true,
				QuoteEmptyFields:          true,
				SortingFunc:               sortKeys,
			},
		}
		formatter, ok = withFormat(opts).(*logrus.TextFormatter)
		if !ok {
			t.Fatalf("formatter type = %T, want TextFormatter", withFormat(opts))
		}
		if formatter.TimestampFormat != opts.Text.TimestampFormat ||
			!formatter.ForceColors ||
			formatter.DisableColors ||
			!formatter.ForceQuote ||
			!formatter.DisableQuote ||
			!formatter.EnvironmentOverrideColors ||
			!formatter.DisableTimestamp ||
			!formatter.FullTimestamp ||
			!formatter.DisableSorting ||
			!formatter.QuoteEmptyFields ||
			formatter.SortingFunc == nil {
			t.Errorf("text formatter options not applied: %+v", formatter)
		}

		defaultTimestamp := withFormat(Options{Format: FormatText}).(*logrus.TextFormatter)
		if defaultTimestamp.TimestampFormat != "2006-01-02T15:04:05Z07:00" {
			t.Errorf("default text timestamp format = %q", defaultTimestamp.TimestampFormat)
		}
	})
}

func TestNestedFormatterOptions(t *testing.T) {
	t.Run("JSON options", func(t *testing.T) {
		got, ok := withFormat(Options{
			Format: FormatJSON,
			JSON: JSONOptions{
				TimestampFormat:  "2006",
				DisableTimestamp: true,
				PrettyPrint:      true,
			},
		}).(*logrus.JSONFormatter)
		if !ok {
			t.Fatalf("formatter type = %T, want JSONFormatter", withFormat(Options{Format: FormatJSON}))
		}
		if got.TimestampFormat != "2006" || !got.DisableTimestamp || !got.PrettyPrint {
			t.Errorf("nested JSON options not applied: %+v", got)
		}
	})

	t.Run("text options", func(t *testing.T) {
		got, ok := withFormat(Options{
			Format: FormatText,
			Text: TextOptions{
				TimestampFormat: "2006",
				EnableColors:    true,
				FullTimestamp:   true,
			},
		}).(*logrus.TextFormatter)
		if !ok {
			t.Fatalf("formatter type = %T, want TextFormatter", withFormat(Options{Format: FormatText}))
		}
		if got.TimestampFormat != "2006" || got.DisableColors || !got.FullTimestamp {
			t.Errorf("nested text options not applied: %+v", got)
		}
	})
}

func TestCustomFormatterOverridesBuiltIn(t *testing.T) {
	custom := &logrus.JSONFormatter{PrettyPrint: true}
	got := withFormatter(Options{
		Format:    FormatText,
		Formatter: custom,
		HideKeys:  true,
	})
	if got != custom {
		t.Errorf("formatter = %T, want provided custom formatter", got)
	}
}

func TestWithFormatterSelectsConfiguredFormat(t *testing.T) {
	tests := []struct {
		name       string
		format     Format
		wantFormat string
	}{
		{"text", FormatText, "text"},
		{"json", FormatJSON, "json"},
		{"dev", FormatDev, "dev"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := withFormatter(Options{Format: test.format})
			switch test.wantFormat {
			case "json":
				if _, ok := got.(*logrus.JSONFormatter); !ok {
					t.Errorf("formatter type = %T, want JSONFormatter", got)
				}
			case "text":
				if _, ok := got.(*logrus.TextFormatter); !ok {
					t.Errorf("formatter type = %T, want TextFormatter", got)
				}
			case "dev":
				if _, ok := got.(*Formatter); !ok {
					t.Errorf("formatter type = %T, want development Formatter", got)
				}
			}
		})
	}
}

func TestDevelopmentFormatterOutputAndSeparators(t *testing.T) {
	opts := Options{
		Format: FormatDev,
		Text: TextOptions{
			TimestampFormat: "2006-01-02 15:04:05",
		},
	}
	entry := logrus.NewEntry(logrus.New())
	entry.Time = time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	entry.Level = logrus.InfoLevel
	entry.Message = "request completed"
	entry.Data = logrus.Fields{
		"z": "two words",
		"a": "first",
	}

	output, err := withFormatter(opts).Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	got := string(output)
	want := `[APP] 2025-01-02 03:04:05 | INFO: request completed | a    first | z    two words` + "\n"
	if got != want {
		t.Errorf("default keyed output = %q, want %q", got, want)
	}

	opts.HideKeys = true
	output, err = withFormatter(opts).Format(entry)
	if err != nil {
		t.Fatalf("Format() with HideKeys error = %v", err)
	}
	want = `[APP] 2025-01-02 03:04:05 | INFO: request completed | first | two words` + "\n"
	if got = string(output); got != want {
		t.Errorf("values-only output = %q, want %q", got, want)
	}

	opts.HideKeys = false
	opts.FieldSeparator = "; "
	opts.KeyValueSeparator = " -> "
	output, err = withFormatter(opts).Format(entry)
	if err != nil {
		t.Fatalf("Format() with custom separators error = %v", err)
	}
	want = `[APP] 2025-01-02 03:04:05 | INFO: request completed | a -> first; z -> two words` + "\n"
	if got = string(output); got != want {
		t.Errorf("custom separator output = %q, want %q", got, want)
	}

	opts.HideKeys = true
	output, err = withFormatter(opts).Format(entry)
	if err != nil {
		t.Fatalf("Format() with custom field separator error = %v", err)
	}
	want = `[APP] 2025-01-02 03:04:05 | INFO: request completed | first; two words` + "\n"
	if got = string(output); got != want {
		t.Errorf("custom values separator output = %q, want %q", got, want)
	}
}

func TestDevelopmentFormatterUsesDefaultTimestampLayout(t *testing.T) {
	got := withFormatter(Options{Format: FormatDev}).(*Formatter)
	if got.TimestampFormat != FormatTimestampDev {
		t.Errorf("development timestamp format = %q, want %q", got.TimestampFormat, FormatTimestampDev)
	}
	entry := logrus.NewEntry(logrus.New())
	entry.Time = time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	entry.Level = logrus.InfoLevel
	entry.Message = "started"

	output, err := got.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	want := "[APP] 2025/01/02 - 03:04:05 | INFO: started\n"
	if string(output) != want {
		t.Errorf("default local output = %q, want %q", output, want)
	}

	got.Prefix = "API"
	output, err = got.Format(entry)
	if err != nil {
		t.Fatalf("Format() with custom prefix error = %v", err)
	}
	want = "[API] 2025/01/02 - 03:04:05 | INFO: started\n"
	if string(output) != want {
		t.Errorf("custom prefix output = %q, want %q", output, want)
	}
}

func TestFieldHelpersAndConversion(t *testing.T) {
	err := errors.New("database unavailable")
	if got := Err(err); got.Key != "error" || got.Value != err {
		t.Errorf("Err() = %#v", got)
	}
	if got := Field("RequestID", "abc"); got.Key != "RequestID" || got.Value != "abc" {
		t.Errorf("Field() = %#v", got)
	}

	fields := toLogrusFields([]Attribute{
		Err(err),
		Field("nil_error", error(nil)),
		Field("number", 42),
	})
	if fields["error"] != err.Error() || fields["nil_error"] != error(nil) || fields["number"] != 42 {
		t.Errorf("converted fields = %#v", fields)
	}
	if got := toLogrusFields(nil); len(got) != 0 {
		t.Errorf("conversion of nil fields = %#v, want empty map", got)
	}
}

func TestNonTerminatingLoggingMethods(t *testing.T) {
	output := captureLogs(t, TraceLevel)
	methods := []struct {
		name string
		call func()
	}{
		{"Trace", func() { Trace("trace message", Field("method", "Trace")) }},
		{"Tracef", func() { Tracef("trace %s", "formatted") }},
		{"Traceln", func() { Traceln("trace", "line") }},
		{"Debug", func() { Debug("debug message", Field("method", "Debug")) }},
		{"Debugf", func() { Debugf("debug %s", "formatted") }},
		{"Debugln", func() { Debugln("debug", "line") }},
		{"Info", func() { Info("info message", Field("method", "Info")) }},
		{"Infof", func() { Infof("info %s", "formatted") }},
		{"Infoln", func() { Infoln("info", "line") }},
		{"Print", func() { Print("print message", Field("method", "Print")) }},
		{"Printf", func() { Printf("print %s", "formatted") }},
		{"Println", func() { Println("print", "line") }},
		{"Warn", func() { Warn("warn message", Field("method", "Warn")) }},
		{"Warnf", func() { Warnf("warn %s", "formatted") }},
		{"Warnln", func() { Warnln("warn", "line") }},
		{"Error", func() { Error("error message", Field("method", "Error")) }},
		{"Errorf", func() { Errorf("error %s", "formatted") }},
		{"Errorln", func() { Errorln("error", "line") }},
	}

	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			before := output.Len()
			method.call()
			if output.Len() <= before {
				t.Errorf("%s did not write a log entry", method.name)
			}
		})
	}
}

func TestPanicLoggingMethods(t *testing.T) {
	methods := []struct {
		name string
		call func()
		want string
	}{
		{"Panic", func() { Panic("panic message", Field("kind", "plain")) }, "panic message"},
		{"Panicf", func() { Panicf("panic %s", "formatted") }, "panic formatted"},
		{"Panicln", func() { Panicln("panic", "line") }, "panic line"},
	}

	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			output := captureLogs(t, TraceLevel)
			func() {
				defer func() {
					if got := recover(); got == nil {
						t.Error("expected panic")
					}
				}()
				method.call()
			}()
			if !strings.Contains(output.String(), method.want) {
				t.Errorf("log output %q does not contain %q", output.String(), method.want)
			}
		})
	}
}

func TestFatalLoggingMethods(t *testing.T) {
	methods := []struct {
		name string
		call func()
		want string
	}{
		{"Fatal", func() { Fatal("fatal message", Field("kind", "plain")) }, "fatal message"},
		{"Fatalf", func() { Fatalf("fatal %s", "formatted") }, "fatal formatted"},
		{"Fatalln", func() { Fatalln("fatal", "line") }, "fatal line"},
	}

	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			output := captureLogs(t, TraceLevel)
			previousExitFunc := defaultLogger.ExitFunc
			exitCode := 0
			defaultLogger.ExitFunc = func(code int) {
				exitCode = code
			}
			t.Cleanup(func() {
				defaultLogger.ExitFunc = previousExitFunc
			})

			method.call()
			if exitCode != 1 {
				t.Errorf("exit code = %d, want 1", exitCode)
			}
			if !strings.Contains(output.String(), method.want) {
				t.Errorf("log output %q does not contain %q", output.String(), method.want)
			}
		})
	}
}

func TestStructuredFieldsAreEncodedAsJSON(t *testing.T) {
	output := captureLogs(t, InfoLevel)
	Info("request completed",
		Field("request_id", "abc-123"),
		Field("status", 200),
		Err(errors.New("not found")),
	)

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
		t.Fatalf("invalid JSON log entry: %v", err)
	}
	if entry["msg"] != "request completed" ||
		entry["request_id"] != "abc-123" ||
		entry["status"] != float64(200) ||
		entry["error"] != "not found" {
		t.Errorf("unexpected JSON entry: %s", output.String())
	}
}

func TestConfigLevelFiltersMessages(t *testing.T) {
	output := captureLogs(t, WarnLevel)
	Debug("debug should be filtered")
	Info("info should be filtered")
	Warn("warning should be written")
	Error("error should be written")

	logs := output.String()
	if strings.Contains(logs, "should be filtered") {
		t.Errorf("below-threshold messages were logged: %s", logs)
	}
	if !strings.Contains(logs, "warning should be written") ||
		!strings.Contains(logs, "error should be written") {
		t.Errorf("expected warning and error entries, got: %s", logs)
	}
}

func TestLogLineFormattingMethods(t *testing.T) {
	output := captureLogs(t, InfoLevel)
	Println("item", 12)
	if !strings.Contains(output.String(), "item 12") {
		t.Errorf("Println output = %q", output.String())
	}

	output.Reset()
	Printf("port %d", 8080)
	if !strings.Contains(output.String(), "port 8080") {
		t.Errorf("Printf output = %q", output.String())
	}
}

func TestAnyAcceptsDifferentValueTypes(t *testing.T) {
	fields := toLogrusFields([]Attribute{
		Field("string", "value"),
		Field("integer", 42),
		Field("boolean", true),
	})
	if fields["string"] != "value" ||
		fields["integer"] != 42 ||
		fields["boolean"] != true {
		t.Errorf("unexpected field values: %#v", fields)
	}
}
