package handler

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/pkg/mapper/entity_restapi/e2r"
	"github.com/openflagr/flagr/swagger_gen/models"
	"github.com/sirupsen/logrus"
)

// enricherScope says where an enricher definition comes from: built into the
// server (global, never stored on a flag) or declared on the flag.
type enricherScope string

const (
	scopeGlobal enricherScope = "global"
	scopeFlag   enricherScope = "flag"
)

// enrichInput is what one enricher sees when it runs. The context is the
// accumulated evaluation context with the enricher's own property prefix
// already masked out, so an enricher never reads its own prior output.
type enrichInput struct {
	ctx        context.Context
	context    map[string]any
	entityID   string
	entityType string
	request    *http.Request
}

// enricher is a resolved enricher for one evaluation. Everything except the
// config is a property of its namespace, set by the namespace builder.
type enricher struct {
	namespace  string
	scope      enricherScope
	prefix     string
	properties []string
	// config is the decoded namespace-specific config, exposed by the read model.
	config any
	// enabled reflects the namespace's env switch. Disabled enrichers are not
	// listed and contribute nothing.
	enabled bool
	run     func(enrichInput) (map[string]any, error)
}

// enricherBuilders is the code registry of enricher namespaces. Adding an
// integration is a code change; the namespace is the identity. Global builders
// ignore the config argument.
var enricherBuilders = map[string]func(configJSON string) (*enricher, error){
	entity.EnricherNamespaceTs:   buildTsEnricher,
	entity.EnricherNamespaceHTTP: buildHTTPEnricher,
	entity.EnricherNamespaceJev:  buildJevEnricher,
}

// globalNamespaceOrder is the execution and display order of global enrichers.
var globalNamespaceOrder = []string{
	entity.EnricherNamespaceTs,
	entity.EnricherNamespaceHTTP,
}

// globalEnrichers returns the built-in enrichers in a fixed order.
func globalEnrichers() []enricher {
	out := make([]enricher, 0, len(globalNamespaceOrder))
	for _, ns := range globalNamespaceOrder {
		e, err := enricherBuilders[ns]("")
		if err != nil {
			logrus.WithError(err).WithField("namespace", ns).Warn("skipping global enricher")
			continue
		}
		out = append(out, *e)
	}
	return out
}

// flagEnrichers resolves a flag's flag-scoped enrichers in declaration order. A
// stored config that no longer parses is skipped with a warning; referencing
// constraints then fail closed.
func flagEnrichers(flag *entity.Flag) []enricher {
	if flag == nil {
		return nil
	}
	src := flag.Enrichers
	out := make([]enricher, 0, len(src))
	for i := range src {
		fe := src[i]
		build, ok := enricherBuilders[fe.Namespace]
		if !ok {
			logrus.WithField("namespace", fe.Namespace).
				Warn("skipping unknown flag enricher namespace")
			continue
		}
		e, err := build(fe.ConfigJSON)
		if err != nil {
			logrus.WithError(err).WithField("namespace", fe.Namespace).
				Warn("skipping invalid flag enricher config")
			continue
		}
		out = append(out, *e)
	}
	return out
}

// effectiveEnrichers returns globals followed by the flag's own enrichers: the
// full catalog, including namespaces disabled by env (used by validation).
func effectiveEnrichers(flag *entity.Flag) []enricher {
	return append(globalEnrichers(), flagEnrichers(flag)...)
}

// enabledEnrichers filters to enrichers whose env switch is on. This is the
// set the eval pipeline runs; a disabled enricher contributes nothing.
func enabledEnrichers(entries []enricher) []enricher {
	out := make([]enricher, 0, len(entries))
	for i := range entries {
		if entries[i].enabled {
			out = append(out, entries[i])
		}
	}
	return out
}

