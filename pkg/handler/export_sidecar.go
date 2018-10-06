package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/export"
	"github.com/go-openapi/runtime/middleware"
)

// SidecarFile struct
type SidecarFile struct {
	Timestamp time.Time `json:"timestamp"`

	// flags will be preloaded
	Flags []entity.Flag `json:"flags"`
}

var exportSidecarHandler = func(export.GetExportSidecarParams) middleware.Responder {
	fs := GetEvalCache().GetAll()
	content, err := json.Marshal(SidecarFile{
		Timestamp: time.Now().UTC(),
		Flags:     fs,
	})
	if err != nil {
		return export.NewGetExportSidecarDefault(500).WithPayload(ErrorMessage("%s", err))
	}

	file := ioutil.NopCloser(bytes.NewReader(content))
	defer file.Close()

	return export.NewGetExportSidecarOK().WithPayload(file)
}
