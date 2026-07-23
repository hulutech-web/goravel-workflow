package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
	workflow "github.com/hulutech-web/goravel-workflow/services/workflow"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
)

type ProcController struct {
	workflow *workflow.Workflow
}

func NewProcController() *ProcController {
	return &ProcController{
		workflow: workflow.NewBaseWorkflow(),
	}
}

func (r *ProcController) Index(ctx http.Context) http.Response {
	entry_id := ctx.Request().InputInt("entry_id")
	var procs []models.Proc
	facades.Orm().Query().Model(&models.Proc{}).Where("entry_id=?", entry_id).With("Entry.Emp").Find(&procs)
	return httpfacades.NewResult(ctx).Success("", procs)
}

func (r *ProcController) Pass(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	process_id := ctx.Request().InputInt("process_id")
	content := ctx.Request().Input("content")
	err := r.workflow.Pass(process_id, user, content)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "审批失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("审批成功", nil)
}

func (r *ProcController) UnPass(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	withUser := models.Emp{}
	facades.Orm().Query().Model(&models.Emp{}).Where("id=?", user.ID).With("Dept").Find(&withUser)
	proc_id := ctx.Request().InputInt("proc_id")
	content := ctx.Request().Input("content")

	r.workflow.UnPass(proc_id, withUser, content)
	return httpfacades.NewResult(ctx).Success("驳回成功", nil)
}

// Revoke 撤回功能
func (r *ProcController) Revoke(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	entry_id := ctx.Request().InputInt("entry_id")
	err := r.workflow.Revoke(uint(entry_id), user)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "撤回失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("撤回成功", nil)
}

// AddSign 加签功能
func (r *ProcController) AddSign(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	entry_id := ctx.Request().InputInt("entry_id")
	process_id := ctx.Request().InputInt("process_id")
	sign_emp_id := ctx.Request().InputInt("sign_emp_id")
	sign_type := ctx.Request().Input("sign_type") // "before" or "after"

	err := r.workflow.AddSign(uint(entry_id), process_id, sign_emp_id, sign_type, user)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "加签失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("加签成功", nil)
}

// TransferProc 转交功能
func (r *ProcController) TransferProc(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	entry_id := ctx.Request().InputInt("entry_id")
	proc_id := ctx.Request().InputInt("proc_id")
	target_emp_id := ctx.Request().InputInt("target_emp_id")

	err := r.workflow.TransferProc(uint(entry_id), uint(proc_id), target_emp_id, user)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "转交失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("转交成功", nil)
}

// AddComment 评论功能
func (r *ProcController) AddComment(ctx http.Context) http.Response {
	var user models.Emp
	facades.Auth(ctx).User(&user)
	var emp models.Emp
	facades.Orm().Query().Model(&models.Emp{}).Where("user_id=?", user.ID).Find(&emp)

	entry_id := ctx.Request().InputInt("entry_id")
	proc_id := ctx.Request().InputInt("proc_id")
	content := ctx.Request().Input("content")

	err := r.workflow.AddComment(uint(entry_id), uint(proc_id), cast.ToInt(emp.ID), emp.Name, content)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "评论失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("评论成功", nil)
}

// GetComments 获取评论列表
func (r *ProcController) GetComments(ctx http.Context) http.Response {
	entry_id := ctx.Request().InputInt("entry_id")
	comments, err := r.workflow.GetComments(uint(entry_id))
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "获取评论失败", err.Error())
	}
	return httpfacades.NewResult(ctx).Success("", comments)
}
