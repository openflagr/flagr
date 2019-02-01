package internal // import "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql/internal"

import (
	"net"
	"strings"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

// ParseDSN parses various supported DSN types (currently mysql and postgres) into a
// map of key/value pairs which can be used as valid tags.
func ParseDSN(driverName, dsn string) (meta map[string]string, err error) {
	meta = make(map[string]string)
	switch driverName {
	case "mysql":
		meta, err = parseMySQLDSN(dsn)
		if err != nil {
			return
		}
	case "postgres":
		meta, err = parsePostgresDSN(dsn)
		if err != nil {
			return
		}
	default:
		// not supported
	}
	return reduceKeys(meta), nil
}

// reduceKeys takes a map containing parsed DSN information and returns a new
// map containing only the keys relevant as tracing tags, if any.
func reduceKeys(meta map[string]string) map[string]string {
	var keysOfInterest = map[string]string{
		"user":             ext.DBUser,
		"application_name": ext.DBApplication,
		"dbname":           ext.DBName,
		"host":             ext.TargetHost,
		"port":             ext.TargetPort,
	}
	m := make(map[string]string)
	for k, v := range meta {
		if nk, ok := keysOfInterest[k]; ok {
			m[nk] = v
		}
	}
	return m
}

// parseMySQLDSN parses a mysql-type dsn into a map.
func parseMySQLDSN(dsn string) (m map[string]string, err error) {
	var cfg *mySQLConfig
	if cfg, err = mySQLConfigFromDSN(dsn); err == nil {
		host, port, _ := net.SplitHostPort(cfg.Addr)
		m = map[string]string{
			"user":   cfg.User,
			"host":   host,
			"port":   port,
			"dbname": cfg.DBName,
		}
		return m, nil
	}
	return nil, err
}

// parsePostgresDSN parses a postgres-type dsn into a map.
func parsePostgresDSN(dsn string) (map[string]string, error) {
	var err error
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// url form, convert to opts
		dsn, err = parseURL(dsn)
		if err != nil {
			return nil, err
		}
	}
	meta := make(map[string]string)
	if err := parseOpts(dsn, meta); err != nil {
		return nil, err
	}
	// remove sensitive information
	delete(meta, "password")
	return meta, nil
}
