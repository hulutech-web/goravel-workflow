package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"goravel/app/models"
)

type DeptController struct {
	//Dependent services
}

func NewDeptController() *DeptController {
	return &DeptController{
		//Inject services
	}
}

func (r *DeptController) Index(ctx http.Context) http.Response {
	depts := []models.Dept{}
	tx, _ := facades.Orm().Query().Begin()
	tx.Model(&models.Dept{}).With("Director").With("Manager").Find(&depts)
	tx.Commit()
	deptInstance := models.Dept{}

	result := deptInstance.Recursion(depts, "|---", 0, 0)
	return httpfacades.NewResult(ctx).Success("", result)
}

func (r *DeptController) Show(ctx http.Context) http.Response {
	return nil
}

func (r *DeptController) Store(ctx http.Context) http.Response {
	return nil
}

func (r *DeptController) Update(ctx http.Context) http.Response {
	return nil
}

func (r *DeptController) Destroy(ctx http.Context) http.Response {
	return nil
}
func (r *DeptController) BindManager(ctx http.Context) http.Response {
	manager_id := ctx.Request().InputInt("manager_id")
	dept_id := ctx.Request().InputInt("dept_id")
	facades.Orm().Query().Model(&models.Dept{}).Where("id = ?", dept_id).Update("manager_id", manager_id)
	return httpfacades.NewResult(ctx).Success("设置成功", nil)
}

func (r *DeptController) BindDirector(ctx http.Context) http.Response {
	director_id := ctx.Request().InputInt("director_id")
	dept_id := ctx.Request().InputInt("dept_id")
	facades.Orm().Query().Model(&models.Dept{}).Where("id = ?", dept_id).Update("director_id", director_id)
	return httpfacades.NewResult(ctx).Success("设置成功", nil)
}
