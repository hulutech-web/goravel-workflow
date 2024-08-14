package controllers

import (
	"fmt"
	"github.com/goravel/framework/contracts/http"
	httpfacades "github.com/hulutech-web/http_result"
	"goravel/app/services/captcha"
)

type CaptchaController struct {
	//Dependent services
	*captcha.CaptchaService
}

func NewCaptchaController() *CaptchaController {
	return &CaptchaController{
		//Inject servicesGetCaptcha
	}
}

func (r *CaptchaController) GetCaptcha(ctx http.Context) http.Response {
	captcha_key, code, image_base64, thumb_base64, err := r.Generate()
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "error", err)
	}
	return httpfacades.NewResult(ctx).Success("", http.Json{
		"captcha_key":  captcha_key,
		"code":         code,
		"image_base64": image_base64,
		"thumb_base64": thumb_base64,
	})
}

func (r *CaptchaController) ValidateCaptcha(ctx http.Context) http.Response {
	angle := ctx.Request().InputInt64("angle")
	captcha_key := ctx.Request().Input("captcha_key", "captcha_key")
	code, isOk := r.CheckAngle(fmt.Sprintf("%d", angle), captcha_key)
	return httpfacades.NewResult(ctx).Success("", http.Json{"is_ok": isOk, "code": code})
}
