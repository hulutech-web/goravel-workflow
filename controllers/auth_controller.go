package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	"goravel/app/http/controllers/common"
	"goravel/app/models"
)

type AuthController struct {
	*common.WechatService
}

func NewAuthController() *AuthController {
	return &AuthController{
		//Inject services
	}
}

func (r *AuthController) AdminLogin(ctx http.Context) http.Response {
	var user models.User
	ctx.Request().Bind(&user)
	password := user.Password
	//验证
	validator, _ := facades.Validation().Make(map[string]any{
		"mobile":   ctx.Request().Input("mobile", ""),
		"password": ctx.Request().Input("password", ""),
	}, map[string]string{
		"mobile":   "required",
		"password": "required",
	}, validation.Messages(map[string]string{
		"mobile.required":   "名称不能为空",
		"password.required": "密码不能为空",
	}))
	if validator.Fails() {
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{
			"errors": validator.Errors().All(),
		})
	}
	//手机号密码验证
	facades.Orm().Query().Model(&user).Where("mobile", user.Mobile).First(&user)

	if user.ID == 0 {
		ctx.Request().AbortWithStatusJson(401, http.Json{
			"message": "error",
			"fail":    "用户不存在,请点击注册",
		})
		return nil
	}
	var user_exist models.User
	facades.Orm().Query().Model(&user_exist).Where("mobile=?", user.Mobile).First(&user_exist)
	//解密
	//if user_exist.ID != 1 {
	//	return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
	//		"message": "无权登录",
	//	})
	//}

	if !facades.Hash().Check(password, user_exist.Password) {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"message": "密码错误",
		})
	} else {
		//	生成token
		token, err1 := facades.Auth(ctx).Login(&user_exist)
		if err1 != nil {
			return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
				"message": "token生成失败",
			})
		}

		return ctx.Response().Status(http.StatusOK).Json(http.Json{
			"message": "登录成功",
			"data": struct {
				Token string      `json:"token"`
				User  models.User `json:"user"`
			}{
				Token: token,
				User:  user_exist,
			},
		})
	}
}

func (r *AuthController) H5Login(ctx http.Context) http.Response {
	var user models.User
	ctx.Request().Bind(&user)
	password := user.Password
	//验证
	validator, _ := facades.Validation().Make(map[string]any{
		"mobile":   ctx.Request().Input("mobile", ""),
		"password": ctx.Request().Input("password", ""),
	}, map[string]string{
		"mobile":   "required",
		"password": "required",
	}, validation.Messages(map[string]string{
		"mobile.required":   "名称不能为空",
		"password.required": "密码不能为空",
	}))
	if validator.Fails() {
		return ctx.Response().Json(http.StatusUnprocessableEntity, http.Json{
			"errors": validator.Errors().All(),
		})
	}
	//手机号密码验证
	facades.Orm().Query().Model(&user).Where("mobile", user.Mobile).First(&user)

	if user.ID == 0 {
		ctx.Request().AbortWithStatusJson(401, http.Json{
			"message": "error",
			"fail":    "用户不存在,请点击注册",
		})
		return nil
	}
	var user_exist models.User
	facades.Orm().Query().Model(&user_exist).Where("mobile", user.Mobile).First(&user_exist)
	//解密
	if user_exist.ID != 1 {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"message": "无权登录",
		})
	}

	if !facades.Hash().Check(password, user_exist.Password) {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"message": "密码错误",
		})
	} else {
		//	生成token
		token, err1 := facades.Auth(ctx).Login(&user_exist)
		if err1 != nil {
			return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
				"message": "token生成失败",
			})
		}

		return ctx.Response().Status(http.StatusOK).Json(http.Json{
			"message": "登录成功",
			"data": struct {
				Token string      `json:"token"`
				User  models.User `json:"user"`
			}{
				Token: token,
				User:  user_exist,
			},
		})
	}
}

// 登录
func (r *AuthController) Login(ctx http.Context) http.Response {
	var user models.User
	mobile := ctx.Request().Input("mobile", "")
	openid := ctx.Request().Input("openid", "")
	unionid := ctx.Request().Input("unionid", "")
	facades.Log().Info("mobile", mobile)
	facades.Log().Info("openid", openid)
	facades.Log().Info("unionid", unionid)
	if mobile == "" {
		ctx.Request().AbortWithStatusJson(401, http.Json{
			"error": "手机号不能为空",
		})
		return nil
	}
	facades.Orm().Query().Model(&models.User{}).Where("mobile=?", mobile).First(&user)
	if user.ID == 0 {
		ctx.Request().AbortWithStatusJson(500, http.Json{
			"error": "用户不存在,请点击注册",
		})
		return nil
	} else {
		if token, err2 := facades.Auth(ctx).Login(&user); err2 != nil {
			return ctx.Response().Json(http.StatusInternalServerError, http.Json{
				"error": "用户授权失败",
			})

		} else {
			return ctx.Response().Success().Json(http.Json{
				"data": map[string]interface{}{
					"token": token,
					"user":  user,
				},
				"message": "登录成功",
			})
		}
	}
}

func (r *AuthController) Openid(ctx http.Context) http.Response {

	code := ctx.Request().Input("code")
	openid, unionid, err := r.GetOpenidByCode(code)
	if err != nil {
		ctx.Request().AbortWithStatusJson(500, http.Json{
			"message": "获取openid失败" + err.Error(),
		})
		return nil
	}
	return ctx.Response().Success().Json(http.Json{
		"openid":  openid,
		"unionid": unionid,
	})
}

// Phone 获取手机号
func (r *AuthController) Phone(ctx http.Context) http.Response {
	code := ctx.Request().Input("code")
	phone, err := r.GetPhoneNumberByCode(code)
	if err != nil {
		ctx.Request().AbortWithStatusJson(500, http.Json{
			"msg": "获取手机号失败" + err.Error(),
		})
		return nil
	}
	return ctx.Response().Success().Json(http.Json{
		"phone": phone,
	})
}

// Logout 退出登录
func (r *AuthController) Logout(ctx http.Context) http.Response {
	return nil
}

func (r *AuthController) Regist(ctx http.Context) http.Response {
	var user models.User
	ctx.Request().Bind(&user)
	if user.Mobile == "" {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"error": "手机号不能为空",
		})
	}
	facades.Orm().Query().Model(&user).Where("mobile", user.Mobile).First(&user)
	if user.ID > 0 {
		return ctx.Response().Status(http.StatusInternalServerError).Json(http.Json{
			"error": "手机号已存在",
		})
	}
	facades.Orm().Query().Create(&user)
	//直接登录
	token, _ := facades.Auth(ctx).Login(&user)

	return ctx.Response().Success().Json(http.Json{
		"message": "注册成功",
		"data": map[string]interface{}{
			"token": token,
			"user":  user,
		},
	})
}
