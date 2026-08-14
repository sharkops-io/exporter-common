package config_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"

	"gitlab.com/sharkops/exporters/exporter-common/config"
)

func newApp() *kingpin.Application {
	app := kingpin.New("test-exporter", "test")
	app.UsageWriter(&bytes.Buffer{})
	app.ErrorWriter(&bytes.Buffer{})
	return app
}

func TestDefaultsApply(t *testing.T) {
	app := newApp()
	base := config.Register(app, config.Defaults{ListenAddress: ":9618"})
	if err := config.Parse(app, base, nil); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if base.ListenAddress != ":9618" {
		t.Errorf("ListenAddress = %q, want :9618", base.ListenAddress)
	}
	if base.TelemetryPath != "/metrics" {
		t.Errorf("TelemetryPath = %q, want /metrics", base.TelemetryPath)
	}
	if base.Timeout != 25*time.Second {
		t.Errorf("Timeout = %v, want 25s", base.Timeout)
	}
	if base.HTTPTimeout != 10*time.Second {
		t.Errorf("HTTPTimeout = %v, want 10s", base.HTTPTimeout)
	}
	if base.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want info", base.LogLevel)
	}
}

func TestFlagsOverrideDefaults(t *testing.T) {
	app := newApp()
	base := config.Register(app, config.Defaults{ListenAddress: ":9618"})
	args := []string{"--web.listen-address=:1234", "--log.level=debug", "--collector.timeout=90s"}
	if err := config.Parse(app, base, args); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if base.ListenAddress != ":1234" {
		t.Errorf("ListenAddress = %q, want :1234", base.ListenAddress)
	}
	if base.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want debug", base.LogLevel)
	}
	if base.Timeout != 90*time.Second {
		t.Errorf("Timeout = %v, want 90s", base.Timeout)
	}
}

// --help must be distinguishable from an empty command line. Parse disables
// kingpin's own os.Exit(0) so it stays callable from a test, which makes
// --help otherwise return a zero-value config and a nil error — identical, to
// a caller, to "ran with no flags", which callers treat as fatal.
func TestHelpRequestedIsItsOwnSignal(t *testing.T) {
	app := newApp()
	base := config.Register(app, config.Defaults{ListenAddress: ":9618"})
	err := config.Parse(app, base, []string{"--help"})
	if !errors.Is(err, config.ErrHelpRequested) {
		t.Fatalf("Parse(--help) = %v, want ErrHelpRequested", err)
	}
}

func TestInvalidLogLevelRejected(t *testing.T) {
	app := newApp()
	base := config.Register(app, config.Defaults{ListenAddress: ":9618"})
	if err := config.Parse(app, base, []string{"--log.level=chatty"}); err == nil {
		t.Fatal("Parse accepted an invalid log level")
	}
}

func TestFallbackOnlyFillsEmpty(t *testing.T) {
	set := "from-flag"
	config.Fallback(&set, "from-env")
	if set != "from-flag" {
		t.Errorf("Fallback overwrote a set value: %q", set)
	}
	empty := ""
	config.Fallback(&empty, "from-env")
	if empty != "from-env" {
		t.Errorf("Fallback did not fill an empty value: %q", empty)
	}
}

func TestLogHandlerFormats(t *testing.T) {
	for _, tc := range []struct {
		format string
		want   string
	}{
		{"logfmt", "msg=hello"},
		{"json", `"msg":"hello"`},
	} {
		app := newApp()
		base := config.Register(app, config.Defaults{ListenAddress: ":9618"})
		if err := config.Parse(app, base, []string{"--log.format=" + tc.format}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		var buf bytes.Buffer
		slog.New(base.LogHandler(&buf)).Info("hello")
		if !strings.Contains(buf.String(), tc.want) {
			t.Errorf("format %s: output %q missing %q", tc.format, buf.String(), tc.want)
		}
	}
}
