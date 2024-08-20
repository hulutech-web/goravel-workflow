package official_plugins

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/facades"
)

// 分配任务插件
type DistributePlugin struct {
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
		return c.AutoMigrate()
	}
}

type RuleItem struct {
	RuleName  string `json:"rule_name" form:"rule_name"`
	RuleTitle string `json:"rule_title" form:"rule_title"`
	RuleValue string `json:"rule_value" form:"rule_value"`
}
type Rule []RuleItem

func (t *Rule) Scan(value interface{}) error {
	bytesValue, _ := value.([]byte)
	return json.Unmarshal(bytesValue, t)
}

func (t Rule) Value() (driver.Value, error) {
	//如果t为nil,返回nil
	return json.Marshal(t)
}

// 分配数量插件配置
type PluginConfig struct {
	orm.Model
	PluginID    uint `gorm:"column:plugin_id;comment:'插件ID'" json:"plugin_id" form:"plugin_id"`
	EntryID     uint `gorm:"column:dept_id;comment:'部门ID'" json:"dept_id" form:"dept_id"`
	FlowID      uint `gorm:"column:flow_id" json:"flow_id" form:"flow_id"`
	ProcessID   uint `gorm:"column:process_id" json:"process_id" form:"process_id"`
	EntrydataID uint `gorm:"column:entrydata_id;comment:'输入内容中的哪一个字段'" json:"entrydata_id" form:"entrydata_id"`
	Rules       Rule `gorm:"column:config" json:"config" form:"config"`
}

type Plugin struct {
	orm.Model
	Name        string `gorm:"column:name;unique;comment:'插件名称'" json:"name" form:"name"`
	Version     string `gorm:"column:version;comment:'版本号'" json:"version" form:"version"`
	Status      int    `gorm:"column:status;comment:'状态'" json:"status" form:"status"`
	Description string `gorm:"column:description;comment:'描述'" json:"description" form:"description"`
	Author      string `gorm:"column:author;comment:'作者'" json:"author" form:"author"`
}

func (c *DistributePlugin) AutoMigrate() error {
	db := BootMS()
	err := db.AutoMigrate(&Plugin{}, &PluginConfig{})
	if err != nil {
		db.Create(&Plugin{
			Name:    "数据二次分配",
			Version: "v1.0",
			Status:  1,
			Description: "1、设计流程时，将某一个数字类型的字段绑定插件，" +
				"2、节点设计规则，在某一个节点”如：主管审批“，规则为：员工1：500元，员工2：1000元，绑定该字段进行二次分配，下一个节点的审批人获取到该规则，" +
				"3、数据查看，下一节点审批人，根据规则获取到自身可以查看的规则内容，员工1：看到奖励500元，员工2：看到奖励1000元",
			Author: "hulu-web",
		})
	}
	return err
}

// 插件路由方法，为节点绑定执行插件
func (c *DistributePlugin) RouteApi(app foundation.Application) {
	router := app.MakeRoute()
	distributeCtrl := NewDeptController()

	//开发者安装插件
	router.Post("api/plugin/install", distributeCtrl.Install)
	//开发者提交插件信息，产出插件
	router.Post("api/plugin/product", distributeCtrl.Product)
	//流程绑定插件
	router.Post("api/flow/bind_plugin", distributeCtrl.Bind)
}

// 插件执行方法
func (c *DistributePlugin) Execute(flowID uint, processID uint) error {
	//当当前节点执行时，先查询该flowID和processID中是否存在数据，如果存在，则将flowID对应的entry_data中的
	//扩展字段找出，并应用执行方案
	return nil
}
