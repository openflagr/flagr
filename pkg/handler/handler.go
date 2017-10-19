package handler

import (
	"encoding/gob"

	"github.com/checkr/flagr/pkg/config"
	"github.com/checkr/flagr/swagger_gen/restapi/operations"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/constraint"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/distribution"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/evaluation"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/segment"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/variant"
	raven "github.com/getsentry/raven-go"
	"github.com/zhouzhuojie/conditions"
)

// Setup initialize all the handler functions
func Setup(api *operations.FlagrAPI) {
	setupRaven()
	setupGob()
	setupCRUD(api)
	setupEvaluation(api)
}

func setupGob() {
	gob.Register(conditions.BinaryExpr{})
	gob.Register(conditions.VarRef{})
}

func setupRaven() {
	if config.Config.SentryEnabled {
		raven.SetDSN(config.Config.SentryDSN)
	}
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

	// segments
	api.SegmentCreateSegmentHandler = segment.CreateSegmentHandlerFunc(c.CreateSegment)
	api.SegmentFindSegmentsHandler = segment.FindSegmentsHandlerFunc(c.FindSegments)
	api.SegmentPutSegmentHandler = segment.PutSegmentHandlerFunc(c.PutSegment)
	api.SegmentDeleteSegmentHandler = segment.DeleteSegmentHandlerFunc(c.DeleteSegment)

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
}
