package commands

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/services/workflow/official_plugins"
)

type Plugin struct{}

func NewPlugin() *Plugin {
	return &Plugin{}
}

// Signature The name and signature of the console command.
func (receiver *Plugin) Signature() string {
	return "make:plugin"
}

// Description The console command description.
func (receiver *Plugin) Description() string {
	return "您正在创建一个流程框架插件"
}

// Extend The console command extend.
func (receiver *Plugin) Extend() command.Extend {
	return command.Extend{}
}

// Handle Execute the console command.
func (receiver *Plugin) Handle(ctx console.Context) error {
	name, _ := ctx.Ask("插件名称?")
	version, _ := ctx.Ask("插件版本?")
	description, _ := ctx.Ask("功能描述?")
	author, _ := ctx.Ask("插件作者?")
	if _isOk, err := ctx.Confirm("确认添加吗?", console.ConfirmOption{
		Default: true,
	}); err != nil && _isOk {
		err = facades.Orm().Query().Model(&official_plugins.Plugin{}).Create(&official_plugins.Plugin{
			Name:        name,
			Version:     version,
			Description: description,
			Author:      author,
		})
		return err
	}
	return nil
}
