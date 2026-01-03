// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoidentityextension

// Config defines the configuration for the TFO identity extension.
type Config struct {
	// ID is the unique collector identifier.
	// If empty, a UUID will be auto-generated.
	ID string `mapstructure:"id"`

	// Hostname is the collector hostname.
	// If empty, it will be auto-detected.
	Hostname string `mapstructure:"hostname"`

	// Name is a human-readable collector name.
	Name string `mapstructure:"name"`

	// Description is a human-readable collector description.
	Description string `mapstructure:"description"`

	// Tags are custom key-value pairs for labeling and filtering.
	Tags map[string]string `mapstructure:"tags"`

	// EnrichResources enables adding collector identity to all telemetry resources.
	// Default: true
	EnrichResources bool `mapstructure:"enrich_resources"`
}

// Validate checks the configuration for errors.
func (cfg *Config) Validate() error {
	// No required fields - all can be auto-generated or optional
	return nil
}
