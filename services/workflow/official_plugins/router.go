package official_plugins

import "github.com/goravel/framework/contracts/foundation"

// 插件路由方法，为节点绑定执行插件
func (c *DistributePlugin) RouteApi(app foundation.Application) {
	router := app.MakeRoute()
	distributeCtrl := NewDeptController()

	//1、命令行新建一个插件
	//2、开发者通过设计，设计出该插件的一些选项和规则
	router.Post("api/plugin/product", distributeCtrl.Product)
	//3、为流程选择某些插件
	router.Post("api/flow/choose_plugins", distributeCtrl.ChoosePlugins)
	//4、获取系统中已有的插件
	router.Get("api/plugin/list", distributeCtrl.List)
}
