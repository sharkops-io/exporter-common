# exporter-common

Shared plumbing for `sharkops/exporters/*` Prometheus exporters. It exists so
that a second exporter can reuse this logic instead of copying it — the
packages here started as `internal/*` inside
[`unas-protect-exporter`](https://gitlab.com/sharkops/exporters/unas-protect-exporter)
and were promoted to a standalone module because `internal/` cannot be
imported across a module boundary.

## Packages

- **`httpx`** — a small `*http.Client` builder plus a `GetBytes` helper that
  wraps the do/status-check/bounded-read/close boilerplate every hand-rolled
  device HTTP client needs, including the response-size cap and TLS
  transport handling for self-signed local targets.
- **`metric`** — `Sample`, the one intermediate representation between a
  device payload and Prometheus exposition (`Gauge`, `Counter`,
  `Sample.Metric()`, `Sample.Signature()`), plus the metric-name validation
  and duplicate-series detection that goes with it.
- **`collector`** — adapts a `Source` (`Namespace() string`,
  `Collect(ctx) ([]metric.Sample, error)`) to `prometheus.Collector`, with a
  per-collector timeout and the `<namespace>_scrape_collector_success` /
  `_duration_seconds` self-monitoring series.
- **`testutil`** — `MetricSlice`, a fixed-slice `prometheus.Collector` for
  asserting exposition output without any collector machinery.

## What didn't move

`config` was deliberately **not** extracted. Its generic half (flag/env
parsing, precedence rules) is entangled with a struct whose fields include
`UNASEnabled` and `Protect protect.Config` — device-specific concerns baked
into what would otherwise be shared code. Rather than guess at the seam
between "generic config plumbing" and "unas-protect-exporter-specific
config," that split should be revealed by a second consumer actually needing
it.

## Compatibility

Consumers should track the same pinned versions of
`github.com/prometheus/client_golang` and `github.com/prometheus/common` this
module uses (see `go.mod`) — `metric.Sample.Metric()` relies on
version-specific validation behaviour documented in `metric/sample.go`.
