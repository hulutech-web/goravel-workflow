# workflow
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

### 三、使用
#### 2.1 使用说明:自定义默认返回
发布资源后，config/workflow.go中的配置文件中有默认的关联映射，根据需要自行修改和修改