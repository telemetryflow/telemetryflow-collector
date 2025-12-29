// Package version provides build and version information for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
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
package version

import (
	"runtime"
	"strings"
	"testing"
)

// TestConstants tests that all constants are properly defined
func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		wantNot  string
	}{
		{"ProductName", ProductName, ""},
		{"ProductShortName", ProductShortName, ""},
		{"ProductDescription", ProductDescription, ""},
		{"Motto", Motto, ""},
		{"Vendor", Vendor, ""},
		{"VendorURL", VendorURL, ""},
		{"Developer", Developer, ""},
		{"DeveloperURL", DeveloperURL, ""},
		{"Copyright", Copyright, ""},
		{"License", License, ""},
		{"LicenseURL", LicenseURL, ""},
		{"SupportURL", SupportURL, ""},
		{"OTELVersion", OTELVersion, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant == tt.wantNot {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}

// TestProductName tests the product name constant
func TestProductName(t *testing.T) {
	if ProductName != "TelemetryFlow Collector" {
		t.Errorf("Expected ProductName 'TelemetryFlow Collector', got '%s'", ProductName)
	}
}

// TestProductShortName tests the short product name
func TestProductShortName(t *testing.T) {
	if ProductShortName != "tfo-collector" {
		t.Errorf("Expected ProductShortName 'tfo-collector', got '%s'", ProductShortName)
	}
}

// TestLicense tests the license constant
func TestLicense(t *testing.T) {
	if License != "Apache-2.0" {
		t.Errorf("Expected License 'Apache-2.0', got '%s'", License)
	}
}

// TestGet tests the Get function
func TestGet(t *testing.T) {
	info := Get()

	if info.Product != ProductName {
		t.Errorf("Expected Product '%s', got '%s'", ProductName, info.Product)
	}

	if info.Description != ProductDescription {
		t.Errorf("Expected Description '%s', got '%s'", ProductDescription, info.Description)
	}

	if info.Version != Version {
		t.Errorf("Expected Version '%s', got '%s'", Version, info.Version)
	}

	if info.OTELVersion != OTELVersion {
		t.Errorf("Expected OTELVersion '%s', got '%s'", OTELVersion, info.OTELVersion)
	}

	if info.GitCommit != GitCommit {
		t.Errorf("Expected GitCommit '%s', got '%s'", GitCommit, info.GitCommit)
	}

	if info.GitBranch != GitBranch {
		t.Errorf("Expected GitBranch '%s', got '%s'", GitBranch, info.GitBranch)
	}

	if info.BuildTime != BuildTime {
		t.Errorf("Expected BuildTime '%s', got '%s'", BuildTime, info.BuildTime)
	}

	if info.GoVersion != GoVersion {
		t.Errorf("Expected GoVersion '%s', got '%s'", GoVersion, info.GoVersion)
	}

	if info.OS != runtime.GOOS {
		t.Errorf("Expected OS '%s', got '%s'", runtime.GOOS, info.OS)
	}

	if info.Arch != runtime.GOARCH {
		t.Errorf("Expected Arch '%s', got '%s'", runtime.GOARCH, info.Arch)
	}

	if info.Vendor != Vendor {
		t.Errorf("Expected Vendor '%s', got '%s'", Vendor, info.Vendor)
	}

	if info.VendorURL != VendorURL {
		t.Errorf("Expected VendorURL '%s', got '%s'", VendorURL, info.VendorURL)
	}

	if info.Developer != Developer {
		t.Errorf("Expected Developer '%s', got '%s'", Developer, info.Developer)
	}

	if info.License != License {
		t.Errorf("Expected License '%s', got '%s'", License, info.License)
	}

	if info.SupportURL != SupportURL {
		t.Errorf("Expected SupportURL '%s', got '%s'", SupportURL, info.SupportURL)
	}
}

// TestString tests the String function
func TestString(t *testing.T) {
	s := String()

	if s == "" {
		t.Error("Expected non-empty string")
	}

	// Verify it contains key information
	mustContain := []string{
		ProductName,
		Version,
		OTELVersion,
		GitCommit,
		GitBranch,
		BuildTime,
		GoVersion,
		runtime.GOOS,
		runtime.GOARCH,
		Vendor,
		VendorURL,
		Developer,
		License,
		SupportURL,
		Copyright,
	}

	for _, substr := range mustContain {
		if !strings.Contains(s, substr) {
			t.Errorf("Expected String() to contain '%s'", substr)
		}
	}
}

// TestShort tests the Short function
func TestShort(t *testing.T) {
	s := Short()

	if s != Version {
		t.Errorf("Expected Short() to return '%s', got '%s'", Version, s)
	}
}

// TestUserAgent tests the UserAgent function
func TestUserAgent(t *testing.T) {
	ua := UserAgent()

	if ua == "" {
		t.Error("Expected non-empty user agent")
	}

	// Should contain product short name and version
	if !strings.Contains(ua, ProductShortName) {
		t.Errorf("Expected UserAgent to contain '%s'", ProductShortName)
	}

	if !strings.Contains(ua, Version) {
		t.Errorf("Expected UserAgent to contain '%s'", Version)
	}

	if !strings.Contains(ua, runtime.GOOS) {
		t.Errorf("Expected UserAgent to contain '%s'", runtime.GOOS)
	}

	if !strings.Contains(ua, runtime.GOARCH) {
		t.Errorf("Expected UserAgent to contain '%s'", runtime.GOARCH)
	}
}

// TestBanner tests the Banner function
func TestBanner(t *testing.T) {
	b := Banner()

	if b == "" {
		t.Error("Expected non-empty banner")
	}

	// Should contain key product information
	mustContain := []string{
		ProductName,
		Version,
		OTELVersion,
		Motto,
		runtime.GOOS,
		runtime.GOARCH,
		GoVersion,
		GitCommit,
		BuildTime,
		Vendor,
		VendorURL,
		Developer,
		License,
		SupportURL,
		Copyright,
	}

	for _, substr := range mustContain {
		if !strings.Contains(b, substr) {
			t.Errorf("Expected Banner() to contain '%s'", substr)
		}
	}

	// Should contain ASCII art elements
	if !strings.Contains(b, "___") {
		t.Error("Expected Banner() to contain ASCII art")
	}
}

// TestOneLiner tests the OneLiner function
func TestOneLiner(t *testing.T) {
	ol := OneLiner()

	if ol == "" {
		t.Error("Expected non-empty one liner")
	}

	if !strings.Contains(ol, ProductName) {
		t.Errorf("Expected OneLiner to contain '%s'", ProductName)
	}

	if !strings.Contains(ol, Version) {
		t.Errorf("Expected OneLiner to contain '%s'", Version)
	}

	if !strings.Contains(ol, runtime.GOOS) {
		t.Errorf("Expected OneLiner to contain '%s'", runtime.GOOS)
	}

	if !strings.Contains(ol, runtime.GOARCH) {
		t.Errorf("Expected OneLiner to contain '%s'", runtime.GOARCH)
	}

	if !strings.Contains(ol, Motto) {
		t.Errorf("Expected OneLiner to contain '%s'", Motto)
	}
}

// TestGetMotto tests the GetMotto function
func TestGetMotto(t *testing.T) {
	m := GetMotto()

	if m != Motto {
		t.Errorf("Expected GetMotto() to return '%s', got '%s'", Motto, m)
	}
}

// TestGetProductInfo tests the GetProductInfo function
func TestGetProductInfo(t *testing.T) {
	pi := GetProductInfo()

	if pi == "" {
		t.Error("Expected non-empty product info")
	}

	if !strings.Contains(pi, ProductName) {
		t.Errorf("Expected GetProductInfo to contain '%s'", ProductName)
	}

	if !strings.Contains(pi, ProductDescription) {
		t.Errorf("Expected GetProductInfo to contain '%s'", ProductDescription)
	}
}

// TestGetSupportInfo tests the GetSupportInfo function
func TestGetSupportInfo(t *testing.T) {
	si := GetSupportInfo()

	if si == "" {
		t.Error("Expected non-empty support info")
	}

	if !strings.Contains(si, SupportURL) {
		t.Errorf("Expected GetSupportInfo to contain '%s'", SupportURL)
	}
}

// TestFull tests the Full function
func TestFull(t *testing.T) {
	f := Full()

	if f == "" {
		t.Error("Expected non-empty full version")
	}

	if !strings.Contains(f, ProductName) {
		t.Errorf("Expected Full to contain '%s'", ProductName)
	}

	if !strings.Contains(f, Version) {
		t.Errorf("Expected Full to contain '%s'", Version)
	}

	if !strings.Contains(f, "v") {
		t.Error("Expected Full to contain 'v' prefix for version")
	}
}

// TestBuildInfo tests the BuildInfo function
func TestBuildInfo(t *testing.T) {
	bi := BuildInfo()

	if bi == nil {
		t.Fatal("Expected non-nil build info map")
	}

	expectedKeys := []string{
		"version",
		"product_name",
		"otel_version",
		"git_commit",
		"git_branch",
		"build_time",
		"go_version",
		"os",
		"arch",
		"vendor",
		"developer",
		"license",
	}

	for _, key := range expectedKeys {
		if _, ok := bi[key]; !ok {
			t.Errorf("Expected BuildInfo to contain key '%s'", key)
		}
	}

	// Verify values
	if bi["version"] != Version {
		t.Errorf("Expected version '%s', got '%s'", Version, bi["version"])
	}

	if bi["product_name"] != ProductName {
		t.Errorf("Expected product_name '%s', got '%s'", ProductName, bi["product_name"])
	}

	if bi["otel_version"] != OTELVersion {
		t.Errorf("Expected otel_version '%s', got '%s'", OTELVersion, bi["otel_version"])
	}

	if bi["os"] != runtime.GOOS {
		t.Errorf("Expected os '%s', got '%s'", runtime.GOOS, bi["os"])
	}

	if bi["arch"] != runtime.GOARCH {
		t.Errorf("Expected arch '%s', got '%s'", runtime.GOARCH, bi["arch"])
	}
}

// TestInfoStruct tests the Info struct
func TestInfoStruct(t *testing.T) {
	info := Info{
		Product:     "Test Product",
		Description: "Test Description",
		Version:     "1.0.0",
		OTELVersion: "0.142.0",
		GitCommit:   "abc123",
		GitBranch:   "main",
		BuildTime:   "2024-01-01T00:00:00Z",
		GoVersion:   "go1.21",
		OS:          "linux",
		Arch:        "amd64",
		Vendor:      "Test Vendor",
		VendorURL:   "https://test.com",
		Developer:   "Test Dev",
		License:     "MIT",
		SupportURL:  "https://support.test.com",
	}

	if info.Product != "Test Product" {
		t.Errorf("Expected Product 'Test Product', got '%s'", info.Product)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", info.Version)
	}
}

// TestGoVersionRuntime tests that GoVersion matches runtime.Version
func TestGoVersionRuntime(t *testing.T) {
	if GoVersion != runtime.Version() {
		t.Errorf("Expected GoVersion '%s', got '%s'", runtime.Version(), GoVersion)
	}
}

// TestURLsAreValid tests that URLs are properly formatted
func TestURLsAreValid(t *testing.T) {
	urls := []struct {
		name string
		url  string
	}{
		{"VendorURL", VendorURL},
		{"DeveloperURL", DeveloperURL},
		{"LicenseURL", LicenseURL},
		{"SupportURL", SupportURL},
	}

	for _, tt := range urls {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.HasPrefix(tt.url, "https://") && !strings.HasPrefix(tt.url, "http://") {
				t.Errorf("Expected %s to start with http(s)://, got '%s'", tt.name, tt.url)
			}
		})
	}
}

// Benchmark tests
func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get()
	}
}

func BenchmarkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = String()
	}
}

func BenchmarkBanner(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Banner()
	}
}

func BenchmarkBuildInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BuildInfo()
	}
}

func BenchmarkUserAgent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UserAgent()
	}
}
