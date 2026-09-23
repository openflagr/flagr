package entity

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// Enricher namespaces. The namespace is both the enricher's identity on a flag
// and the selector for its implementation in the handler registry. Namespaces
// are hardcoded in code: adding an integration is a code change.
const (
	// EnricherNamespaceTs contributes @ts, @ts_hour, @ts_weekday, @ts_month.
	// Global: supplied by the server, never stored on a flag.
	EnricherNamespaceTs = "ts"
	// EnricherNamespaceHTTP contributes @http_<header>. Global.
	EnricherNamespaceHTTP = "http"
	// EnricherNamespaceJev contributes @jev_<question> from a System One call.
	// Flag-scoped.
	EnricherNamespaceJev = "jev"
)

// Enricher is a flag-scoped context enricher definition. ConfigJSON holds the
// namespace-specific configuration, decoded and validated by the namespace
// implementation (handler registry), not here.
type Enricher struct {
	gorm.Model

	FlagID uint `gorm:"index:idx_enricher_flagid"`
	// Namespace is the enricher's identity; at most one enricher per namespace
	// per flag.
	Namespace string `gorm:"type:varchar(64);index:idx_enricher_flagid_namespace"`
	// ConfigJSON is the namespace-specific config.
	ConfigJSON string `gorm:"type:text"`
}

// Validate checks storage invariants only: a namespace and well-formed JSON
// config. Namespace scope and config schema are validated by the handler
// registry, which owns the namespace implementations.
func (e *Enricher) Validate() error {
	if e == nil {
		return fmt.Errorf("enricher is required")
	}
	if strings.TrimSpace(e.Namespace) == "" {
		return fmt.Errorf("enricher namespace is required")
	}
	if trimmed := strings.TrimSpace(e.ConfigJSON); trimmed != "" && !json.Valid([]byte(trimmed)) {
		return fmt.Errorf("enricher config is not valid JSON")
	}
	return nil
}

// SetConfig marshals cfg into ConfigJSON for the given namespace.
func (e *Enricher) SetConfig(namespace string, cfg any) error {
	if strings.TrimSpace(namespace) == "" {
		return fmt.Errorf("enricher namespace is required")
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding enricher config: %w", err)
	}
	e.Namespace = namespace
	e.ConfigJSON = string(b)
	return nil
}

// DecodeConfig unmarshals ConfigJSON into dest.
func (e *Enricher) DecodeConfig(dest any) error {
	if strings.TrimSpace(e.ConfigJSON) == "" {
		return fmt.Errorf("enricher config is required")
	}
	if err := json.Unmarshal([]byte(e.ConfigJSON), dest); err != nil {
		return fmt.Errorf("invalid enricher config: %w", err)
	}
	return nil
}
