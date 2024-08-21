package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
	"github.com/hulutech-web/goravel-workflow/requests"
	httpfacades "github.com/hulutech-web/http_result"
)

type TemplateformController struct {
	//Dependent services
}

func NewTemplateformController() *TemplateformController {
	return &TemplateformController{
		//Inject services
	}
}

func (r *TemplateformController) Index(ctx http.Context) http.Response {
	template_id := ctx.Request().RouteInt("id")
	template_forms := []models.TemplateForm{}
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("template_id=?", template_id).
		Order("sort asc").Order("id desc").Find(&template_forms)
	return httpfacades.NewResult(ctx).Success("", template_forms)
}

func (r *TemplateformController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var templateform models.TemplateForm
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("id=?", id).Find(&templateform)
	return httpfacades.NewResult(ctx).Success("", templateform)
}

func (r *TemplateformController) Store(ctx http.Context) http.Response {
	//validator, _ := facades.Validation().Make(map[string]any{
	//	"template_id":         ctx.Request().InputInt("template_id"),
	//	"field":               ctx.Request().Input("field"),
	//	"field_name":          ctx.Request().Input("field_name"),
	//	"field_type":          ctx.Request().Input("field_type"),
	//	"field_value":         ctx.Request().Input("field_value"),
	//	"field_default_value": ctx.Request().Input("field_default_value"),
	//	"field_rules":         ctx.Request().Input("field_rules"),
	//	"sort":                ctx.Request().InputInt("sort"),
	//}, map[string]string{
	//	"template_id": "required",
	//	"field":       "required|alpha_rule",
	//	"field_name":  "required",
	//	"field_type":  "required",
	//}, validation.Messages(map[string]string{
	//	"template_id.required": "模板不能为空",
	//	"field.required":       "字段名称不能为空",
	//	"field_name.required":  "字段标题不能为空",
	//	"field_type.required":  "字段类型不能为空",
	//}))
	//if validator.Fails() {
	//	return httpfacades.NewResult(ctx).ValidError("参数错误", validator.Errors().All())
	//}
	var templateformRequest requests.TemplateformRequest
	errors, err := ctx.Request().ValidateRequest(&templateformRequest)
	if errors != nil || err != nil {
		return httpfacades.NewResult(ctx).ValidError("参数错误", errors.All())
	}
	tpform := models.TemplateForm{}
	tpform.TemplateID = templateformRequest.TemplateID
	tpform.Field = templateformRequest.Field
	tpform.FieldName = templateformRequest.FieldName
	tpform.FieldType = templateformRequest.FieldType
	tpform.FieldValue = templateformRequest.FieldValue
	tpform.FieldDefaultValue = templateformRequest.FieldDefaultValue
	tpform.Sort = templateformRequest.Sort
	tpform.FieldRules = templateformRequest.FieldRules
	if err != nil {
		return httpfacades.NewResult(ctx).ValidError("", errors.All())
	}
	facades.Orm().Query().Model(&models.TemplateForm{}).Create(&tpform)
	return httpfacades.NewResult(ctx).Success("创建成功", nil)
}

func (r *TemplateformController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var templateformRequest requests.TemplateformRequest
	errors, err := ctx.Request().ValidateRequest(&templateformRequest)
	if errors != nil || err != nil {
		return httpfacades.NewResult(ctx).ValidError("参数错误", errors.All())
	}

	existTpform := models.TemplateForm{}
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("id=?", id).Find(&existTpform)
	existTpform.TemplateID = templateformRequest.TemplateID
	existTpform.Field = templateformRequest.Field
	existTpform.FieldName = templateformRequest.FieldName
	existTpform.FieldType = templateformRequest.FieldType
	existTpform.FieldValue = templateformRequest.FieldValue
	existTpform.FieldDefaultValue = templateformRequest.FieldDefaultValue
	existTpform.Sort = templateformRequest.Sort
	existTpform.FieldRules = templateformRequest.FieldRules
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "参数错误", map[string]any{})
	}
	facades.Orm().Query().Save(&existTpform)
	return httpfacades.NewResult(ctx).Success("修改成功", nil)
}

func (r *TemplateformController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("id=?", id).Delete(&models.TemplateForm{})
	return httpfacades.NewResult(ctx).Success("删除成功", nil)
}
