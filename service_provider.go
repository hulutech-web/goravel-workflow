package workflow

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	commands "github.com/hulutech-web/goravel-workflow/commands"
	"github.com/hulutech-web/goravel-workflow/routes"
)

const Binding = "workflow"

var App foundation.Application

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	App = app

	//配置文件
	app.Bind(Binding, func(app foundation.Application) (any, error) {
		config := app.MakeConfig()
		config.Add("workflow", map[string]any{
			"Dept": "Department", //部门关联应用中的模型
			"Emp":  "User",       //员工关联应用中的模型
		})
		return NewWorkflow(nil), nil
	})
	//	路由
	routes.Api(app)

	//	数据库迁移
	app.Publishes("github.com/hulutech-web/goravel-workflow", map[string]string{
		"migrations": app.DatabasePath("migrations"),
		"seeders":    app.DatabasePath("seeders"),
		"factories":  app.DatabasePath("factories"),
	})

	// 配置文件
	app.Publishes("github.com/hulutech-web/goravel-workflow", map[string]string{
		"config/workflow.go": app.ConfigPath("workflow.go"),
	})

	app.Commands([]console.Command{
		commands.NewPlugin(),
	})

}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]console.Command{
		commands.NewPublishWorkflow(),
	})

}
