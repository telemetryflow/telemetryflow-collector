// Package banner provides ASCII art banner for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
//
// LEGO Building Block - Self-contained within tfo-collector project.
package banner

import (
	"fmt"
	"strings"
)

// Config holds banner configuration
type Config struct {
	ProductName string
	Version     string
	BasedOn     string // e.g., "OTEL Collector 0.142.0"
	Motto       string
	GitCommit   string
	BuildTime   string
	GoVersion   string
	Platform    string
	Vendor      string
	VendorURL   string
	Developer   string
	License     string
	SupportURL  string
	Copyright   string
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		ProductName: "TelemetryFlow Collector",
		Version:     "1.1.0",
		BasedOn:     "OTEL Collector 0.114.0",
		Motto:       "Community Enterprise Observability Platform (CEOP)",
		GitCommit:   "unknown",
		BuildTime:   "unknown",
		GoVersion:   "unknown",
		Platform:    "unknown",
		Vendor:      "TelemetryFlow",
		VendorURL:   "https://telemetryflow.id",
		Developer:   "DevOpsCorner Indonesia",
		License:     "Apache-2.0",
		SupportURL:  "https://docs.telemetryflow.id",
		Copyright:   "Copyright (c) 2024-2026 DevOpsCorner Indonesia",
	}
}

// Generate creates the banner string
func Generate(cfg Config) string {
	basedOn := ""
	if cfg.BasedOn != "" {
		basedOn = fmt.Sprintf(" (Based on %s)", cfg.BasedOn)
	}

	return fmt.Sprintf(`
    ___________    .__                        __
    \__    ___/___ |  |   ____   _____   _____/  |________ ___.__.
      |    |_/ __ \|  | _/ __ \ /     \_/ __ \   __\_  __ <   |  |
      |    |\  ___/|  |_\  ___/|  Y Y  \  ___/|  |  |  | \/\___  |
      |____| \___  >____/\___  >__|_|  /\___  >__|  |__|   / ____|
                 \/          \/      \/     \/             \/
                    ___________.__
                    \_   _____/|  |   ______  _  __
                     |    __)  |  |  /  _ \ \/ \/ /
                     |     \   |  |_(  <_> )     /
                     |___  /   |____/\____/ \/\_/
                         \/
               _________        .__  .__                 __
               \_   ___ \  ____ |  | |  |   ____   _____/  |_  ___________
               /    \  \/ /  _ \|  | |  | _/ __ \_/ ___\   __\/  _ \_  __ \
               \     \___(  <_> )  |_|  |_\  ___/\  \___|  | (  <_> )  | \/
                \______  /\____/|____/____/\___  >\___  >__|  \____/|__|
                       \/                      \/     \/

  %s
    %s v%s%s
    %s
  %s
    Platform     %s
    Go Version   %s
    Commit       %s
    Built        %s
  %s
    Vendor       %s (%s)
    Developer    %s
    License      %s
    Support      %s
  %s
    %s
  %s

`, strings.Repeat("═", 78),
		cfg.ProductName, cfg.Version, basedOn, cfg.Motto,
		strings.Repeat("═", 78),
		cfg.Platform, cfg.GoVersion, cfg.GitCommit, cfg.BuildTime,
		strings.Repeat("─", 78),
		cfg.Vendor, cfg.VendorURL, cfg.Developer, cfg.License, cfg.SupportURL,
		strings.Repeat("─", 78),
		cfg.Copyright,
		strings.Repeat("═", 78))
}

// GenerateCompact creates a compact banner
func GenerateCompact(cfg Config) string {
	basedOn := ""
	if cfg.BasedOn != "" {
		basedOn = fmt.Sprintf(" (Based on %s)", cfg.BasedOn)
	}

	return fmt.Sprintf(`
  %s
    %s v%s%s - %s
  %s
    %s
  %s

`, strings.Repeat("═", 78),
		cfg.ProductName, cfg.Version, basedOn, cfg.Motto,
		strings.Repeat("═", 78),
		cfg.Copyright,
		strings.Repeat("═", 78))
}
