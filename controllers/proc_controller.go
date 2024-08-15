package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
	workflow "github.com/hulutech-web/goravel-workflow/services/workflow"
	httpfacades "github.com/hulutech-web/http_result"
)

type ProcController struct {
	//Dependent services
	workflow *workflow.Workflow
}

func NewProcController() *ProcController {
	return &ProcController{
		workflow: workflow.NewWorkflow(),
	}
}

func (r *ProcController) Index(ctx http.Context) http.Response {
	entry_id := ctx.Request().QueryInt("entry_id")
	var procs []models.Proc
	facades.Orm().Query().Model(&models.Proc{}).Where("entry_id=?", entry_id).With("Entry.Emp").Find(&procs)
	return httpfacades.NewResult(ctx).Success("", procs)
}

func (r *ProcController) Children(ctx http.Context) http.Response {
	return nil
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
	facades.Orm().Query().Model(&models.Emp{}).Where("id=?", user.ID).With("Dept").First(&withUser)
	proc_id := ctx.Request().InputInt("proc_id")
	content := ctx.Request().Input("content")

	r.workflow.UnPass(proc_id, withUser, content)
	return httpfacades.NewResult(ctx).Success("驳回成功", nil)
}
