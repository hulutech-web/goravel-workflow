package controllers

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/hulutech-web/goravel-workflow/services/Upload"
	httpfacades "github.com/hulutech-web/http_result"
)

type UploadController struct {
	//Dependent services
}

func NewUploadController() *UploadController {
	return &UploadController{
		//Inject services
	}
}

func (r *UploadController) Upload(ctx http.Context) http.Response {
	file, err := ctx.Request().File("file")
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "上传失败", nil)
	}
	if att, err := Upload.NewUploadService().Upload(ctx, file); err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "上传失败", nil)
	} else {
		return httpfacades.NewResult(ctx).Success("上传成功", att.Path)
	}
}
