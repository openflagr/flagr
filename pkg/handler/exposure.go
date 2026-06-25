package handler

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/openflagr/flagr/pkg/config"
	"github.com/openflagr/flagr/swagger_gen/models"
	exposureapi "github.com/openflagr/flagr/swagger_gen/restapi/operations/exposure"
)

// Exposure handles POST /exposures.
type Exposure interface {
	PostExposures(exposureapi.PostExposuresParams) middleware.Responder
}

// NewExposure creates an Exposure handler.
func NewExposure() Exposure {
	return &exposureHandler{}
}

type exposureHandler struct{}

func (h *exposureHandler) PostExposures(params exposureapi.PostExposuresParams) middleware.Responder {
	if params.Body == nil || len(params.Body.Exposures) == 0 {
		return exposureapi.NewPostExposuresDefault(400).WithPayload(
			ErrorMessage("exposures array is required and must not be empty"))
	}

	exposures := params.Body.Exposures
	if max := config.Config.ExposureBatchSize; max > 0 && len(exposures) > max {
		return exposureapi.NewPostExposuresDefault(400).WithPayload(
			ErrorMessage("exposure batch size %d exceeds maximum allowed size of %d", len(exposures), max))
	}

	var logged int64
	var rowErrors []*models.ExposureRowError

	for i, row := range exposures {
		n, errRow := processExposureRow(int64(i), row)
		logged += n
		if errRow != nil {
			rowErrors = append(rowErrors, errRow)
		}
	}

	resp := exposureapi.NewPostExposuresOK()
	resp.SetPayload(&models.ExposuresResponse{
		LoggedCount: logged,
		Message:     "Exposures logged successfully",
		Errors:      rowErrors,
	})
	return resp
}

// processExposureRow validates one exposure and optionally records it to the data pipeline.
// Returns 1 if written to recorders, 0 otherwise, and a per-row error when validation fails.
func processExposureRow(index int64, row *models.Exposure) (logged int64, rowErr *models.ExposureRowError) {
	if row == nil {
		logExposureIngestStatsd("rejected", 0, "")
		return 0, exposureRowErr(index, "exposure row is null")
	}

	synthetic, err := validateAndBuildExposure(row)
	if err != nil {
		logExposureIngestStatsd("rejected", 0, "")
		return 0, exposureRowErr(index, err.Error())
	}

	logExposureIngestStatsd("accepted", synthetic.FlagID, synthetic.FlagKey)

	flag := GetEvalCache().GetByFlagKeyOrID(synthetic.FlagID)
	if !shouldRecordPipelineEvent(flag) {
		return 0, nil
	}

	recordPipelineEvent(synthetic)
	logExposureIngestStatsd("recorded", synthetic.FlagID, synthetic.FlagKey)
	return 1, nil
}

func exposureRowErr(index int64, message string) *models.ExposureRowError {
	return &models.ExposureRowError{Index: index, Message: message}
}

var logExposureIngestStatsd = func(status string, flagID int64, flagKey string) {
	if config.Global.StatsdClient == nil {
		return
	}
	tags := []string{fmt.Sprintf("status:%s", status)}
	if flagID > 0 {
		tags = append(tags, fmt.Sprintf("FlagID:%d", flagID))
	}
	if flagKey != "" {
		tags = append(tags, fmt.Sprintf("FlagKey:%s", flagKey))
	}
	metric := "exposure.ingest"
	if status == "recorded" {
		metric = "exposure.recorded"
	}
	config.Global.StatsdClient.Incr(metric, tags, 1)
}