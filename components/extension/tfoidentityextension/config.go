// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
