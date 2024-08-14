package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("workflow", map[string]any{
		"Dept": "Department", //部门关联应用中的模型
		"Emp":  "User",       //员工关联应用中的模型
	})
}
