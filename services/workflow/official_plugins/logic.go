package official_plugins

import (
	"errors"
	"fmt"
	"github.com/goravel/framework/facades"
	"sync"
)

var (
	Once sync.Once
)

// 分配任务插件
type DistributePlugin struct {
	HookName string
}

func NewDistributePlugin() *DistributePlugin {
	return &DistributePlugin{}
}

func (c *DistributePlugin) Register() string {
	fmt.Println("register distribute_plugin called")
	// 分配数据插件
	app := facades.App()
	c.RouteApi(app)
	return "distribute_plugin"
}

func (c *DistributePlugin) Action() func(string) error {
	fmt.Println("distribute plugin action called")
	return func(task string) error {
		c.AddHook(task)
		return c.AutoMigrate()
	}
}

func (c *DistributePlugin) AddHook(hook string) {
	c.HookName = hook
}

func (c *DistributePlugin) AutoMigrate() error {
	err_ := errors.New("")
	Once.Do(func() {
		orm := BootMS()

		if !orm.Migrator().HasTable(&Plugin{}) {
			err_ = orm.AutoMigrate(&Plugin{})
		}
		if !orm.Migrator().HasTable(&PluginConfig{}) {
			err_ = orm.AutoMigrate(&PluginConfig{})
		}
		if !orm.Migrator().HasTable(&FlowPlugin{}) {
			err_ = orm.AutoMigrate(&FlowPlugin{})
		}

		row := orm.FirstOrCreate(&Plugin{
			Name:    "数据二次分配",
			Version: "v1.0",
			Status:  1,
			Description: "1、设计流程时，将某一个数字类型的字段绑定插件，" +
				"2、节点设计规则，在某一个节点”如：主管审批“，规则为：员工1：500元，员工2：1000元，绑定该字段进行二次分配，下一个节点的审批人获取到该规则，" +
				"3、数据查看，下一节点审批人，根据规则获取到自身可以查看的规则内容，员工1：看到奖励500元，员工2：看到奖励1000元",
			Author: "hulu-web",
		})
		if row.RowsAffected == 0 || row.Error != nil {
			err_ = row.Error
		} else {
			fmt.Println("AutoMigrate successful")
		}
	})
	return err_
}

// 插件执行方法，当流程执行到某一个流程的某一个节点，会自动调用该执行方法，将数据交给下一级
func (c *DistributePlugin) Execute(plugin_name string, args ...interface{}) error {
	//当当前节点执行时，先查询该flowID和processID中是否存在数据，如果存在，则将flowID对应的entry_data中的
	//扩展字段找出，并应用执行方案
	if plugin_name != c.HookName {
		return nil
	} else {
		//	从args中获取flow_id,process_id
		flow_id := args[0].(uint)
		process_id := args[1].(uint)
		fmt.Println("distribute_plugin execute"+",flow_id:", flow_id, ",process_id:", process_id)
	}
	return nil
}
