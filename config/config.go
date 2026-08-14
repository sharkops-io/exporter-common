// Package config binds the flags every exporter in this family shares, each
// with an environment fallback so a Helm values file can stay declarative.
//
// It owns only the common half. Device-shaped configuration — which
// collectors exist, their credentials, their cross-field validation — stays
// in each exporter, because the consumers need genuinely different shapes:
// unas-protect-exporter is a fixed pair of optional collectors, while
// adguard-exporter is an N-instance list. Forcing both through one
// abstraction is the entanglement this extraction was meant to avoid.
package config

import (
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

// ErrHelpRequested is returned by Parse when the user passed --help. It is
// not a configuration error — kingpin has already written the usage text — so
// a caller must check for it with errors.Is and exit 0, rather than treating
// it like every other Parse error (which means "print this and exit 1").
//
// This exists because Parse calls app.Terminate(nil): kingpin's own
// os.Exit(0)-on-help behaviour is disabled on purpose, so that Parse stays
// callable from a test without ending the test binary's process. With
// terminate turned into a no-op, kingpin's Parse falls through --help exactly
// like an empty command line and returns a zero-value config with a nil error
// — indistinguishable, to a caller, from "the user ran this with no flags at
// all", which every exporter here correctly treats as a fatal
// misconfiguration. That check is load-bearing (an exporter with nothing
// configured must fail loudly rather than serve an empty /metrics forever),
// so it cannot simply be relaxed; --help needs its own distinct signal.
var ErrHelpRequested = errors.New("help requested")

// Defaults are the per-exporter values Register cannot guess.
type Defaults struct {
	// ListenAddress differs per exporter (:9101 for unas-protect, :9618 for
	// adguard), so it is a parameter rather than a constant.
	ListenAddress string
}

// Base is the configuration every exporter in this family has.
type Base struct {
	ListenAddress string
	TelemetryPath string

	// Timeout bounds a whole collector scrape; HTTPTimeout bounds each
	// individual request. Two different clocks: raising only the former does
	// nothing, because every request is still capped by the latter.
	Timeout     time.Duration
	HTTPTimeout time.Duration

	LogLevel  slog.Level
	LogFormat string

	logLevel      string
	helpRequested bool
}

// Register adds the common flags to app and returns the Base they populate.
// The caller registers its own flags on the same app before calling Parse.
func Register(app *kingpin.Application, d Defaults) *Base {
	b := &Base{}

	app.Flag("web.listen-address", "Address to listen on.").
		Default(d.ListenAddress).StringVar(&b.ListenAddress)
	app.Flag("web.telemetry-path", "Path under which to expose metrics.").
		Default("/metrics").StringVar(&b.TelemetryPath)
	app.Flag("collector.timeout", "Per-collector timeout for a whole scrape. Keep below Prometheus' scrape_timeout.").
		Default("25s").DurationVar(&b.Timeout)
	app.Flag("http.timeout", "Per-HTTP-request timeout. Raise this too when raising --collector.timeout.").
		Default("10s").DurationVar(&b.HTTPTimeout)
	app.Flag("log.level", "Log level: debug, info, warn or error.").
		Default("info").EnumVar(&b.logLevel, "debug", "info", "warn", "error")
	app.Flag("log.format", "Log output format: logfmt or json.").
		Default("logfmt").EnumVar(&b.LogFormat, "logfmt", "json")

	app.Terminate(nil)
	app.HelpFlag.IsSetByUser(&b.helpRequested)

	return b
}

// Parse runs app against args and finalises the derived fields on base. The
// caller applies its own environment fallbacks and validation afterwards.
func Parse(app *kingpin.Application, base *Base, args []string) error {
	if _, err := app.Parse(args); err != nil {
		return err
	}
	if base.helpRequested {
		return ErrHelpRequested
	}

	// EnumVar already rejected anything else, so this cannot fall silently
	// into a wrong level.
	switch base.logLevel {
	case "debug":
		base.LogLevel = slog.LevelDebug
	case "warn":
		base.LogLevel = slog.LevelWarn
	case "error":
		base.LogLevel = slog.LevelError
	default:
		base.LogLevel = slog.LevelInfo
	}

	return nil
}

// LogHandler builds the slog handler --log.level and --log.format select.
// "logfmt" maps to slog's TextHandler, whose key=value output is the logfmt
// the Prometheus toolkit's own flag of this name produces.
func (b *Base) LogHandler(w io.Writer) slog.Handler {
	opts := &slog.HandlerOptions{Level: b.LogLevel}
	if b.LogFormat == "json" {
		return slog.NewJSONHandler(w, opts)
	}
	return slog.NewTextHandler(w, opts)
}

// Fallback applies an environment value only where the flag left the field
// empty, so a flag always wins.
func Fallback(dst *string, env string) {
	if *dst == "" {
		*dst = env
	}
}
