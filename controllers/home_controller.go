package controllers

import (
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"goravel/app/models"
)

type HomeController struct {
	//Dependent services
}

func NewHomeController() *HomeController {
	return &HomeController{
		//Inject services
	}
}

func (r *HomeController) Index(ctx http.Context) http.Response {
	entries := []models.Entry{}
	user := models.User{}
	facades.Auth(ctx).User(&user)
	emp := models.Emp{}
	facades.Orm().Query().Model(&models.Emp{}).Where("user_id=?", user.ID).First(&emp)
	query := facades.Orm().Query()
	//我的申请
	query.Model(&models.Entry{}).With("Emp").With("Flow").With("Procs", func(q orm.Query) orm.Query {
		return q.Order("id desc").Limit(1)
	}).With("Process").Where("emp_id=?", user.ID).Where("pid=?", 0).Order("id desc").Find(&entries)

	//我的代办
	procs := []models.Proc{}
	query.Model(&models.Proc{}).With("Emp").With("Entry", func(query orm.Query) orm.Query {
		return query.With("Emp")
	}).Where("emp_id=?", user.ID).Where("status=?", 0).
		Order("is_read asc").Order("status asc").Order("id desc").Find(&procs)

	//工作流
	flows := []models.Flow{}
	query.Model(&models.Flow{}).Where("is_publish=?", 1).Where("is_show=?", 1).Find(&flows)
	//待处理
	handle_procs := []models.Proc{}
	facades.Orm().Query().Model(&models.Proc{}).With("Emp").With("Entry").
		Where("emp_id=?", emp.ID).
		Where("status!=?", 0).
		Order("id desc").Find(&handle_procs)
	//query.Model(&models.Proc{}).With("Emp").With("Entry", func(query orm.Query) {
	//	query.With("Emp")
	//}).Where("emp_id=?", user.ID).Where("status !=?", 0).Order("entry_id desc").
	//	Order("id asc").Group("entry_id").Find(&handle_procs)
	return httpfacades.NewResult(ctx).Success("", map[string]interface{}{
		"entries":      entries,
		"procs":        procs,
		"flows":        flows,
		"handle_procs": handle_procs,
	})
}
