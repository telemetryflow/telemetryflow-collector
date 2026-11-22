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
//
// healthcheck is a tiny, statically linked (CGO_ENABLED=0) probe used by the
// container HEALTHCHECK. The distroless/static runtime image ships no shell and
// no curl, so this binary replaces the previous `curl -f http://localhost:13133/`
// healthcheck. It performs a single GET against the collector's health_check
// extension and exits 0 when healthy, 1 otherwise.
//
// The target endpoint can be overridden with TFO_HEALTHCHECK_URL (defaults to
// http://localhost:13133/).

package main

import (
	"net/http"
	"os"
	"time"
)

const (
	defaultURL     = "http://localhost:13133/"
	requestTimeout = 5 * time.Second
)

func main() {
	url := os.Getenv("TFO_HEALTHCHECK_URL")
	if url == "" {
		url = defaultURL
	}

	client := &http.Client{Timeout: requestTimeout}

	resp, err := client.Get(url)
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Healthy responses are any 2xx/3xx status; the health_check extension
	// returns 200 when serving and 503 when the pipeline is not ready.
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		os.Exit(1)
	}

	os.Exit(0)
}
