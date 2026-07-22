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
package tfoauthextension_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/extension/extensiontest"

	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoauthextension"
)

// authGetter is a local interface matching the unexported AuthProvider surface.
// We use it to drive GetAPIKeyID/GetAPIKeySecret on the unexported extension
// type returned by the factory.
type authGetter interface {
	GetAPIKeyID() string
	GetAPIKeySecret() string
}

func newStartedAuthExtension(t *testing.T, cfg *tfoauthextension.Config) component.Component {
	t.Helper()
	factory := tfoauthextension.NewFactory()
	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))
	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, ext.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = ext.Shutdown(context.Background()) })
	return ext
}

func TestExtension_GetAPIKeyID(t *testing.T) {
	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	cfg.APIKeyID = configopaque.String("tfk_unit_key_id_12345")
	cfg.APIKeySecret = configopaque.String("tfs_unit_secret_12345")

	ext := newStartedAuthExtension(t, cfg)

	getter, ok := ext.(authGetter)
	require.True(t, ok, "extension must satisfy authGetter")
	assert.Equal(t, "tfk_unit_key_id_12345", getter.GetAPIKeyID())
	assert.Equal(t, "tfs_unit_secret_12345", getter.GetAPIKeySecret())
}

func TestExtension_GetAPIKeySecret_Empty(t *testing.T) {
	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	// Both empty — passthrough mode
	ext := newStartedAuthExtension(t, cfg)

	getter, ok := ext.(authGetter)
	require.True(t, ok)
	assert.Empty(t, getter.GetAPIKeyID())
	assert.Empty(t, getter.GetAPIKeySecret())
}

func TestExtension_PassthroughStart(t *testing.T) {
	// Empty config — neither ValidateOnStart nor ValidationEndpoint — should start cleanly.
	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	ext := newStartedAuthExtension(t, cfg)
	assert.NotNil(t, ext)
}

func TestMaskAPIKey_ShortKey(t *testing.T) {
	// A short key (<=8 chars) takes the "****" branch inside maskAPIKey.
	// We exercise maskAPIKey indirectly through Start's info log.
	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	cfg.APIKeyID = configopaque.String("tfk_ab") // 6 chars — short
	cfg.APIKeySecret = configopaque.String("tfs_ab")
	ext := newStartedAuthExtension(t, cfg)
	assert.NotNil(t, ext)
}

func TestValidateCredentials_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "tfk_test", r.Header.Get("X-TelemetryFlow-Key-ID"))
		assert.Equal(t, "tfs_test", r.Header.Get("X-TelemetryFlow-Key-Secret"))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	cfg.APIKeyID = configopaque.String("tfk_test")
	cfg.APIKeySecret = configopaque.String("tfs_test")
	cfg.ValidateOnStart = true
	cfg.ValidationEndpoint = srv.URL

	ext := newStartedAuthExtension(t, cfg)
	assert.NotNil(t, ext)
}

func TestValidateCredentials_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	cfg.APIKeyID = configopaque.String("tfk_test")
	cfg.APIKeySecret = configopaque.String("tfs_test")
	cfg.ValidateOnStart = true
	cfg.ValidationEndpoint = srv.URL

	factory := tfoauthextension.NewFactory()
	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))
	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)
	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API credentials")
}

func TestValidateCredentials_Forbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	cfg.APIKeyID = configopaque.String("tfk_test")
	cfg.APIKeySecret = configopaque.String("tfs_test")
	cfg.ValidateOnStart = true
	cfg.ValidationEndpoint = srv.URL

	factory := tfoauthextension.NewFactory()
	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))
	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)
	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API credentials")
}

func TestValidateCredentials_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	cfg.APIKeyID = configopaque.String("tfk_test")
	cfg.APIKeySecret = configopaque.String("tfs_test")
	cfg.ValidateOnStart = true
	cfg.ValidationEndpoint = srv.URL

	factory := tfoauthextension.NewFactory()
	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))
	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)
	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected validation response: 500")
}

func TestValidateCredentials_NetworkError(t *testing.T) {
	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	cfg.APIKeyID = configopaque.String("tfk_test")
	cfg.APIKeySecret = configopaque.String("tfs_test")
	cfg.ValidateOnStart = true
	// Use an invalid endpoint so the request fails before any server logic runs.
	cfg.ValidationEndpoint = "http://127.0.0.1:0/validate"

	factory := tfoauthextension.NewFactory()
	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))
	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)
	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation request failed")
}

func TestValidateCredentials_BadEndpointURL(t *testing.T) {
	cfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	cfg.APIKeyID = configopaque.String("tfk_test")
	cfg.APIKeySecret = configopaque.String("tfs_test")
	cfg.ValidateOnStart = true
	// An invalid URL produces a request-construction error before any I/O.
	cfg.ValidationEndpoint = "http://192.168.0.%31/"

	factory := tfoauthextension.NewFactory()
	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))
	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)
	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.Error(t, err)
	// Either "failed to create validation request" or "validation request failed".
	assert.Contains(t, err.Error(), "validation")
}
