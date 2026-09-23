package handler

import (
	"context"
	"net/http"
	"sort"
	"strings"

	"github.com/openflagr/flagr/pkg/entity"
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
// catalog the UI lists and the set the eval pipeline runs.
func enabledEnrichers(entries []enricher) []enricher {
	out := make([]enricher, 0, len(entries))
	for i := range entries {
		if entries[i].enabled {
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
