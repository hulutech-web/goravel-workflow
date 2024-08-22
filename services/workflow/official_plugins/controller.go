package official_plugins

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
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

// 为流程选择插件
func (r *DistributeController) InstallPlugin(ctx http.Context) http.Response {
	type SelRequest struct {
		FlowID   int `json:"flow_id" form:"flow_id"`
		PluginID int `json:"plugin_id" form:"plugin_id"`
	}
	var selRequest SelRequest
	ctx.Request().Bind(&selRequest)
	var flow models.Flow
	facades.Orm().Query().Model(&flow).Where("id=?", selRequest.FlowID).Find(&flow)
	var plugin Plugin
	facades.Orm().Query().Model(&Plugin{}).Where("id=?", selRequest.PluginID).Find(&plugin)

	facades.Orm().Query().Model(&flow).Association("Plugins").Append(&plugin)
	return httpfacades.NewResult(ctx).Success("安装成功", "")
}

func (r *DistributeController) List(ctx http.Context) http.Response {
	var plugins []Plugin
	err := facades.Orm().Query().Model(&plugins).With("PluginConfigs").Find(&plugins)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "获取失败", err)
	}
	return httpfacades.NewResult(ctx).Success("", plugins)
}

// 开发者提交插件信息，通过设计生成插件的选项
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
