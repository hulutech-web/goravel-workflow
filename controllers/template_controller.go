package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	httpfacades "github.com/hulutech-web/http_result"
	"goravel/app/models"
)

type TemplateController struct {
	//Dependent services
}

func NewTemplateController() *TemplateController {
	return &TemplateController{
		//Inject services
	}
}

func (r *TemplateController) Index(ctx http.Context) http.Response {
	temps := []models.Template{}
	queries := ctx.Request().Queries()
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries).ResultPagination(&temps)
	return result
}

func (r *TemplateController) Show(ctx http.Context) http.Response {
	return nil
}

func (r *TemplateController) Store(ctx http.Context) http.Response {
	validator, err := facades.Validation().Make(map[string]any{
		"template_name": ctx.Request().Input("template_name"),
	}, map[string]string{
		"template_name": "required|max_len:255"},
		validation.Messages(map[string]string{
			"template_name.required": "标题不能为空",
		}))
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "验证失败", err)
	}
	var template models.Template
	if validator.Fails() {
		return httpfacades.NewResult(ctx).ValidError("验证失败", validator.Errors().All())
	}
	template.TemplateName = ctx.Request().Input("template_name")
	err = facades.Orm().Query().Model(&models.Template{}).Create(&template)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "添加失败", err)
	}
	return httpfacades.NewResult(ctx).Success("添加成功", template)
}

func (r *TemplateController) Update(ctx http.Context) http.Response {
	return nil
}

func (r *TemplateController) Destroy(ctx http.Context) http.Response {
	return nil
}
