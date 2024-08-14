package facades

import (
	"log"

	"goravel/packages/workflow"
	"goravel/packages/workflow/contracts"
)

func Workflow() contracts.Workflow {
	instance, err := workflow.App.Make(workflow.Binding)
	if err != nil {
		log.Println(err)
		return nil
	}

	return instance.(contracts.Workflow)
}
