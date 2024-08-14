package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
	"goravel/app/models"
)

type EmpController struct {
	//Dependent services
}

func NewEmpController() *EmpController {
	return &EmpController{
		//Inject services
	}
}

func (r *EmpController) Index(ctx http.Context) http.Response {
	emps := []models.Emp{}
	queries := ctx.Request().Queries()
	result, _ := httpfacades.NewResult(ctx).SearchByParams(queries).ResultPagination(&emps, []string{"Dept"}...)
	return result
}

func (r *EmpController) Show(ctx http.Context) http.Response {
	return nil
}

func (r *EmpController) Store(ctx http.Context) http.Response {
	return nil
}

func (r *EmpController) Update(ctx http.Context) http.Response {
	return nil
}

func (r *EmpController) Destroy(ctx http.Context) http.Response {
	return nil
}

func (r *EmpController) Search(ctx http.Context) http.Response {
	name := ctx.Request().Input("name", "")
	emps := []models.Emp{}
	facades.Orm().Query().Model(&models.Emp{}).Where("name like ?", "%"+name+"%").OrWhere("workno like ?", "%"+name+"%").Find(&emps)
	return httpfacades.NewResult(ctx).Success("", emps)
}

func (r *EmpController) Options(ctx http.Context) http.Response {
	emps := []models.Emp{}
	facades.Orm().Query().Model(&models.Emp{}).Find(&emps)
	return httpfacades.NewResult(ctx).Success("", emps)
}

// 员工绑定用户
func (r *EmpController) BindUser(ctx http.Context) http.Response {
	emp_id := ctx.Request().InputInt("emp_id")
	user_id := ctx.Request().InputInt("user_id")
	facades.Orm().Query().Model(&models.Emp{}).Where("id=?", emp_id).Update("user_id", cast.ToUint(user_id))
	return httpfacades.NewResult(ctx).Success("绑定成功", nil)
}
