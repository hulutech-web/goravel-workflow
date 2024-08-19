![goravel-workflow](https://github.com/hulutech-web/goravel-workflow/blob/master/assets/workflow.png?raw=true #pic_center =300x300)
### 文档


[文档](https://hulutech-web.github.io/goravel-workflow.github.io/)


### 一、安装
```go
go get  github.com/hulutech-web/goravel-workflow
```
#### 1.1 注册服务提供者:config/app.go
```go
import	"github.com/hulutech-web/goravel-workflow"
```

#### 1.2 注册服务提供者:config/app.go
```go
func init() {
"providers": []foundation.ServiceProvider{
	....
	&workflow.ServiceProvider{},
}
}
```
### 二、发布资源，默认将发布2类资源，一是配置文件，而是数据表迁移
#### 2.1 发布资源:config/app.go
```go
go run . artisan vendor:publish --package=github.com/hulutech-web/goravel-workflow

```
#### 2.2 发布迁移文件:database
```go
artisan vendor:publish --package=github.com/hulutech-web/goravel-workflow
```
#### 2.3 执行迁移建表
在database/seeders/database_seeder.go下的添加
```go
func (s *DatabaseSeeder) Run() error {
	return facades.Seeder().Call([]seeder.Seeder{
		&WorkflowDatabaseSeeder{},
	})
}

```
#### 2.4 执行迁移
```go
go run . artisan migrate:refresh --seed
```

#### 2.5 检查路由重名
如果启动项目报错，请检查路由是否有重名，并修改路由
#### 2.6 模型映射
发布资源后，config/workflow.go中的配置文件中有默认的关联映射，根据需要自行修改和修改
### 三、实现Hook接口（可选）
用户自定义User结构中注入流程框架，并实现框架中的Hook接口
```go
type User struct {
	orm.Model
	Name     string `gorm:"column:name;type:varchar(255);not null" form:"name" json:"name"`
	WorkNo   string `gorm:"column:workno;not null;unique_index:users_workno_unique" json:"workno" form:"workno"`
	Password string `gorm:"column:password;type:varchar(255);not null" form:"password" json:"password"`
	...
	Workflow *Workflow
	orm.SoftDeletes
}
```
实现接口
```go
// 通知发起人，在被驳回时调用，或者整个流程结束时调用。
func (u *User) NotifySendOne(id uint) error {

	fmt.Printf("custom ======User %d unpasshook called.\n", id)
	return nil
}

// 通知下一个审批人，当当前环节的审批人通过时，触发。
func (u *User) NotifyNextAuditor(id uint) error {
	fmt.Printf("custom ======User %d passhook called.\n", id)
	return nil
}

```

### 实例化workflow
框架提供了2个``hooks``，供开发者自行实现逻辑，可以发送邮件通知，短信通知等
``app/providers/app_services_provider.go``
实例化workflow，并注入服务
```go
func (receiver *AppServiceProvider) Boot(app foundation.Application) {
	wf := workflow.NewBaseWorkflow()
	// 注册子级的方法到工作流中
	user := &models.User{Workflow: wf}
	wf.RegisterHook("NotifySendOneHook", reflect.ValueOf(user.NotifySendOne))
	wf.RegisterHook("NotifyNextAuditorHook", reflect.ValueOf(user.NotifyNextAuditor))
}

回调参数将在User结构中的NotifySendOne和NotifyNextAuditor方法中执行后续操作，由开发者自行实现

```
### 二、框架路由说明
```go
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
		//模板控件
		templateformCtrl := controllers.NewTemplateformController()
		router.Get("template/{id}/templateform", templateformCtrl.Index)
		router.Post("templateform", templateformCtrl.Store)
		router.Put("templateform/{id}", templateformCtrl.Update)
		router.Delete("templateform/{id}", templateformCtrl.Destroy)
		router.Get("templateform/{id}", templateformCtrl.Show)
		//模板
		templateCtrl := controllers.NewTemplateController()
		router.Resource("template", templateCtrl)

		//	流程
		processCtrl := controllers.NewProcessController()
		router.Resource("process", processCtrl)
		router.Get("process/attribute", processCtrl.Attribute)
		router.Post("process/con", processCtrl.Condition)

		//	审批流转
		procCtrl := controllers.NewProcController()
		router.Get("proc/{entry_id}", procCtrl.Index)
		//同意
		router.Post("pass", procCtrl.Pass)
		//驳回
		router.Post("unpass", procCtrl.UnPass)
	})
}

```

### 三、前端集成
请访问前端框架[goravel-workflow-vue](https://github.com/hulutech-web/goravel-workflow-vue)下载安装扩展，并按照文档进行集成

### 四、接口文档
请访问前端框架[goravel-workflow-doc](https://github.com/hulutech-web/goravel-workflow-vuepress)进行查看
