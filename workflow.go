package workflow

import "github.com/goravel/framework/contracts/http"

type Workflow struct {
	Context http.Context
}

func NewWorkflow(ctx http.Context) *Workflow {
	return &Workflow{
		Context: ctx,
	}
}
