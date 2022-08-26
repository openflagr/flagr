package handler

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"

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

var exportEvalCacheJSONHandler = func(export.GetExportEvalCacheJSONParams) middleware.Responder {
	return export.NewGetExportEvalCacheJSONOK().WithPayload(
		GetEvalCache().export(),
	)
}
