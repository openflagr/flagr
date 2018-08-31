package handler

import (
	"github.com/checkr/flagr/pkg/entity"
	"github.com/checkr/flagr/swagger_gen/restapi/operations"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/evaluation"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/export"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/health"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/variant"
	"github.com/go-openapi/runtime/middleware"
)

var getDB = entity.GetDB

// Setup initialize all the handler functions
func Setup(api *operations.FlagrAPI) {
	setupCRUD(api)
	setupEvaluation(api)
	setupHealth(api)
	setupExport(api)
}

func setupCRUD(api *operations.FlagrAPI) {
	c := NewCRUD()
	// flags
	api.FlagFindFlagsHandler = flag.FindFlagsHandlerFunc(c.FindFlags)
	api.FlagCreateFlagHandler = flag.CreateFlagHandlerFunc(c.CreateFlag)
	api.FlagGetFlagHandler = flag.GetFlagHandlerFunc(c.GetFlag)
	api.FlagPutFlagHandler = flag.PutFlagHandlerFunc(c.PutFlag)
	api.FlagDeleteFlagHandler = flag.DeleteFlagHandlerFunc(c.DeleteFlag)
	api.FlagSetFlagEnabledHandler = flag.SetFlagEnabledHandlerFunc(c.SetFlagEnabledState)
	api.FlagGetFlagSnapshotsHandler = flag.GetFlagSnapshotsHandlerFunc(c.GetFlagSnapshots)

	// segments
	api.SegmentCreateSegmentHandler = segment.CreateSegmentHandlerFunc(c.CreateSegment)
	api.SegmentFindSegmentsHandler = segment.FindSegmentsHandlerFunc(c.FindSegments)
	api.SegmentPutSegmentHandler = segment.PutSegmentHandlerFunc(c.PutSegment)
	api.SegmentDeleteSegmentHandler = segment.DeleteSegmentHandlerFunc(c.DeleteSegment)
	api.SegmentPutSegmentsReorderHandler = segment.PutSegmentsReorderHandlerFunc(c.PutSegmentsReorder)

	// constraints
	api.ConstraintCreateConstraintHandler = constraint.CreateConstraintHandlerFunc(c.CreateConstraint)
	api.ConstraintFindConstraintsHandler = constraint.FindConstraintsHandlerFunc(c.FindConstraints)
	api.ConstraintPutConstraintHandler = constraint.PutConstraintHandlerFunc(c.PutConstraint)
	api.ConstraintDeleteConstraintHandler = constraint.DeleteConstraintHandlerFunc(c.DeleteConstraint)

	// distributions
	api.DistributionFindDistributionsHandler = distribution.FindDistributionsHandlerFunc(c.FindDistributions)
	api.DistributionPutDistributionsHandler = distribution.PutDistributionsHandlerFunc(c.PutDistributions)

	// variants
	api.VariantCreateVariantHandler = variant.CreateVariantHandlerFunc(c.CreateVariant)
	api.VariantFindVariantsHandler = variant.FindVariantsHandlerFunc(c.FindVariants)
	api.VariantPutVariantHandler = variant.PutVariantHandlerFunc(c.PutVariant)
	api.VariantDeleteVariantHandler = variant.DeleteVariantHandlerFunc(c.DeleteVariant)
}

func setupEvaluation(api *operations.FlagrAPI) {
	ec := GetEvalCache()
	ec.Start()

	e := NewEval()
	api.EvaluationPostEvaluationHandler = evaluation.PostEvaluationHandlerFunc(e.PostEvaluation)
	api.EvaluationPostEvaluationBatchHandler = evaluation.PostEvaluationBatchHandlerFunc(e.PostEvaluationBatch)
}

func setupHealth(api *operations.FlagrAPI) {
	api.HealthGetHealthHandler = health.GetHealthHandlerFunc(
		func(health.GetHealthParams) middleware.Responder { return &health.GetHealthOK{} },
	)
}

func setupExport(api *operations.FlagrAPI) {
	api.ExportGetExportSqliteHandler = export.GetExportSqliteHandlerFunc(exportSQLiteHandler)
}
