<p align="center">
  <img src="https://github.com/hulutech-web/goravel-workflow/blob/master/assets/workflow.png?raw=true" width="300" />
</p>

### 演示

<p align="center">
  <img src="https://github.com/hulutech-web/goravel-workflow/blob/master/assets/flow_preview.png?raw=true" width="800" />
</p>

### 在线演示
[地址](http://workflow.xiaohongpao.top/#/auth/login)

### 文档

使用手册请访问[文档](https://hulutech-web.github.io/goravel-workflow.github.io/)


### 一、安装
```shell
go get  github.com/hulutech-web/goravel-workflow
```
#### 1.1 注册服务提供者:config/app.go
```shell

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
```shell
go run . artisan vendor:publish --package=github.com/hulutech-web/goravel-workflow
```

可选标签：
- `--tag=migrations` — 只发布迁移文件
- `--tag=seeders` — 只发布种子数据
- `--tag=config` — 只发布配置文件
- 不指定标签则全部发布
#### 2.2 发布迁移文件:database

```shell
artisan vendor:publish --package=github.com/hulutech-web/goravel-workflow --tag=migrations
```

发布后得到两个 Go 迁移文件：
- `2024_06_24_000000_create_workflow_base_tables.go` — 创建工作流专用表（flows、flowtypes、processes、processvars、templates、templateforms、entries、entrydatas、procs、attachments、products，共 11 张表）。注：users、depts、emps 表由宿主应用提供，不包含在此迁移中。
- `2026_07_23_000000_add_workflow_features.go` — 添加新增功能字段（concurrency_type、approver_rule、unpass_target_process_id、cc_emp_ids）及新建表（proc_add_signs、proc_comments、cc_records）

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
```shell
go run . artisan migrate --seed
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
### 四、框架路由说明
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
		//撤回流程
		router.Post("entry/revoke", entryCtrl.Revoke)
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
		router.Post("process/list", processCtrl.List)

		//	审批流转
		procCtrl := controllers.NewProcController()
		router.Get("proc/{entry_id}", procCtrl.Index)
		//同意
		router.Post("pass", procCtrl.Pass)
		//驳回
		router.Post("unpass", procCtrl.UnPass)
		//撤回
		router.Post("revoke", procCtrl.Revoke)
		//加签
		router.Post("addsign", procCtrl.AddSign)
		//转交
		router.Post("transfer", procCtrl.TransferProc)
		//评论
		router.Post("comment", procCtrl.AddComment)
		router.Get("comments/{entry_id}", procCtrl.GetComments)

		//抄送
		ccCtrl := controllers.NewCcController()
		router.Get("cc/list", ccCtrl.List)
		router.Get("cc/entry/{entry_id}", ccCtrl.GetEntryCC)
	})
}

```

### 五、新增功能说明

| 功能 | 说明 | API |
|------|------|-----|
| **撤回** | 发起人撤回未处理的流程 | `POST /api/entry/revoke` / `POST /api/revoke` |
| **加签** | 当前审批人可添加额外审批人 | `POST /api/addsign` |
| **转交** | 将审批任务转交给其他员工 | `POST /api/transfer` |
| **评论** | 多轮对话式评论/留言 | `POST /api/comment` / `GET /api/comments/{entry_id}` |
| **抄送** | 节点完成后自动抄送相关人员 | `GET /api/cc/list` / `GET /api/cc/entry/{entry_id}` |
| **超时检查** | 定时检查超时的待办任务 | `go run . artisan workflow:timeout-check` |
| **会签** | 同一步骤所有审批人都需通过，全部通过才进入下一步 | Flowlink.ConcurrencyType = 1 |
| **或签** | 同一步骤一人通过即进入下一步，其余审批人自动跳过 | Flowlink.ConcurrencyType = 2 |
| **驳回到指定节点** | 驳回到任意历史节点，而非仅上一步 | ProcController.UnPass 支持 targetProcessID 参数 |
| **表单指定审批人** | 从表单字段动态读取审批人 ID | Flowlink.Auditor = -1003, ApproverRule 存储字段名 |
| **动态表达式审批人** | 通过表达式规则计算审批人 | Flowlink.Auditor = -1004, ApproverRule 存储映射键 |

#### 状态码说明
- `entries.status`: -2 = 已撤销
- `procs.status`: 3 = 已转交, 4 = 已跳过（或签）, 9 = 会签通过

#### 并发类型（Flowlink.ConcurrencyType）
| 值 | 含义 |
|----|------|
| 0 | 依次审批（默认） |
| 1 | 会签：所有审批人都需通过 |
| 2 | 或签：一人通过其余跳过 |

#### 审批人特殊值（Flowlink.Auditor）
| 值 | 含义 |
|----|------|
| -1000 | 发起人自己 |
| -1001 | 部门主管 |
| -1002 | 部门经理 |
| -1003 | 从表单字段读取审批人（ApproverRule 存储字段名） |
| -1004 | 动态表达式计算审批人（ApproverRule 存储映射键） |

### 六、前端集成
请访问前端框架[goravel-workflow-vue](https://github.com/hulutech-web/goravel-workflow-vue)下载安装扩展，并按照文档进行集成

### 七、接口文档
请访问前端框架[goravel-workflow-doc](https://github.com/hulutech-web/goravel-workflow-vuepress)进行查看
