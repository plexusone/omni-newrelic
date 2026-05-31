# Omni-NewRelic

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/plexusone/omni-newrelic/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/plexusone/omni-newrelic/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/plexusone/omni-newrelic/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/plexusone/omni-newrelic/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/plexusone/omni-newrelic/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/plexusone/omni-newrelic/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/plexusone/omni-newrelic
 [goreport-url]: https://goreportcard.com/report/github.com/plexusone/omni-newrelic
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/plexusone/omni-newrelic
 [docs-godoc-url]: https://pkg.go.dev/github.com/plexusone/omni-newrelic
 [docs-mkdoc-svg]: https://img.shields.io/badge/docs-guide-blue.svg
 [docs-mkdoc-url]: https://plexusone.github.io/omni-newrelic
 [viz-svg]: https://img.shields.io/badge/repo-visualization-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=plexusone%2Fomni-newrelic
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/plexusone/omni-newrelic/blob/main/LICENSE

New Relic integrations for PlexusOne abstraction layers.

## Overview

`omni-newrelic` wraps the official [newrelic-client-go](https://github.com/newrelic/newrelic-client-go) SDK and provides implementations for multiple PlexusOne packages:

| Package | Status | Description |
|---------|--------|-------------|
| `omnisignal` | ✅ Implemented | Alert and incident ingestion |
| `omniobserve` | ✅ Implemented | Telemetry export and data querying |

## Installation

```bash
go get github.com/plexusone/omni-newrelic
```

## Usage

### Signal Ingestion (omnisignal)

```go
import (
    "github.com/plexusone/omnisignal"
    _ "github.com/plexusone/omni-newrelic/omnisignal" // Register New Relic provider
)

func main() {
    provider, err := omnisignal.New("newrelic", omnisignal.Config{
        APIKey: os.Getenv("NEW_RELIC_API_KEY"),
        Options: map[string]any{
            "account_id": 12345,
            "region":     "US", // or "EU"
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Close()

    // Fetch incidents from the last 24 hours
    signals, err := provider.Fetch(ctx, omnisignal.FetchOptions{
        Since: time.Now().Add(-24 * time.Hour),
    })
    if err != nil {
        log.Fatal(err)
    }

    for _, sig := range signals {
        fmt.Printf("Signal: %s - %s\n", sig.ID, sig.Summary)
    }
}
```

### Telemetry Export (omniobserve)

Import the package to register the New Relic telemetry provider:

```go
import (
    "github.com/plexusone/omniobserve/observops"
    _ "github.com/plexusone/omni-newrelic/omniobserve" // Register New Relic provider
)

func main() {
    provider, err := observops.Open("newrelic",
        observops.WithAPIKey(os.Getenv("NEW_RELIC_LICENSE_KEY")),
        observops.WithServiceName("my-service"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(context.Background())

    // Create metrics
    counter, _ := provider.Meter().Counter("requests_total")
    counter.Add(ctx, 1)

    // Create traces
    ctx, span := provider.Tracer().Start(ctx, "ProcessRequest")
    defer span.End()
}
```

### Data Querying (omniobserve)

Query observability data from New Relic using NRQL, APM APIs, and entity search:

```go
import "github.com/plexusone/omni-newrelic/omniobserve"

func main() {
    qp, err := omniobserve.NewQueryProvider(omniobserve.QueryConfig{
        APIKey:    os.Getenv("NEW_RELIC_API_KEY"),
        AccountID: 12345,
        Region:    "US",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer qp.Close()

    // Execute NRQL queries
    result, err := qp.QueryNRQL(ctx, "SELECT count(*) FROM Transaction SINCE 1 hour ago")

    // List APM applications
    apps, err := qp.ListAPMApplications(ctx)

    // Search entities
    entities, err := qp.SearchEntities(ctx, omniobserve.EntitySearchOptions{
        Domain: "APM",
        Type:   "APPLICATION",
    })

    // Query logs
    logs, err := qp.QueryLogs(ctx, "level='ERROR'", 100, time.Hour)

    // Query distributed traces
    traces, err := qp.QueryTraces(ctx, "name='WebTransaction'", 50, time.Hour)
}
```

## Configuration

### omnisignal

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `APIKey` | string | Yes | New Relic Personal API Key |
| `account_id` | int | Yes | New Relic Account ID |
| `region` | string | No | "US" (default) or "EU" |

### omniobserve (Telemetry Export)

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `WithAPIKey` | string | Yes | New Relic License Key (Ingest) |
| `WithServiceName` | string | Yes | Service name for telemetry |
| `WithNewRelicRegion` | string | No | "us" (default) or "eu" |

### omniobserve (Data Query)

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `APIKey` | string | Yes | New Relic Personal API Key (User key) |
| `AccountID` | int | Yes | New Relic Account ID |
| `Region` | string | No | "US" (default) or "EU" |

## Capabilities

### omnisignal Provider

- **Fetch**: List incidents with filtering by time and status
- **Acknowledge**: Acknowledge open incidents
- **Close**: Close open incidents
- **Streaming**: Not supported (requires webhook configuration)

### omniobserve Provider

**Telemetry Export:**

- Metrics (counters, gauges, histograms)
- Distributed tracing
- Structured logging with trace correlation

**Data Query:**

- NRQL query execution
- APM application listing and metrics
- Entity search across all domains
- Log querying via NRQL
- Distributed trace querying via NRQL

## Related Packages

- [omnisignal](https://github.com/plexusone/omnisignal) - Signal ingestion abstraction
- [omniobserve](https://github.com/plexusone/omniobserve) - Observability abstraction
- [signal-spec](https://github.com/plexusone/signal-spec) - Canonical signal data model
- [newrelic-client-go](https://github.com/newrelic/newrelic-client-go) - Official New Relic Go SDK

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history and [RELEASE_NOTES_v0.1.0.md](RELEASE_NOTES_v0.1.0.md) for detailed release notes.

## License

MIT
