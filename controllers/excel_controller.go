package controllers

import (
	"github.com/goravel/framework/contracts/http"
	httpfacades "github.com/hulutech-web/http_result"
	"goravel/app/services/Excel"
)

type ExcelController struct {
	//Dependent services
	excelService *Excel.ExcelService
}

func NewExcelController() *ExcelController {
	return &ExcelController{
		excelService: Excel.NewExcelService(),
		//Inject services
	}
}

// 使用他
func (r *ExcelController) Pdf(ctx http.Context) http.Response {
	//err := r.excelService.GetTemplateFile()
	err := r.excelService.GenPdf()
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, err.Error(), nil)
	}
	return httpfacades.NewResult(ctx).Success("转换成功", nil)
}
