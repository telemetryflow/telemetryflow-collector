// TelemetryFlow Collector - AI-Powered Observability & Incident Response Management (IRM) Platform
// Copyright (c) 2024-2026 Telemetri Data Indonesia. All rights reserved.
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
package tfoexporter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/exporter/exportertest"

	"github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"
)

// TestExporter_Start_ToClientError exercises the ToClient error branch in
// start() by configuring an invalid TLS CA bundle. ToClient fails when it
// cannot construct a working *http.Client from the supplied TLS settings.
func TestExporter_Start_ToClientError(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = "https://example.invalid"
	// Point the TLS CA bundle at a path that does not exist so that
	// ToClient returns a "reading CA cert" error.
	cfg.TLS.CAFile = "/nonexistent/path/ca.crt"
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeyID:     configopaque.String("tfk_test"),
		APIKeySecret: configopaque.String("tfs_test"),
	}
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)

	startErr := tracesExp.Start(context.Background(), newExtHost(nil))
	require.Error(t, startErr)
	assert.Contains(t, startErr.Error(), "failed to create HTTP client")
}

// TestExporter_ConfigAuth_OnlySecret exercises the half-configured direct
// auth case where only APIKeySecret is set (no APIKeyID). The Validate path
// returns an error for this case, but we want to make sure that's enforced.
func TestExporter_ConfigAuth_OnlySecret(t *testing.T) {
	cfg := factoryEmptyConfig()
	cfg.Endpoint = "http://localhost:4318"
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeySecret: configopaque.String("tfs_only"),
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth requires either")
}

// factoryEmptyConfig returns a fresh Config via the factory so we don't have
// to construct one through the squashed confighttp.ClientConfig literal
// (which Go rejects because of the embedded configtls.ClientConfig's
// unexported anti-literal field).
func factoryEmptyConfig() *tfoexporter.Config {
	return tfoexporter.NewFactory().CreateDefaultConfig().(*tfoexporter.Config)
}
