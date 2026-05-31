// Package omniobserve provides New Relic observability integration for omniobserve.
//
// This package provides two capabilities:
//
// 1. Telemetry Export: Re-exports the OTLP-based telemetry provider from
// omniobserve/observops/newrelic for sending metrics, traces, and logs.
//
// 2. Data Query: Provides a QueryProvider for reading observability data
// from New Relic via NRQL queries, APM metrics, and entity search.
//
// # Telemetry Export Usage
//
//	import (
//	    "github.com/plexusone/omniobserve/observops"
//	    _ "github.com/plexusone/omni-newrelic/omniobserve" // Register New Relic provider
//	)
//
//	provider, err := observops.Open("newrelic",
//	    observops.WithAPIKey(os.Getenv("NEW_RELIC_LICENSE_KEY")),
//	    observops.WithServiceName("my-service"),
//	)
//
// # Data Query Usage
//
//	import "github.com/plexusone/omni-newrelic/omniobserve"
//
//	qp, err := omniobserve.NewQueryProvider(omniobserve.QueryConfig{
//	    APIKey:    os.Getenv("NEW_RELIC_API_KEY"),
//	    AccountID: 12345,
//	    Region:    "US",
//	})
//
//	results, err := qp.QueryNRQL(ctx, "SELECT count(*) FROM Transaction")
package omniobserve

import (
	"context"
	"fmt"
	"time"

	omninewrelic "github.com/plexusone/omni-newrelic"

	// Register the OTLP-based telemetry provider from omniobserve.
	// This blank import triggers the init() function which calls observops.Register().
	_ "github.com/plexusone/omniobserve/observops/newrelic"

	"github.com/newrelic/newrelic-client-go/v2/pkg/apm"
	"github.com/newrelic/newrelic-client-go/v2/pkg/common"
	"github.com/newrelic/newrelic-client-go/v2/pkg/entities"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nrdb"
)

// QueryConfig holds configuration for the QueryProvider.
type QueryConfig struct {
	// APIKey is the New Relic Personal API Key (User key).
	// This is different from the License key used for telemetry export.
	APIKey string

	// AccountID is the New Relic account ID.
	AccountID int

	// Region is the New Relic region ("US" or "EU").
	// Defaults to "US".
	Region string
}

// QueryProvider provides methods for querying observability data from New Relic.
type QueryProvider struct {
	client *omninewrelic.Client
	config QueryConfig
}

// NewQueryProvider creates a new QueryProvider for reading data from New Relic.
func NewQueryProvider(cfg QueryConfig) (*QueryProvider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("omniobserve: APIKey is required")
	}
	if cfg.AccountID == 0 {
		return nil, fmt.Errorf("omniobserve: AccountID is required")
	}

	region := cfg.Region
	if region == "" {
		region = "US"
	}

	client, err := omninewrelic.NewClient(omninewrelic.Config{
		APIKey:    cfg.APIKey,
		AccountID: cfg.AccountID,
		Region:    region,
	})
	if err != nil {
		return nil, fmt.Errorf("omniobserve: creating client: %w", err)
	}

	return &QueryProvider{
		client: client,
		config: cfg,
	}, nil
}

// NRQLResult represents the result of a NRQL query.
type NRQLResult struct {
	// Results contains the query results as a slice of maps.
	Results []map[string]any

	// Metadata contains query metadata including event types and time window.
	Metadata *NRQLMetadata
}

// NRQLMetadata contains metadata about a NRQL query result.
type NRQLMetadata struct {
	// EventTypes lists the event types queried.
	EventTypes []string

	// BeginTime is the start of the query time window.
	BeginTime time.Time

	// EndTime is the end of the query time window.
	EndTime time.Time

	// Messages contains any informational messages.
	Messages []string
}

// QueryNRQL executes a NRQL query and returns the results.
func (p *QueryProvider) QueryNRQL(ctx context.Context, query string) (*NRQLResult, error) {
	resp, err := p.client.Nrdb.QueryWithContext(ctx, p.config.AccountID, nrdb.NRQL(query))
	if err != nil {
		return nil, fmt.Errorf("executing NRQL query: %w", err)
	}

	result := &NRQLResult{
		Results: make([]map[string]any, 0, len(resp.Results)),
	}

	// Convert results - NRDBResult is map[string]interface{}
	for _, r := range resp.Results {
		// NRDBResult is already a map[string]interface{}, just convert to map[string]any
		result.Results = append(result.Results, map[string]any(r))
	}

	// Extract metadata
	result.Metadata = &NRQLMetadata{}
	if resp.Metadata.EventTypes != nil {
		result.Metadata.EventTypes = resp.Metadata.EventTypes
	}
	if resp.Metadata.Messages != nil {
		result.Metadata.Messages = resp.Metadata.Messages
	}
	// Convert time window - nrtime.EpochMilliseconds can be cast directly to time.Time
	result.Metadata.BeginTime = time.Time(resp.Metadata.TimeWindow.Begin)
	result.Metadata.EndTime = time.Time(resp.Metadata.TimeWindow.End)

	return result, nil
}