// visibleEnrichers is the catalog the API lists: enabled enrichers plus
// flag-scoped declarations even when their namespace is disabled server-side.
// A flag may declare a Jev enricher before the deployment has a System One
// endpoint; hiding it would make the declaration unmanageable. Disabled
// entries carry enabled=false and are still skipped by the eval pipeline.
func visibleEnrichers(entries []enricher) []enricher {
	out := make([]enricher, 0, len(entries))
	for i := range entries {
		if entries[i].enabled || entries[i].scope == scopeFlag {
			out = append(out, entries[i])
		}
	}
	return out
}

// unknownEnrichedProperties returns the distinct `@`-prefixed constraint
// properties no effective enricher provides, sorted. Warn-only: an unknown
// reference is accepted and fails closed at evaluation.
func unknownEnrichedProperties(flag *entity.Flag) []string {
	entries := effectiveEnrichers(flag)
	seen := map[string]bool{}
	out := make([]string, 0)
	for i := range flag.Segments {
		for j := range flag.Segments[i].Constraints {
			p := flag.Segments[i].Constraints[j].Property
			if !strings.HasPrefix(p, "@") || seen[p] {
				continue
			}
			seen[p] = true
			if !propertyProvided(entries, p) {
				out = append(out, p)
			}
		}
	}
	sort.Strings(out)
	return out
}

// propertyProvided reports whether an effective enricher owns the property. The
// http namespace accepts any property under its prefix, since the exposed
// headers are runtime config.
func propertyProvided(entries []enricher, property string) bool {
	for i := range entries {
		e := &entries[i]
		if e.namespace == entity.EnricherNamespaceHTTP && strings.HasPrefix(property, e.prefix) {
			return true
		}
		for _, p := range e.properties {
			if p == property {
				return true
			}
		}
	}
	return false
}

// validateEnricher validates an enricher definition for a write: the namespace
// must be a registered flag-scoped one and its config must pass the namespace's
// validation. Unknown or global namespaces are rejected (they are server-owned
// or nonexistent, and never stored on a flag).
func validateEnricher(e *entity.Enricher) error {
	if e == nil {
		return fmt.Errorf("enricher is required")
	}
	build, ok := enricherBuilders[e.Namespace]
	if !ok {
		return fmt.Errorf("unknown enricher namespace %q", e.Namespace)
	}
	built, err := build(e.ConfigJSON)
	if err != nil {
		return err
	}
	if built.scope != scopeFlag {
		return fmt.Errorf("enricher namespace %q is not flag-scoped", e.Namespace)
	}
	return nil
}

// effectiveEnricherModels maps a flag's visible effective enrichers to the API
// read model: its own flag-scoped enrichers (even disabled ones) plus the
// enabled server built-ins.
func effectiveEnricherModels(flag *entity.Flag) []*models.Enricher {
	entries := visibleEnrichers(effectiveEnrichers(flag))
	out := make([]*models.Enricher, 0, len(entries))
	for i := range entries {
		e := &entries[i]
		namespace := e.namespace
		enabled := e.enabled
		out = append(out, &models.Enricher{
			Namespace:  &namespace,
			Scope:      string(e.scope),
			Enabled:    &enabled,
			Properties: e.properties,
			Config:     e.config,
		})
	}
	return out
}

// mapFlagWithEnrichers maps a flag and attaches its effective enricher catalog.
// It is the default for the e2rMapFlag indirection in crud.go; tests can still
// stub that var to bypass mapping entirely.
func mapFlagWithEnrichers(f *entity.Flag) (*models.Flag, error) {
	r, err := e2r.MapFlag(f)
	if err != nil {
		return nil, err
	}
	r.Enrichers = effectiveEnricherModels(f)
	return r, nil
}

// mapFlagsWithEnrichers maps flags and attaches each effective enricher catalog.
func mapFlagsWithEnrichers(fs []entity.Flag) ([]*models.Flag, error) {
	ret := make([]*models.Flag, len(fs))
	for i := range fs {
		r, err := mapFlagWithEnrichers(&fs[i])
		if err != nil {
			return nil, err
		}
		ret[i] = r
	}
	return ret, nil
}
