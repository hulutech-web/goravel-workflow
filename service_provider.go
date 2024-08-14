package workflow

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"goravel/packages/workflow/commands"
)

const Binding = "workflow"

var App foundation.Application

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.Bind(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		config.Add("workflow", map[string]any{
			"Dept": "Department", //部门关联应用中的模型
			"Emp":  "User",       //员工关联应用中的模型
		})
		return NewWorkflow(nil), nil
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]console.Command{
		commands.NewPublishWorkflow(),
	})
	app.Publishes("./packages/workflow", map[string]string{
		"config/workflow.go": app.ConfigPath("workflow.go"),
	})
}