// APMApplication represents a New Relic APM application.
type APMApplication struct {
	ID             int
	Name           string
	Language       string
	HealthStatus   string
	Reporting      bool
	LastReportedAt string

	// Summary metrics
	ResponseTime  float64 // in seconds
	Throughput    float64 // requests per minute
	ErrorRate     float64 // percentage
	ApdexTarget   float64
	ApdexScore    float64
	HostCount     int
	InstanceCount int
}

// ListAPMApplications returns all APM applications for the account.
func (p *QueryProvider) ListAPMApplications(ctx context.Context) ([]APMApplication, error) {
	apps, err := p.client.APM.ListApplicationsWithContext(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("listing APM applications: %w", err)
	}

	result := make([]APMApplication, 0, len(apps))
	for _, app := range apps {
		a := APMApplication{
			ID:             app.ID,
			Name:           app.Name,
			Language:       app.Language,
			HealthStatus:   app.HealthStatus,
			Reporting:      app.Reporting,
			LastReportedAt: app.LastReportedAt,
			ResponseTime:   app.Summary.ResponseTime,
			Throughput:     app.Summary.Throughput,
			ErrorRate:      app.Summary.ErrorRate,
			ApdexTarget:    app.Summary.ApdexTarget,
			ApdexScore:     app.Summary.ApdexScore,
			HostCount:      app.Summary.HostCount,
			InstanceCount:  app.Summary.InstanceCount,
		}
		result = append(result, a)
	}

	return result, nil
}

// GetAPMApplication returns a specific APM application by ID.
func (p *QueryProvider) GetAPMApplication(ctx context.Context, appID int) (*APMApplication, error) {
	app, err := p.client.APM.GetApplicationWithContext(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("getting APM application %d: %w", appID, err)
	}

	return &APMApplication{
		ID:             app.ID,
		Name:           app.Name,
		Language:       app.Language,
		HealthStatus:   app.HealthStatus,
		Reporting:      app.Reporting,
		LastReportedAt: app.LastReportedAt,
		ResponseTime:   app.Summary.ResponseTime,
		Throughput:     app.Summary.Throughput,
		ErrorRate:      app.Summary.ErrorRate,
		ApdexTarget:    app.Summary.ApdexTarget,
		ApdexScore:     app.Summary.ApdexScore,
		HostCount:      app.Summary.HostCount,
		InstanceCount:  app.Summary.InstanceCount,
	}, nil
}

// MetricDataOptions configures metric data retrieval.
type MetricDataOptions struct {
	// Names specifies which metrics to retrieve.
	Names []string

	// From is the start time for metrics.
	From time.Time

	// To is the end time for metrics.
	To time.Time

	// Period is the metric time slice duration in seconds.
	// Defaults to 60 (1 minute).
	Period int

	// Summarize returns a single summarized value instead of time series.
	Summarize bool
}

// MetricData represents time series metric data.
type MetricData struct {
	Name       string
	Timeslices []MetricTimeslice
}

// MetricTimeslice represents a single time slice of metric data.
type MetricTimeslice struct {
	From   time.Time
	To     time.Time
	Values map[string]float64
}

// GetAPMMetrics retrieves metric data for an APM application.
func (p *QueryProvider) GetAPMMetrics(ctx context.Context, appID int, opts MetricDataOptions) ([]MetricData, error) {
	params := apm.MetricDataParams{
		Names: opts.Names,
	}

	if !opts.From.IsZero() {
		params.From = &opts.From
	}
	if !opts.To.IsZero() {
		params.To = &opts.To
	}
	if opts.Period > 0 {
		params.Period = opts.Period
	}
	if opts.Summarize {
		params.Summarize = true
	}

	metrics, err := p.client.APM.GetMetricDataWithContext(ctx, appID, params)
	if err != nil {
		return nil, fmt.Errorf("getting APM metrics for app %d: %w", appID, err)
	}

	result := make([]MetricData, 0, len(metrics))
	for _, m := range metrics {
		md := MetricData{
			Name:       m.Name,
			Timeslices: make([]MetricTimeslice, 0, len(m.Timeslices)),
		}

		for _, ts := range m.Timeslices {
			slice := MetricTimeslice{
				Values: make(map[string]float64),
			}
			if ts.From != nil {
				slice.From = *ts.From
			}
			if ts.To != nil {
				slice.To = *ts.To
			}
			// Convert MetricTimesliceValues to map using known fields
			slice.Values["as_percentage"] = ts.Values.AsPercentage
			slice.Values["average_time"] = ts.Values.AverageTime
			slice.Values["calls_per_minute"] = ts.Values.CallsPerMinute
			slice.Values["max_value"] = ts.Values.MaxValue

			md.Timeslices = append(md.Timeslices, slice)
		}

		result = append(result, md)
	}

	return result, nil
}

