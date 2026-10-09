<div align="center">

# logger-go

**A small, structured logging wrapper for Go applications.**

[![Go Reference](https://pkg.go.dev/badge/github.com/pinedadaniel/logger-go.svg)](https://pkg.go.dev/github.com/pinedadaniel/logger-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

`logger-go` provides package-level logging functions, configurable text or JSON
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
go get github.com/pinedadaniel/logger-go
```

## Quick start

```go
package main

import (
	"errors"

	log "github.com/pinedadaniel/logger-go/pkg/log"
)

func main() {
	// Configure once, before starting goroutines that log.
	log.Config(log.LoggerOptions{
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

	log "github.com/pinedadaniel/logger-go/pkg/log"
)

func configureLogging() {
	log.Config(log.LoggerOptions{
		Level:           log.DebugLevel,
		Format:          log.FormatText,
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		EnableColors:    true,
	})
}
```

### Options

| Option | Default | Description |
| --- | --- | --- |
| `Level` | `InfoLevel` | Minimum severity to log. |
| `Format` | `FormatText` | `FormatText` or `FormatJSON`. Unknown values use text. |
| `TimestampFormat` | `time.RFC3339` | Go time layout used for timestamps. |
| `ForceColors` | `false` | Force ANSI colors even when output is not a terminal. |
| `EnableColors` | `false` | Enable colors in text output when supported. |
| `ForceQuote` | `false` | Quote all text field values. |
| `DisableQuote` | `false` | Disable quoting text field values unless `ForceQuote` is enabled. |
| `EnvironmentOverrideColors` | `false` | Let `CLICOLOR` and `CLICOLOR_FORCE` control text colors. |
| `DisableTimestamp` | `false` | Omit timestamps. |
| `FullTimestamp` | `false` | Show a full timestamp in text output. |
| `DisableSorting` | `false` | Disable sorting of text fields. |
| `QuoteEmptyFields` | `false` | Quote empty text field values. |
| `PrettyPrint` | `false` | Indent JSON output for readability. |
| `SortingFunc` | default | Custom text-field key sorting function. |

Supported levels, from most severe to most verbose, are `PanicLevel`,
`FatalLevel`, `ErrorLevel`, `WarnLevel`, `InfoLevel`, `DebugLevel`, and
`TraceLevel`. An empty or unrecognized level uses `InfoLevel`. A configured
level is a threshold: for example, `WarnLevel` includes warning and error
entries, while filtering out info and debug entries.

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
