package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("workflow", map[string]any{
		"Code":    200,
		"Message": "返回成功",
	})
}
