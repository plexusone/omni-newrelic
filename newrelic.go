// Package omninewrelic provides New Relic integrations for PlexusOne packages.
//
// This package wraps the official newrelic-client-go SDK and provides
// implementations for multiple PlexusOne abstraction layers:
//
//   - omnisignal: Alert and incident ingestion
//   - omniobserve: Observability data access (future)
//
// Usage:
//
//	import (
//	    "github.com/plexusone/omnisignal"
//	    _ "github.com/plexusone/omni-newrelic/omnisignal" // Register New Relic signal provider
//	)
//
//	provider, err := omnisignal.New("newrelic", omnisignal.Config{
//	    APIKey: os.Getenv("NEW_RELIC_API_KEY"),
//	    Options: map[string]any{
//	        "account_id": 12345,
//	        "region":     "US", // or "EU"
//	    },
//	})
package omninewrelic

import (
	"github.com/newrelic/newrelic-client-go/v2/newrelic"
)

// Client wraps the New Relic client for use across PlexusOne packages.
type Client struct {
	*newrelic.NewRelic
	accountID int
}

// Config holds New Relic configuration.
type Config struct {
	// APIKey is the New Relic Personal API Key.
	APIKey string

	// AccountID is the New Relic account ID.
	AccountID int

	// Region is the New Relic region ("US" or "EU").
	// Defaults to "US".
	Region string
}

// NewClient creates a new New Relic client wrapper.
func NewClient(cfg Config) (*Client, error) {
	// Determine region string for ConfigRegion
	regStr := "US"
	if cfg.Region == "EU" {
		regStr = "EU"
	}

	// Create New Relic client using functional options
	nr, err := newrelic.New(
		newrelic.ConfigPersonalAPIKey(cfg.APIKey),
		newrelic.ConfigRegion(regStr),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		NewRelic:  nr,
		accountID: cfg.AccountID,
	}, nil
}

// AccountID returns the configured account ID.
func (c *Client) AccountID() int {
	return c.accountID
}
