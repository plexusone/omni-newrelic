// Package omnisignal provides a New Relic signal provider for omnisignal.
//
// This is a thick provider that uses the official newrelic-client-go SDK.
//
// Usage:
//
//	import (
//	    "github.com/plexusone/omnisignal"
//	    _ "github.com/plexusone/omni-newrelic/omnisignal"
//	)
//
//	provider, err := omnisignal.New("newrelic", omnisignal.Config{
//	    APIKey: os.Getenv("NEW_RELIC_API_KEY"),
//	    Options: map[string]any{
//	        "account_id": 12345,
//	        "region":     "US",
//	    },
//	})
package omnisignal

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/newrelic/newrelic-client-go/v2/pkg/alerts"
	omninewrelic "github.com/plexusone/omni-newrelic"
	"github.com/plexusone/omnisignal"
	"github.com/plexusone/signal-spec/pkg/common"
	"github.com/plexusone/signal-spec/pkg/signal"
)

const (
	// ProviderName is the identifier for this provider.
	ProviderName = "newrelic"
)

func init() {
	omnisignal.Register(ProviderName, NewProvider, omnisignal.PriorityThick)
}

// Provider implements omnisignal.Provider for New Relic.
type Provider struct {
	client *omninewrelic.Client
	config omnisignal.Config
}

// NewProvider creates a new New Relic signal provider.
func NewProvider(cfg omnisignal.Config) (omnisignal.Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("%w: APIKey is required", omnisignal.ErrInvalidConfig)
	}

	// Get account ID from options
	accountID := 0
	if id, ok := cfg.Options["account_id"].(int); ok {
		accountID = id
	} else if id, ok := cfg.Options["account_id"].(float64); ok {
		accountID = int(id)
	}

	if accountID == 0 {
		return nil, fmt.Errorf("%w: account_id is required in Options", omnisignal.ErrInvalidConfig)
	}

	// Get region
	region := "US"
	if r, ok := cfg.Options["region"].(string); ok {
		region = r
	}

	// Create New Relic client
	nrClient, err := omninewrelic.NewClient(omninewrelic.Config{
		APIKey:    cfg.APIKey,
		AccountID: accountID,
		Region:    region,
	})
	if err != nil {
		return nil, fmt.Errorf("creating New Relic client: %w", err)
	}

	return &Provider{
		client: nrClient,
		config: cfg,
	}, nil
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return ProviderName
}

// Fetch retrieves alerts/incidents from New Relic.
func (p *Provider) Fetch(ctx context.Context, opts omnisignal.FetchOptions) ([]signal.Signal, error) {
	var signals []signal.Signal

	// Determine if we should only fetch open incidents
	onlyOpen := false
	if len(opts.Statuses) > 0 {
		for _, s := range opts.Statuses {
			if s == "open" || s == "triggered" {
				onlyOpen = true
				break
			}
		}
	}

	// Fetch incidents
	incidents, err := p.client.Alerts.ListIncidentsWithContext(ctx, onlyOpen, false)
	if err != nil {
		return nil, fmt.Errorf("fetching incidents: %w", err)
	}

	for _, incident := range incidents {
		// Filter by time if specified
		if incident.OpenedAt != nil {
			openedAt := time.Time(*incident.OpenedAt)
			if !opts.Since.IsZero() && openedAt.Before(opts.Since) {
				continue
			}
			if !opts.Until.IsZero() && openedAt.After(opts.Until) {
				continue
			}
		}

		sig := p.normalizeIncident(incident)
		signals = append(signals, sig)

		if opts.Limit > 0 && len(signals) >= opts.Limit {
			break
		}
	}

	return signals, nil
}

// Subscribe is not supported for New Relic.
// New Relic supports webhooks via workflow destinations.
func (p *Provider) Subscribe(ctx context.Context, opts omnisignal.SubscribeOptions) (<-chan signal.Signal, error) {
	return nil, omnisignal.ErrNotSupported
}

// Capabilities returns what this provider supports.
func (p *Provider) Capabilities() omnisignal.Capabilities {
	return omnisignal.Capabilities{
		SupportsStreaming:   false,
		SupportsBatchFetch:  true,
		SupportsFiltering:   true,
		SupportsAcknowledge: true,
		MaxBatchSize:        0, // No documented limit
		RateLimitPerMinute:  0, // Varies
		SignalTypes: []signal.Type{
			signal.TypeAlert,
			signal.TypeOutage,
		},
	}
}

// Close releases resources.
func (p *Provider) Close() error {
	return nil
}

// normalizeIncident converts a New Relic incident to a signal-spec Signal.
func (p *Provider) normalizeIncident(incident *alerts.Incident) signal.Signal {
	// Parse timestamps
	var observedAt time.Time
	if incident.OpenedAt != nil {
		observedAt = time.Time(*incident.OpenedAt)
	} else {
		observedAt = time.Now()
	}

	// Determine if closed
	status := signal.StatusNew
	if incident.ClosedAt != nil {
		status = signal.StatusArchived
	}

	// Build domain from policy
	domain := common.Domain{
		Name:      "monitoring",
		Subdomain: "alerts",
	}

	// Severity - New Relic incidents don't have direct severity
	// We'd need to look up the policy/condition to determine this
	severity := common.SeverityHigh // Default to high for incidents

	return signal.Signal{
		ID:     fmt.Sprintf("nr-%d", incident.ID),
		Type:   signal.TypeAlert,
		Status: status,
		Source: common.SourceSystem{
			Type:       "monitoring",
			Name:       "newrelic",
			ExternalID: strconv.Itoa(incident.ID),
		},
		Domain:     domain,
		Severity:   severity,
		Summary:    fmt.Sprintf("New Relic Incident #%d", incident.ID),
		ObservedAt: observedAt,
		ReceivedAt: time.Now(),
		Metadata: map[string]any{
			"newrelic_incident_id":         incident.ID,
			"newrelic_policy_id":           incident.Links.PolicyID,
			"newrelic_incident_preference": incident.IncidentPreference,
			"newrelic_violation_ids":       incident.Links.Violations,
		},
	}
}

// AcknowledgeIncident acknowledges a New Relic incident.
func (p *Provider) AcknowledgeIncident(ctx context.Context, incidentID int) error {
	_, err := p.client.Alerts.AcknowledgeIncidentWithContext(ctx, incidentID)
	return err
}

// CloseIncident closes a New Relic incident.
func (p *Provider) CloseIncident(ctx context.Context, incidentID int) error {
	_, err := p.client.Alerts.CloseIncidentWithContext(ctx, incidentID)
	return err
}
