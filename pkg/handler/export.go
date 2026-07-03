package handler

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openflagr/flagr/pkg/entity"
	"github.com/openflagr/flagr/swagger_gen/restapi/operations/export"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var exportSQLiteHandler = func(p export.GetExportSqliteParams) middleware.Responder {
	f, done, err := exportSQLiteFile(p.ExcludeSnapshots)
	defer done()
	if err != nil {
		return export.NewGetExportSqliteDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return export.NewGetExportSqliteOK().WithPayload(f)
}

var exportSQLiteFile = func(excludeSnapshots *bool) (file io.ReadCloser, done func(), err error) {
	fname := path.Join(os.TempDir(), fmt.Sprintf("flagr_%d.sqlite", rand.Int31()))
	done = func() {
		os.Remove(fname)
		logrus.WithField("file", fname).Debugf("removing the tmp file")
	}

	tmpDB := entity.NewSQLiteDB(fname)
	sqlDB, dbErr := tmpDB.DB()
	if dbErr != nil {
		return nil, done, dbErr
	}
	defer sqlDB.Close()

	if err := exportFlags(tmpDB); err != nil {
		return nil, done, err
	}
	if excludeSnapshots == nil || !*excludeSnapshots {
		if err := exportFlagSnapshots(tmpDB); err != nil {
			return nil, done, err
		}
	}
	if err := exportFlagEntityTypes(tmpDB); err != nil {
		return nil, done, err
	}

	content, err := os.ReadFile(fname)
	if err != nil {
		return nil, done, err
	}
	file = io.NopCloser(bytes.NewReader(content))
	return file, done, nil
}

var exportFlags = func(tmpDB *gorm.DB) error {
	flags, err := fetchAllFlags()
	if err != nil {
		return err
	}
	for _, f := range flags {
		if err := tmpDB.Create(&f).Error; err != nil {
			return err
		}
	}
	logrus.WithField("count", len(flags)).Debugf("export flags")
	return nil
}

var exportFlagSnapshots = func(tmpDB *gorm.DB) error {
	var snapshots []entity.FlagSnapshot
	if err := getDB().Find(&snapshots).Error; err != nil {
		return err
	}
	for _, s := range snapshots {
		if err := tmpDB.Create(&s).Error; err != nil {
			return err
		}
	}
	logrus.WithField("count", len(snapshots)).Debugf("export flag snapshots")
	return nil
}

var exportFlagEntityTypes = func(tmpDB *gorm.DB) error {
	var ts []entity.FlagEntityType
	if err := getDB().Find(&ts).Error; err != nil {
		return err
	}
	for _, s := range ts {
		if err := tmpDB.Create(&s).Error; err != nil {
			return err
		}
	}
	logrus.WithField("count", len(ts)).Debugf("export flag entity types")
	return nil
}

var exportEvalCacheJSONHandler = func(p export.GetExportEvalCacheJSONParams) middleware.Responder {
	result := GetEvalCache().exportWithFilter(p)
	return export.NewGetExportEvalCacheJSONOK().WithPayload(result)
}

// exportWithFilter returns the eval cache JSON filtered by the request params.
func (ec *EvalCache) exportWithFilter(p export.GetExportEvalCacheJSONParams) EvalCacheJSON {
	flags := ec.export().Flags
	flags = filterFlags(flags, p)
	return EvalCacheJSON{Flags: flags}
}

// filterFlags applies AND predicates across all provided filter groups.
// Within ids/keys, values are OR'd.
func filterFlags(flags []entity.Flag, p export.GetExportEvalCacheJSONParams) []entity.Flag {
	if p.Ids == nil && p.Keys == nil && p.Enabled == nil && p.Tags == nil {
		return flags
	}

	var idFilter map[uint]bool
	if p.Ids != nil {
		idFilter = parseCSVUint(*p.Ids)
	}

	var keyFilter map[string]bool
	if p.Keys != nil {
		keyFilter = parseCSVString(*p.Keys)
	}

	var tagFilter []string
	if p.Tags != nil {
		tagFilter = parseCSVStrings(*p.Tags)
	}

	useAll := p.All != nil && *p.All

	result := make([]entity.Flag, 0, len(flags))
	for _, f := range flags {
		if !matchFlag(f, idFilter, keyFilter, p.Enabled, tagFilter, useAll) {
			continue
		}
		result = append(result, f)
	}
	return result
}

func matchFlag(f entity.Flag, ids map[uint]bool, keys map[string]bool, enabled *bool, tags []string, useAll bool) bool {
	if ids != nil && !ids[f.ID] {
		return false
	}
	if keys != nil && !keys[f.Key] {
		return false
	}
	if enabled != nil && f.Enabled != *enabled {
		return false
	}
	if len(tags) > 0 && !matchTags(f.Tags, tags, useAll) {
		return false
	}
	return true
}

func matchTags(flagTags []entity.Tag, filterTags []string, useAll bool) bool {
	tagSet := make(map[string]bool, len(flagTags))
	for _, t := range flagTags {
		tagSet[t.Value] = true
	}
	if useAll {
		for _, ft := range filterTags {
			if !tagSet[ft] {
				return false
			}
		}
		return true
	}
	// ANY
	for _, ft := range filterTags {
		if tagSet[ft] {
			return true
		}
	}
	return false
}

func parseCSVUint(s string) map[uint]bool {
	parts := strings.Split(s, ",")
	m := make(map[uint]bool, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if v, err := strconv.ParseUint(p, 10, 64); err == nil {
			m[uint(v)] = true
		}
	}
	return m
}

func parseCSVString(s string) map[string]bool {
	parts := strings.Split(s, ",")
	m := make(map[string]bool, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			m[p] = true
		}
	}
	return m
}

func parseCSVStrings(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
