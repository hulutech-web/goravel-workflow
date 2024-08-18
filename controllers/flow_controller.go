package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	"github.com/hulutech-web/goravel-workflow/models"
	httpfacades "github.com/hulutech-web/http_result"
)

type FlowController struct {
	//Dependent services
}

func NewFlowController() *FlowController {
	return &FlowController{
		//Inject services
	}
}

func (r *FlowController) Index(ctx http.Context) http.Response {
	flows := []models.Flow{}
	queries := ctx.Request().Queries()
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries).ResultPagination(&flows)
	return result
}

func (r *FlowController) List(ctx http.Context) http.Response {
	flows := []models.Flow{}
	facades.Orm().Query().Model(&models.Flow{}).Where("is_publish=?", 1).Find(&flows)
	return httpfacades.NewResult(ctx).Success("", flows)
}

func (r *FlowController) Create(ctx http.Context) http.Response {
	var templates []models.Template
	var flowtypes []models.Flowtype
	facades.Orm().Query().Model(&models.Template{}).Find(&templates)
	facades.Orm().Query().Model(&models.Flowtype{}).Find(&flowtypes)
	return httpfacades.NewResult(ctx).Success("", map[string]any{
		"templates": templates,
		"flowtypes": flowtypes,
	})
}

func (r *FlowController) Show(ctx http.Context) http.Response {
	return nil
}

func (r *FlowController) Store(ctx http.Context) http.Response {

	validator, _ := facades.Validation().Make(map[string]any{
		"flow_no":     ctx.Request().Input("flow_no"),
		"flow_name":   ctx.Request().Input("flow_name"),
		"template_id": ctx.Request().InputInt("template_id"),
		"type_id":     ctx.Request().InputInt("type_id"),
	}, map[string]string{
		"flow_no":     "required",
		"flow_name":   "required",
		"template_id": "required",
		"type_id":     "required",
	}, validation.Messages(map[string]string{
		"flow_no.required":     "编号不能为空",
		"flow_name.required":   "名称不能为空",
		"template_id.required": "模板不能为空",
		"type_id.required":     "类型不能为空",
	}))
	if validator.Fails() {
		return httpfacades.NewResult(ctx).ValidError("参数错误", validator.Errors().All())
	}
	flow := models.Flow{}
	err := validator.Bind(&flow)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "参数错误", map[string]any{})
	}
	facades.Orm().Query().Model(&models.Flow{}).Create(&flow)
	return httpfacades.NewResult(ctx).Success("创建成功", nil)
}

func (r *FlowController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var flow models.Flow
	ctx.Request().Bind(&flow)
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", id).Update(&flow)
	return httpfacades.NewResult(ctx).Success("保存成功", flow)

}

func (r *FlowController) Destroy(ctx http.Context) http.Response {
	return nil
}

func (r *FlowController) FlowDesign(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	flow := models.Flow{}
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", id).First(&flow)
	return httpfacades.NewResult(ctx).Success("", flow)
}

// Publish 发布流程
func (r *FlowController) Publish(ctx http.Context) http.Response {
	flow_id := ctx.Request().InputInt("flow_id")
	flow := models.Flow{}
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", flow_id).First(&flow)

	//如果设置了多个个开始步骤
	link_starts := []models.Flowlink{}
	facades.Orm().Query().Model(&models.Flowlink{}).Where("flow_id=?", flow_id).Where("position=?", 0).Find(&link_starts)
	if len(link_starts) > 1 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "发布失败，只能设置一个开始步骤", nil)
	}
	var fkCount1 int64
	facades.Orm().Query().Model(&models.Flowlink{}).Where("flow_id=?", flow_id).Where("type=?", "Condition").
		Count(&fkCount1)
	if fkCount1 <= 1 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "发布失败，至少需要两个步骤", nil)
	}

	var fkCount2 int64
	facades.Orm().Query().Model(&models.Flowlink{}).Where("flow_id=?", flow_id).Where("type=?", "Condition").
		Where("next_process_id=?", -1).Count(&fkCount2)
	if fkCount2 > 1 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "发布失败，有步骤没有创建连线", nil)
	}
	type Countf struct {
		Fid uint `json:"fid"`
		Pid uint `json:"pid"`
	}
	var flowlinkExists bool

	facades.Orm().Query().Table("flowlinks").
		Join("left join processes on flowlinks.process_id=processes.id").
		Where("flowlinks.flow_id=?", flow_id).Where("processes.position=?", 0).Exists(&flowlinkExists)
	if !flowlinkExists {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "发布失败，请设置结束步骤", nil)
	}
	flowlinks := []models.Flowlink{}
	facades.Orm().Query().Table("flowlinks").Select("flowlinks.*").
		Join("join processes on flowlinks.process_id=processes.id").
		Where("flowlinks.flow_id=?", flow_id).
		Where("flowlinks.type !=?", "Condition").
		Where("processes.position !=?", 0).
		Find(&flowlinks)
	for _, flowlink := range flowlinks {
		var cConditionMet bool
		facades.Orm().Query().Table("flowlinks").
			Join("join processes on flowlinks.process_id=processes.id").
			Where("flowlinks.flow_id=?", flow_id).
			Where("flowlinks.process_id=?", flowlink.ProcessID).
			Where("flowlinks.type !=?", "Condition").
			Where("processes.position !=?", 0).Exists(&cConditionMet)
		if !cConditionMet {
			return httpfacades.NewResult(ctx).
				Error(http.StatusInternalServerError, "发布失败，请给设置步骤审批权限", nil)
		}
	}

	flow.IsPublish = true
	facades.Orm().Query().Model(&models.Flow{}).Where("id=?", flow.ID).Save(&flow)

	return httpfacades.NewResult(ctx).Success("发布成功", flow)
}
