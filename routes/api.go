package routes

import (
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/route"
	"github.com/goravel/framework/facades"
	controllers "github.com/hulutech-web/goravel-workflow/controllers"
	"github.com/hulutech-web/goravel-workflow/middleware"
)

func Api(app foundation.Application) {
	router := app.MakeRoute()

	authController := controllers.NewAuthController()
	router.Post("/api/auth/login", authController.AdminLogin)
	router.Post("/api/h5/login", authController.H5Login)
	captchaController := controllers.NewCaptchaController()
	router.Get("/api/captcha/get", captchaController.GetCaptcha)
	router.Post("/api/captcha/validate", captchaController.ValidateCaptcha)

	facades.Route().Middleware(middleware.Jwt()).Prefix("/api").Group(func(router route.Router) {

		//文件上传
		uploadCtrl := controllers.NewUploadController()
		router.Post("/upload", uploadCtrl.Upload)

		homeCtrl := controllers.NewHomeController()
		router.Get("/home", homeCtrl.Index)

		//	部门
		deptCtrl := controllers.NewDeptController()
		router.Resource("dept", deptCtrl)
		router.Get("dept/list", deptCtrl.List)
		router.Post("dept/bindmanager", deptCtrl.BindManager)
		router.Post("dept/binddirector", deptCtrl.BindDirector)

		//	员工
		empCtrl := controllers.NewEmpController()
		router.Resource("emp", empCtrl)
		router.Post("emp/search", empCtrl.Search)
		router.Get("emp/options", empCtrl.Options)
		router.Post("emp/bind", empCtrl.BindUser)
		//流程
		flowCtrl := controllers.NewFlowController()
		router.Resource("flow", flowCtrl)
		router.Get("flow/list", flowCtrl.List)
		router.Get("flow/create", flowCtrl.Create)
		//流程设计
		router.Get("flow/flowchart/{id}", flowCtrl.FlowDesign)
		router.Post("flow/publish", flowCtrl.Publish)

		//entry节点
		entryCtrl := controllers.NewEntryController()
		router.Get("flow/{id}/entry", entryCtrl.Create)
		router.Post("entry", entryCtrl.Store)
		router.Get("entry/{id}", entryCtrl.Show)
		router.Get("entry/{id}/entrydata", entryCtrl.EntryData)
		//流程重发
		router.Post("entry/resend", entryCtrl.Resend)
		//流程轨迹
		flowlinkCtrl := controllers.NewFlowlinkController()
		router.Post("flowlink", flowlinkCtrl.Update)

		//模板
		templateCtrl := controllers.NewTemplateController()
		router.Resource("template", templateCtrl)

		//模板控件
		templateformCtrl := controllers.NewTemplateformController()
		router.Get("template/{id}/templateform", templateformCtrl.Index)
		router.Post("templateform", templateformCtrl.Store)
		router.Put("templateform/{id}", templateformCtrl.Update)
		router.Delete("templateform/{id}", templateformCtrl.Destroy)
		router.Get("templateform/{id}", templateformCtrl.Show)
		router.Post("flow/templateform", templateformCtrl.FlowTemplateForm)

		//	流程
		processCtrl := controllers.NewProcessController()
		router.Resource("process", processCtrl)
		router.Get("process/attribute", processCtrl.Attribute)
		router.Post("process/con", processCtrl.Condition)
		router.Post("process/list", processCtrl.List)

		//	审批流转
		procCtrl := controllers.NewProcController()
		router.Get("proc/{entry_id}", procCtrl.Index)
		//同意
		router.Post("pass", procCtrl.Pass)
		//驳回
		router.Post("unpass", procCtrl.UnPass)
	})
}
