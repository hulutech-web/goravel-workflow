package facades

import (
	"github.com/hulutech-web/goravel-workflow"
	"github.com/hulutech-web/goravel-workflow/contracts"
	"log"
)

func Workflow() contracts.Workflow {
	instance, err := workflow.App.Make(workflow.Binding)
	if err != nil {
		log.Println(err)
		return nil
	}

	return instance.(contracts.Workflow)
}
