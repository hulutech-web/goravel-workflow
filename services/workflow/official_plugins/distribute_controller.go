package official_plugins

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
)

type DistributeController struct {
	//Dependent services
}

func NewDeptController() *DistributeController {
	return &DistributeController{
		//Inject services
	}
}

type DistributeRequest struct {
	FlowID    uint `json:"flow_id" form:"flow_id"`
	ProcessID uint `json:"process_id" form:"process_id"`
	Rules     Rule `json:"rules" form:"rules"`
}

func (r *DistributeController) Bind(ctx http.Context) http.Response {
	var distributeRequest DistributeRequest
	ctx.Request().Bind(&distributeRequest)
	var distriutepluginConfig DistributePluginConfig
	err := facades.Orm().Query().Model(&DistributePluginConfig{}).Create(&distriutepluginConfig)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(500, "创建失败", err)
	}
	return httpfacades.NewResult(ctx).Success("创建成功", distriutepluginConfig)
}
