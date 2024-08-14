package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
)

type UserRequest struct {
	Name   string `form:"name" json:"name"`
	Mobile string `form:"mobile" json:"mobile"`
}

func (r *UserRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *UserRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   "required",
		"mobile": "required",
	}
}

func (r *UserRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"name.required":   "用户名不能为空",
		"mobile.required": "手机号不能为空",
	}
}

func (r *UserRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"name":   "用户名",
		"mobile": "手机号",
	}
}

func (r *UserRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	return nil
}
