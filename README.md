<div align="center">

# go-logger

**A small, structured logging wrapper for Go applications.**

[![Go Reference](https://pkg.go.dev/badge/github.com/pinedadaniel/go-logger.svg)](https://pkg.go.dev/github.com/pinedadaniel/go-logger)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

`go-logger` provides package-level logging functions, configurable text or JSON
output, and structured key-value fields. It uses Logrus internally and starts
with a usable default configuration.

## Features

- Log immediately with a package-wide logger.
- Configure severity and text or JSON output at application startup.
- Attach structured fields to log entries.
- Use plain, formatted (`f`), or space-separated (`ln`) message methods.
- Choose from trace, debug, info, warn, error, panic, and fatal levels.

## Installation

```sh
go get github.com/pinedadaniel/go-logger
```

## Quick start

```go
package main

import (
	"errors"

	log "github.com/pinedadaniel/go-logger/pkg/log"
)

func main() {
	// Configure once, before starting goroutines that log.
	log.Config(log.Options{
		Level:  log.InfoLevel,
		Format: log.FormatJSON,
	})

	log.Info("service started",
		log.Field("service", "orders"),
		log.Field("port", 8080),
	)

	if err := run(); err != nil {
		log.Error("service failed", log.Err(err))
	}
}

func run() error {
	return errors.New("example failure")
}
```

`Config` is optional. Without it, logging uses text output, the `info` level,
and `os.Stdout`.

## Configuration

Configure the logger in `main`, after loading application configuration and
before starting the server or other goroutines:

```go
package main

import (
	"time"

	log "github.com/pinedadaniel/go-logger/pkg/log"
)

func configureLogging() {
	log.Config(log.Options{
		Level:  log.DebugLevel,
		Format: log.FormatText,
		Text: log.TextOptions{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
			EnableColors:    true,
		},
	})
}
```

### Options

| Option | Default | Description |
| --- | --- | --- |
| `Level` | `InfoLevel` | Minimum severity to log. |
| `Format` | `FormatText` | `FormatText` or `FormatJSON`. Unknown values use text. |
| `Formatter` | `nil` | Optional custom Logrus formatter; takes precedence over `Format`. |
| `Text` | zero value | Options for the built-in text formatter. |
| `JSON` | zero value | Options for the built-in JSON formatter. |
| `IsLocal` | `false` | Use the package's custom text formatter instead of the built-in formatter selected by `Format`. |
| `ShowKeys` | `false` | Include field names in local custom text output. By default, only field values are shown. |
| `FieldSeparator` | `", "` or `" "` | Separator between local fields; default is `", "` when keys are hidden and `" "` when shown. |
| `KeyValueSeparator` | `"="` | Separator between a local custom formatter key and value. |

`TextOptions` groups text-only settings: `TimestampFormat`, `ForceColors`,
`EnableColors`, `ForceQuote`, `DisableQuote`, `EnvironmentOverrideColors`,
`DisableTimestamp`, `FullTimestamp`, `DisableSorting`, `QuoteEmptyFields`,
and `SortingFunc`. `JSONOptions` provides `TimestampFormat`, `DisableTimestamp`,
and `PrettyPrint`. The local package formatter uses `Text.TimestampFormat` and
the top-level `ShowKeys`, `FieldSeparator`, and `KeyValueSeparator`; other text
options apply only to Logrus's built-in text formatter.

If `Formatter` is non-nil, it is used as-is and the built-in `Format`, `Text`,
and `JSON` settings are ignored. Otherwise, `IsLocal` selects the package's
custom text formatter. With the default `ShowKeys: false`, fields are values
only:

```text
[2025-01-02 03:04:05] INFO: request completed | 1024, 127.0.0.1, SUCCESS
```

Setting `ShowKeys: true` includes keys:

```text
[2025-01-02 03:04:05] INFO: request completed | id=1024 ip=127.0.0.1 status=SUCCESS
```

The local formatter also allows customizing both separators:

```go
log.Config(log.Options{
	Level:             log.InfoLevel,
	IsLocal:           true,
	ShowKeys:          true,
	FieldSeparator:     ", ",
	KeyValueSeparator: ": ",
	Text: log.TextOptions{
		TimestampFormat: "2006-01-02 15:04:05",
	},
})

log.Info("request completed",
	log.Field("id", 1024),
	log.Field("ip", "127.0.0.1"),
	log.Field("status", "SUCCESS"),
)
// [2025-01-02 03:04:05] INFO: request completed | id: 1024, ip: 127.0.0.1, status: SUCCESS
```

Non-local configuration selects Logrus's text or JSON formatter according to
`Format`.

Supported levels, from most severe to most verbose, are `PanicLevel`,
`FatalLevel`, `ErrorLevel`, `WarnLevel`, `InfoLevel`, `DebugLevel`, and
`TraceLevel`. An empty or unrecognized level uses `InfoLevel`. A configured
level is a threshold: for example, `WarnLevel` includes warning and error
entries, while filtering out info and debug entries.

### Configure from environment variables

Use `ToLevel` and `ToFormat` to convert environment strings without handling
errors. Both functions match values case-insensitively, ignore
surrounding whitespace, and accept an optional fallback. If the input and
fallback are both invalid, they use `InfoLevel` and `FormatText`, respectively.

```go
package main

import (
	"os"

	log "github.com/pinedadaniel/go-logger/pkg/log"
)

func configureLogging() {
	level := log.ToLevel(os.Getenv("APP_LOG_LEVEL"), log.InfoLevel)
	format := log.ToFormat(os.Getenv("APP_LOG_FORMAT"), log.FormatText)

	log.Config(log.Options{
		Level:  level,
		Format: format,
	})
}
```

`Config` updates one logger shared by the process. Set it once during startup;
do not reconfigure it while other goroutines are logging.

## Structured fields

`Field(key, value)` creates an attribute with the exact key and any value type.
`Err(err)` creates an `"error"` attribute. Error values are written as their error
message.

```go
log.Info("request completed",
	log.Field("request_id", "req-123"),
	log.Field("method", "GET"),
	log.Field("status", 200),
	log.Field("cached", false),
)

if err := saveRecord(); err != nil {
	log.Error("saving record failed",
		log.Err(err),
		log.Field("record_id", 42),
	)
}
```

With JSON output, an entry has a message, level, timestamp (unless disabled),
and the supplied fields. For example:

```json
{"level":"info","msg":"request completed","request_id":"req-123","status":200,"time":"2026-10-09T12:00:00Z"}
```

## Logging methods

Each severity offers three styles:

```go
log.Info("connected", log.Field("host", "db.internal")) // message plus fields
log.Infof("connected to %s", "db.internal")           // fmt-style formatting
log.Infoln("connected to", "db.internal")              // arguments separated by spaces
```

The same styles are available for `Trace`, `Debug`, `Info`, `Warn`, and
`Error`. `Panic`, `Fatal`, and `Print` also provide the corresponding
non-structured, formatted, and line methods. `Print`, `Printf`, and `Println`
log at info level.

`Panic*` logs and then panics. `Fatal*` logs and exits the process with a
non-zero status. Use these methods only when those effects are intended.

## License

This project is licensed under the [MIT License](LICENSE).