// Entity represents a New Relic entity (APM app, host, browser app, etc.).
type Entity struct {
	GUID      string
	Name      string
	Type      string
	Domain    string
	AccountID int
	Tags      map[string][]string
}

// EntitySearchOptions configures entity search.
type EntitySearchOptions struct {
	// Query is a search string to match entity names.
	Query string

	// Domain filters by entity domain (APM, BROWSER, INFRA, etc.).
	Domain string

	// Type filters by entity type (APPLICATION, HOST, etc.).
	Type string

	// Tags filters by entity tags.
	Tags map[string]string
}

// SearchEntities searches for entities matching the given options.
func (p *QueryProvider) SearchEntities(ctx context.Context, opts EntitySearchOptions) ([]Entity, error) {
	query := opts.Query

	searchOpts := entities.EntitySearchOptions{
		CaseSensitiveTagMatching: false,
	}

	var queryBuilder entities.EntitySearchQueryBuilder
	if opts.Domain != "" {
		queryBuilder.Domain = entities.EntitySearchQueryBuilderDomain(opts.Domain)
	}
	if opts.Type != "" {
		queryBuilder.Type = entities.EntitySearchQueryBuilderType(opts.Type)
	}
	if len(opts.Tags) > 0 {
		tags := make([]entities.EntitySearchQueryBuilderTag, 0, len(opts.Tags))
		for k, v := range opts.Tags {
			tags = append(tags, entities.EntitySearchQueryBuilderTag{
				Key:   k,
				Value: v,
			})
		}
		queryBuilder.Tags = tags
	}

	resp, err := p.client.Entities.GetEntitySearchWithContext(
		ctx,
		searchOpts,
		query,
		queryBuilder,
		[]entities.EntitySearchSortCriteria{},
		[]entities.SortCriterionWithDirection{},
	)
	if err != nil {
		return nil, fmt.Errorf("searching entities: %w", err)
	}

	result := make([]Entity, 0, len(resp.Results.Entities))
	for _, e := range resp.Results.Entities {
		entity := Entity{
			GUID:      string(e.GetGUID()),
			Name:      e.GetName(),
			Type:      e.GetType(),
			Domain:    e.GetDomain(),
			AccountID: e.GetAccountID(),
			Tags:      make(map[string][]string),
		}

		for _, tag := range e.GetTags() {
			entity.Tags[tag.Key] = tag.Values
		}

		result = append(result, entity)
	}

	return result, nil
}

// GetEntity retrieves a specific entity by GUID.
func (p *QueryProvider) GetEntity(ctx context.Context, guid string) (*Entity, error) {
	resp, err := p.client.Entities.GetEntityWithContext(ctx, common.EntityGUID(guid))
	if err != nil {
		return nil, fmt.Errorf("getting entity %s: %w", guid, err)
	}

	entity := &Entity{
		GUID:      string((*resp).GetGUID()),
		Name:      (*resp).GetName(),
		Type:      (*resp).GetType(),
		Domain:    (*resp).GetDomain(),
		AccountID: (*resp).GetAccountID(),
		Tags:      make(map[string][]string),
	}

	for _, tag := range (*resp).GetTags() {
		entity.Tags[tag.Key] = tag.Values
	}

	return entity, nil
}

// QueryLogs queries logs using NRQL. This is a convenience method that
// wraps QueryNRQL with a Log event type query.
func (p *QueryProvider) QueryLogs(ctx context.Context, where string, limit int, since time.Duration) (*NRQLResult, error) {
	query := "SELECT * FROM Log"
	if where != "" {
		query += " WHERE " + where
	}
	if since > 0 {
		query += fmt.Sprintf(" SINCE %d minutes ago", int(since.Minutes()))
	}
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	return p.QueryNRQL(ctx, query)
}

// QueryTraces queries distributed traces using NRQL.
func (p *QueryProvider) QueryTraces(ctx context.Context, where string, limit int, since time.Duration) (*NRQLResult, error) {
	query := "SELECT * FROM Span"
	if where != "" {
		query += " WHERE " + where
	}
	if since > 0 {
		query += fmt.Sprintf(" SINCE %d minutes ago", int(since.Minutes()))
	}
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	return p.QueryNRQL(ctx, query)
}

// Close releases any resources held by the QueryProvider.
func (p *QueryProvider) Close() error {
	// The underlying newrelic client doesn't require explicit cleanup
	return nil
}
