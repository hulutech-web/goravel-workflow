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

	// 配置文件
	config := app.MakeConfig()
	config.Add("workflow", map[string]any{
		"Dept": "Department", //部门关联应用中的模型
		"Emp":  "User",       //员工关联应用中的模型
	})
	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return NewWorkflow(nil), nil
	})

	// 路由
	routes.Api(app)

	// 数据库迁移 — 发布 Go 迁移文件（带标签）
	app.Publishes("github.com/hulutech-web/goravel-workflow", map[string]string{
		"migrations/2024_06_24_000000_create_workflow_base_tables.go": app.DatabasePath("migrations/2024_06_24_000000_create_workflow_base_tables.go"),
		"migrations/2026_07_23_000000_add_workflow_features.go":       app.DatabasePath("migrations/2026_07_23_000000_add_workflow_features.go"),
	}, "migrations")

	// 种子数据
	app.Publishes("github.com/hulutech-web/goravel-workflow", map[string]string{
		"seeders/workflow_seeder.go":          app.DatabasePath("seeders/workflow_seeder.go"),
		"seeders/workflow_dept_seeder.go":     app.DatabasePath("seeders/workflow_dept_seeder.go"),
		"seeders/workflow_emp_seeder.go":      app.DatabasePath("seeders/workflow_emp_seeder.go"),
		"seeders/workflow_flowtype_seeder.go": app.DatabasePath("seeders/workflow_flowtype_seeder.go"),
	}, "seeders")

	// 配置文件
	app.Publishes("github.com/hulutech-web/goravel-workflow", map[string]string{
		"config/workflow.go": app.ConfigPath("workflow.go"),
	}, "config")

	app.Commands([]console.Command{
		commands.NewPlugin(),
	})

}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Commands([]console.Command{
		commands.NewPublishWorkflow(),
	})

}
