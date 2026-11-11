# exporter-common

Shared plumbing for the [sharkops-io](https://github.com/sharkops-io) Prometheus exporters. It exists so
that a second exporter can reuse this logic instead of copying it — the
packages here started as `internal/*` inside
[`unas-protect-exporter`](https://github.com/sharkops-io/unas-protect-exporter)
and were promoted to a standalone module because `internal/` cannot be
imported across a module boundary.

```sh
go get github.com/sharkops-io/exporter-common@latest
```

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
- **`config`** — the flag/env plumbing every exporter shares: common flags,
  `--help` as a distinct signal, log handler construction, and the
  flag-beats-environment precedence rule. Device-shaped config stays local to
  each exporter.
- **`testutil`** — `MetricSlice`, a fixed-slice `prometheus.Collector` for
  asserting exposition output without any collector machinery.

## What moved, and what didn't

`config` was extracted in v0.2.0, when `adguard-exporter` became the second
consumer and revealed the seam the earlier note was waiting for. The package
owns the flags every exporter shares — `--web.listen-address`,
`--web.telemetry-path`, `--collector.timeout`, `--http.timeout`,
`--log.level`, `--log.format` — plus `ErrHelpRequested`, the log-level
mapping, `LogHandler`, and the flag-beats-environment `Fallback` helper.

The **device half stays in each exporter**, deliberately. The two consumers
need different shapes: `unas-protect-exporter` is a fixed pair of optional
collectors with cross-field credential validation, `adguard-exporter` is an
N-instance list. Generalising over both would recreate exactly the
entanglement that kept this package unextracted in the first place.

## Compatibility

Consumers should track the same pinned versions of
`github.com/prometheus/client_golang` and `github.com/prometheus/common` this
module uses (see `go.mod`) — `metric.Sample.Metric()` relies on
version-specific validation behaviour documented in `metric/sample.go`.

## License

Apache License 2.0. See [LICENSE](LICENSE).
