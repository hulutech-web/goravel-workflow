package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
	httpfacades "github.com/hulutech-web/http_result"
)

type CcController struct{}

func NewCcController() *CcController {
	return &CcController{}
}

// List 获取当前用户的抄送列表
func (r *CcController) List(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	var emp models.Emp
	facades.Orm().Query().Model(&models.Emp{}).Where("user_id=?", user.ID).Find(&emp)

	var ccRecords []models.CcRecord
	facades.Orm().Query().Model(&models.CcRecord{}).Where("emp_id=?", emp.ID).Order("id desc").Find(&ccRecords)
	return httpfacades.NewResult(ctx).Success("", ccRecords)
}

// GetEntryCC 获取某流程的抄送记录
func (r *CcController) GetEntryCC(ctx http.Context) http.Response {
	entry_id := ctx.Request().InputInt("entry_id")
	var ccRecords []models.CcRecord
	facades.Orm().Query().Model(&models.CcRecord{}).Where("entry_id=?", entry_id).Order("id asc").Find(&ccRecords)
	return httpfacades.NewResult(ctx).Success("", ccRecords)
}
