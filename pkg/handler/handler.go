package handler

import (
	"github.com/checkr/flagr/swagger_gen/restapi/operations"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/evaluation"
	"github.com/checkr/flagr/swagger_gen/restapi/operations/flag"
	"github.com/davecgh/go-spew/spew"
)

func adhocTest() {
	c := NewCRUD()
	spew.Dump(c.CreateFlag(flag.NewCreateFlagParams()))
}

// Setup initialize all the hanlder functions
func Setup(api *operations.FlagrAPI) {

	c := NewCRUD()
	api.FlagFindFlagsHandler = flag.FindFlagsHandlerFunc(c.FindFlags)
	api.FlagCreateFlagHandler = flag.CreateFlagHandlerFunc(c.CreateFlag)
	api.FlagGetFlagHandler = flag.GetFlagHandlerFunc(c.GetFlag)

	e := NewEval()
	api.EvaluationPostEvaluationHandler = evaluation.PostEvaluationHandlerFunc(e.PostEvaluation)

	adhocTest()
}
