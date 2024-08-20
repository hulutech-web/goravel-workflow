package official_plugins

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
)

type DistributeController struct {
	//Dependent services
}

func NewDeptController() *DistributeController {
	return &DistributeController{
		//Inject services
	}
}

type DistributeRequest struct {
	FlowID    uint `json:"flow_id" form:"flow_id"`
	ProcessID uint `json:"process_id" form:"process_id"`
	Rules     Rule `json:"rules" form:"rules"`
}

// 开发者提交插件信息，产出插件
func (r *DistributeController) Product(ctx http.Context) http.Response {
	var distributeRequest DistributeRequest
	ctx.Request().Bind(&distributeRequest)
	var pluginConfig PluginConfig
	err := facades.Orm().Query().Model(&PluginConfig{}).Create(&pluginConfig)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "制作成功", err)
	}
	return httpfacades.NewResult(ctx).Success("制作成功", pluginConfig)
}

// 流程绑定插件
func (r *DistributeController) Bind(ctx http.Context) http.Response {
	plugin_id := ctx.Request().QueryInt("plugin_id")
	flow_id := ctx.Request().QueryInt("flow_id")
	process_id := ctx.Request().QueryInt("process_id")
	err := facades.Orm().Query().Model(&PluginConfig{}).Create(&PluginConfig{
		PluginID:  uint(plugin_id),
		FlowID:    uint(flow_id),
		ProcessID: uint(process_id),
	})
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "绑定失败", err)
	}
	return httpfacades.NewResult(ctx).Success("绑定成功", "")
}

// 开发者安装插件
func (r *DistributeController) Install(ctx http.Context) http.Response {
	var plugin Plugin
	ctx.Request().Bind(&plugin)
	err := facades.Orm().Query().Model(&Plugin{}).Create(&plugin)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "创建失败", err)
	}
	return httpfacades.NewResult(ctx).Success("创建成功", "")
}
