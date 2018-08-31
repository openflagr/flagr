package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/export"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

var exportSQLiteHandler = func(export.GetExportSqliteParams) middleware.Responder {
	f, done, err := exportSQLiteFile()
	defer done()
	if err != nil {
		return export.NewGetExportSqliteDefault(500).WithPayload(ErrorMessage("%s", err))
	}
	return export.NewGetExportSqliteOK().WithPayload(f)
}

var exportSQLiteFile = func() (file io.ReadCloser, done func(), err error) {
	fname := fmt.Sprintf("/tmp/flagr_%d.sqlite", rand.Int31())
	done = func() {
		os.Remove(fname)
		logrus.WithField("file", fname).Debugf("removing the tmp file")
	}

	tmpDB := entity.NewSQLiteDB(fname)
	defer tmpDB.Close()

	if err := exportFlags(tmpDB); err != nil {
		return nil, done, err
	}
	if err := exportFlagSnapshots(tmpDB); err != nil {
		return nil, done, err
	}

	content, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, done, err
	}
	file = ioutil.NopCloser(bytes.NewReader(content))
	return file, done, nil
}

var exportFlags = func(tmpDB *gorm.DB) error {
	flags, err := fetchAllFlags()
	if err != nil {
		return err
	}
	for _, f := range flags {
		if err := f.Create(tmpDB); err != nil {
			return err
		}
	}
	logrus.WithField("count", len(flags)).Debugf("export flags")
	return nil
}

var exportFlagSnapshots = func(tmpDB *gorm.DB) error {
	qs := entity.NewFlagSnapshotQuerySet(getDB())
	var snapshots []entity.FlagSnapshot
	if err := qs.All(&snapshots); err != nil {
		return err
	}
	for _, s := range snapshots {
		if err := s.Create(tmpDB); err != nil {
			return err
		}
	}
	logrus.WithField("count", len(snapshots)).Debugf("export flag snapshots")
	return nil
}
