# omni-newrelic

New Relic integrations for PlexusOne abstraction layers.

## Overview

`omni-newrelic` wraps the official [newrelic-client-go](https://github.com/newrelic/newrelic-client-go) SDK and provides implementations for multiple PlexusOne packages:

| Package | Status | Description |
|---------|--------|-------------|
| `omnisignal` | ✅ Implemented | Alert and incident ingestion |
| `omniobserve` | 🔲 Planned | Observability data access |

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

## Configuration

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `APIKey` | string | Yes | New Relic Personal API Key |
| `account_id` | int | Yes | New Relic Account ID |
| `region` | string | No | "US" (default) or "EU" |

## Capabilities

### omnisignal Provider

- **Fetch**: List incidents with filtering by time and status
- **Acknowledge**: Acknowledge open incidents
- **Close**: Close open incidents
- **Streaming**: Not supported (requires webhook configuration)

## Related Packages

- [omnisignal](https://github.com/plexusone/omnisignal) - Signal ingestion abstraction
- [signal-spec](https://github.com/plexusone/signal-spec) - Canonical signal data model
- [newrelic-client-go](https://github.com/newrelic/newrelic-client-go) - Official New Relic Go SDK

## License

MIT
