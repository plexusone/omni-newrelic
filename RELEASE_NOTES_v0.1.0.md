# Release Notes - v0.1.0

**Release Date:** 2026-05-31

## Overview

This is the initial release of `omni-newrelic`, providing New Relic integrations for PlexusOne abstraction layers. The package wraps the official [newrelic-client-go](https://github.com/newrelic/newrelic-client-go) SDK and provides implementations for both `omnisignal` and `omniobserve`.

## Highlights

- **Signal Provider (omnisignal):** Alert and incident ingestion with support for fetching, acknowledging, and closing incidents
- **Observability Provider (omniobserve):** Dual-capability provider supporting both telemetry export and data querying

## Features

### omnisignal Provider

The signal provider enables integration with New Relic's alerting system:

- Fetch incidents with time and status filtering
- Acknowledge open incidents
- Close incidents programmatically
- Automatic normalization to signal-spec Signal format

```go
import (
    "github.com/plexusone/omnisignal"
    _ "github.com/plexusone/omni-newrelic/omnisignal"
)

provider, err := omnisignal.New("newrelic", omnisignal.Config{
    APIKey: os.Getenv("NEW_RELIC_API_KEY"),
    Options: map[string]any{
        "account_id": 12345,
        "region":     "US",
    },
})
```

### omniobserve Provider

The observability provider offers two capabilities:

**Telemetry Export:**

- Metrics (counters, gauges, histograms)
- Distributed tracing with span correlation
- Structured logging with trace context

```go
import (
    "github.com/plexusone/omniobserve/observops"
    _ "github.com/plexusone/omni-newrelic/omniobserve"
)

provider, err := observops.Open("newrelic",
    observops.WithAPIKey(os.Getenv("NEW_RELIC_LICENSE_KEY")),
    observops.WithServiceName("my-service"),
)
```

**Data Query:**

- NRQL query execution with metadata extraction
- APM application listing and metrics
- Entity search across all New Relic domains
- Log and trace querying via NRQL

```go
import "github.com/plexusone/omni-newrelic/omniobserve"

qp, err := omniobserve.NewQueryProvider(omniobserve.QueryConfig{
    APIKey:    os.Getenv("NEW_RELIC_API_KEY"),
    AccountID: 12345,
})

result, err := qp.QueryNRQL(ctx, "SELECT count(*) FROM Transaction")
```

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| newrelic-client-go | v2.86.1 | Official New Relic Go SDK |
| omnisignal | v0.1.0 | Signal ingestion abstraction |
| omniobserve | v0.10.0 | Observability abstraction |
| signal-spec | v0.1.0 | Canonical signal data model |

## Installation

```bash
go get github.com/plexusone/omni-newrelic@v0.1.0
```

## Configuration

### API Keys

This package uses two types of New Relic API keys:

| Key Type | Used By | Description |
|----------|---------|-------------|
| Personal API Key (User key) | omnisignal, QueryProvider | For querying data and managing incidents |
| License Key (Ingest) | Telemetry export | For sending metrics, traces, and logs |

### Regions

Both US and EU New Relic regions are supported. Set `region` to `"US"` (default) or `"EU"`.

## Known Limitations

- **Streaming not supported:** The omnisignal provider does not support real-time streaming. New Relic requires webhook configuration for push-based alerting.
- **Rate limits:** New Relic API rate limits apply. See [New Relic API rate limits](https://docs.newrelic.com/docs/apis/rest-api-v2/requirements/api-overload-protection-handling-429-errors/).

## What's Next

- Additional entity type support in QueryProvider
- Metric aggregation helpers
- Dashboard querying support

## Links

- [GitHub Repository](https://github.com/plexusone/omni-newrelic)
- [New Relic Go SDK](https://github.com/newrelic/newrelic-client-go)
- [omnisignal](https://github.com/plexusone/omnisignal)
- [omniobserve](https://github.com/plexusone/omniobserve)
