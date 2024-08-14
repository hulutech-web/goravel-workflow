package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/hulutech-web/goravel-workflow/controllers/common"
	"github.com/hulutech-web/goravel-workflow/models"
	"github.com/hulutech-web/goravel-workflow/requests"
	"strconv"
)

type UserController struct {
	//Dependent services
	*common.SearchByParamsService
}

type UserSel struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Mobile string `json:"mobile"`
}
type UserList struct {
	ID        uint            `json:"id"`
	Name      string          `gorm:"column:name;" json:"name"`
	AvatarUrl string          `gorm:"column:avatarUrl;" json:"avatarUrl"`
	IdNumber  string          `json:"id_number"`
	Mobile    string          `json:"mobile"`
	State     int             `json:"state"`
	CreatedAt carbon.DateTime `gorm:"autoUpdateTime;column:created_at"`
	IsMember  int             `json:"is_member"`
}

func NewUserController() *UserController {
	return &UserController{
		//Inject services
	}
}

func (r *UserController) Index(ctx http.Context) http.Response {
	pageSize := ctx.Request().Query("pageSize", "1")
	pageSizeInt, _ := strconv.Atoi(pageSize)
	order := ctx.Request().Query("order")
	sort := ctx.Request().Query("sort")
	var total int64
	currentPage := ctx.Request().Query("currentPage", "1")
	currentPageInt, _ := strconv.Atoi(currentPage)
	queries := ctx.Request().Queries()
	userLists := []UserList{}
	if order != "" || sort != "" {
		facades.Orm().Query().Model(&models.User{}).Scopes(r.SearchByParams(ctx.Request().Queries())).
			Order(sort+" "+order).Paginate(currentPageInt, pageSizeInt, &userLists, &total)
	} else {
		facades.Orm().Query().Model(&models.User{}).Scopes(r.SearchByParams(queries)).
			Order("id desc").Paginate(currentPageInt, pageSizeInt, &userLists, &total)
	}

	return ctx.Response().Success().Json(http.Json{
		"data":  userLists,
		"total": total,
		"links": map[string]interface{}{
			"first": "http://" + ctx.Request().Host() + "/api/admin/employee?pageSize=" + pageSize + "&currentPage=1",
			//最后一页应该是总数除以每页的数量取整
			"last": "http://" + ctx.Request().Host() + "/api/admin/employee?pageSize=" + pageSize + "&currentPage=" + strconv.Itoa(int(total)/pageSizeInt),
			"prev": "http://" + ctx.Request().Host() + "/api/admin/employee?pageSize=" + pageSize + "&currentPage=" + strconv.Itoa(currentPageInt-1),
			"next": "http://" + ctx.Request().Host() + "/api/admin/employee?pageSize=" + pageSize + "&currentPage=" + strconv.Itoa(currentPageInt+1),
		},
		"meta": map[string]interface{}{
			"total_page":   int(total) / pageSizeInt,
			"current_page": currentPageInt,
			"per_page":     pageSizeInt,
			"total":        total,
		},
	})
}

func (r *UserController) Store(ctx http.Context) http.Response {
	var userRequest requests.UserRequest
	errors, _ := ctx.Request().ValidateRequest(&userRequest)
	if errors != nil {
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{
			"errors": errors.All(),
		})
	}

	user := models.User{
		Name:      userRequest.Name,
		Mobile:    userRequest.Mobile,
		AvatarUrl: ctx.Request().Input("avatarUrl"),
		IdNumber:  ctx.Request().Input("idNumber"),
		Email:     ctx.Request().Input("email"),
		IsMember:  ctx.Request().InputInt("is_member"),
	}
	existUser := models.User{}
	facades.Orm().Query().Model(&models.User{}).Where("mobile=?", user.Mobile).First(&existUser)
	if existUser.ID != 0 {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"errors": "手机号已存在",
		})
	}
	facades.Orm().Query().Model(&models.User{}).Create(&user)
	return ctx.Response().Success().Json(http.Json{
		"message": "创建成功",
		"data":    user,
	})
}

// Update
func (r *UserController) Update(ctx http.Context) http.Response {
	idInt := ctx.Request().RouteInt("id")
	var userRequest requests.UserRequest
	errors, _ := ctx.Request().ValidateRequest(&userRequest)
	if errors != nil {
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{
			"errors": errors.All(),
		})
	}
	var user models.User
	facades.Orm().Query().Model(&models.User{}).Where("id=?", idInt).Where("mobile=?", userRequest.Mobile).First(&user)
	if user.ID == 0 {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"errors": "手机号不能修改或用户不存在",
		})
	}
	user.Name = userRequest.Name
	user.Mobile = userRequest.Mobile
	user.AvatarUrl = ctx.Request().Input("avatarUrl")

	user.IdNumber = ctx.Request().Input("idNumber")
	user.Email = ctx.Request().Input("email")
	user.IsMember = ctx.Request().InputInt("is_member")
	facades.Orm().Query().Model(&user).Where("id=?", idInt).Update(&user)
	return ctx.Response().Success().Json(http.Json{
		"message": "更新成功",
		"data":    user,
	})

}

// Show
func (r *UserController) Show(ctx http.Context) http.Response {
	idInt := ctx.Request().RouteInt("id")
	var user models.User
	facades.Orm().Query().Model(&models.User{}).Where("id=?", idInt).First(&user)
	return ctx.Response().Success().Json(http.Json{
		"data": user,
	})
}

// Destroy
func (r *UserController) Destroy(ctx http.Context) http.Response {
	idInt := ctx.Request().RouteInt("id")
	if _, err := facades.Orm().Query().Where("id", idInt).Delete(&models.User{}); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"message": err.Error(),
		})
	}
	return ctx.Response().Success().Json(http.Json{
		"message": "删除成功",
	})
}

// userselect
func (r *UserController) UserSel(ctx http.Context) http.Response {
	name := ctx.Request().Query("name", "")
	//如果name为空，则全部返回，否则模糊查询
	var total int64 = 10

	users := []models.User{}
	if name == "" {
		facades.Orm().Query().Model(&models.User{}).Paginate(1, 5, &users, &total)
	} else {
		facades.Orm().Query().Where("nickName like ?", "%"+name+"%").OrWhere("mobile like ?", "%"+name+"%").
			Paginate(1, 5, &users, &total)
	}
	userSels := []UserSel{}
	for _, user := range users {
		userSel := UserSel{}
		userSel.ID = user.ID
		userSel.Name = user.Name
		userSel.Mobile = user.Mobile
		userSels = append(userSels, userSel)
	}
	return ctx.Response().Success().Json(userSels)
}

func (r *UserController) Info(ctx http.Context) http.Response {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{
			"message": err.Error(),
		})
	}
	return ctx.Response().Success().Json(&user)
}

// Logout
func (r *UserController) Logout(ctx http.Context) http.Response {
	facades.Auth(ctx).Logout()
	return ctx.Response().Success().Json(http.Json{
		"message": "退出成功",
	})
}
