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
	query := facades.Orm().Query()
	query.Model(&flow).Where("id=?", selRequest.FlowID).Find(&flow)
	var plugin Plugin
	facades.Orm().Query().Model(&Plugin{}).Where("id=?", selRequest.PluginID).Find(&plugin)
	if flow.ID == 0 || plugin.ID == 0 {
		return httpfacades.NewResult(ctx).Error(500, "流程或插件不存在", "")
	}
	query.Model(&FlowPlugin{}).Create(&FlowPlugin{
		FlowID:   uint(selRequest.FlowID),
		PluginID: uint(selRequest.PluginID),
	})
	query.Model(&flow).Association("Plugins").Append(&plugin)
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

// 添加插件规则
func (r *DistributeController) StorePluginConfig(ctx http.Context) http.Response {
	type PluginConfigRequest struct {
		FieldID   int  `json:"field_id" form:"field_id"`
		FlowID    uint `json:"flow_id" form:"flow_id"`
		PluginID  uint `json:"plugin_id" form:"plugin_id"`
		ProcessID uint `json:"process_id" form:"process_id"`
		Rules     Rule `json:"rules" form:"rules"`
	}
	var pluginConfigRequest PluginConfigRequest
	ctx.Request().Bind(&pluginConfigRequest)
	//创建或者更新
	facades.Orm().Query().UpdateOrCreate(&PluginConfig{}, PluginConfig{
		PluginID:  pluginConfigRequest.PluginID,
		FlowID:    pluginConfigRequest.FlowID,
		ProcessID: pluginConfigRequest.ProcessID,
		FieldID:   pluginConfigRequest.FieldID,
	}, PluginConfig{Rules: pluginConfigRequest.Rules})
	return httpfacades.NewResult(ctx).Success("保存成功", "")
}

func (r *DistributeController) GetPluginConfig(ctx http.Context) http.Response {
	type PluginConfigRequest struct {
		FieldID   int  `json:"field_id" form:"field_id"`
		FlowID    uint `json:"flow_id" form:"flow_id"`
		PluginID  uint `json:"plugin_id" form:"plugin_id"`
		ProcessID uint `json:"process_id" form:"process_id"`
		Rules     Rule `json:"rules" form:"rules"`
	}
	var pluginConfigRequest PluginConfigRequest
	ctx.Request().Bind(&pluginConfigRequest)
	var pluginConfig PluginConfig
	facades.Orm().Query().Model(&PluginConfig{}).
		Where("field_id=?", pluginConfigRequest.FieldID).
		Where("flow_id=?", pluginConfigRequest.FlowID).
		Where("plugin_id=?", pluginConfigRequest.PluginID).
		Where("process_id=?", pluginConfigRequest.ProcessID).Find(&pluginConfig)
	return httpfacades.NewResult(ctx).Success("", pluginConfig)
}

func (r *DistributeController) GetAllPluginConfig(ctx http.Context) http.Response {
	type PluginConfigRequest struct {
		FieldID   int  `json:"field_id" form:"field_id"`
		FlowID    uint `json:"flow_id" form:"flow_id"`
		PluginID  uint `json:"plugin_id" form:"plugin_id"`
		ProcessID uint `json:"process_id" form:"process_id"`
		Rules     Rule `json:"rules" form:"rules"`
	}
	var pluginConfigRequest PluginConfigRequest
	ctx.Request().Bind(&pluginConfigRequest)
	var pluginConfigs []PluginConfig
	facades.Orm().Query().Model(&PluginConfig{}).With("Flow").With("Process").With("TemplateForm").
		Where("flow_id=?", pluginConfigRequest.FlowID).
		Where("plugin_id=?", pluginConfigRequest.PluginID).
		Find(&pluginConfigs)
	return httpfacades.NewResult(ctx).Success("", pluginConfigs)
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

// 卸载插件
func (r *DistributeController) UninstallPlugin(ctx http.Context) http.Response {
	type SelRequest struct {
		FlowID   int `json:"flow_id" form:"flow_id"`
		PluginID int `json:"plugin_id" form:"plugin_id"`
	}
	var selRequest SelRequest
	ctx.Request().Bind(&selRequest)
	var flow models.Flow
	query := facades.Orm().Query()
	query.Model(&flow).Where("id=?", selRequest.FlowID).Find(&flow)
	var plugin Plugin
	facades.Orm().Query().Model(&Plugin{}).Where("id=?", selRequest.PluginID).Find(&plugin)
	if flow.ID == 0 || plugin.ID == 0 {
		return httpfacades.NewResult(ctx).Error(500, "流程或插件不存在", "")
	}
	query.Model(&FlowPlugin{}).Where("flow_id=?", selRequest.FlowID).Where("plugin_id=?", selRequest.FlowID).
		Delete(&FlowPlugin{})
	query.Model(&flow).Association("Plugins").Delete(&plugin)
	return httpfacades.NewResult(ctx).Success("卸载成功", "")
}
