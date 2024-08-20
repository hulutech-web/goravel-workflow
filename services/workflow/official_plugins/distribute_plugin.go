package official_plugins

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/database/orm"
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
	Name          string         `gorm:"column:name;unique;comment:'插件名称'" json:"name" form:"name"`
	Version       string         `gorm:"column:version;comment:'版本号'" json:"version" form:"version"`
	Status        int            `gorm:"column:status;comment:'状态'" json:"status" form:"status"`
	Description   string         `gorm:"column:description;comment:'描述'" json:"description" form:"description"`
	Author        string         `gorm:"column:author;comment:'作者'" json:"author" form:"author"`
	PluginConfigs []PluginConfig `gorm:"foreignKey:PluginID;references:ID"`
}

// flow_plugin中间表
type FlowPlugin struct {
	orm.Model
	PluginID uint `gorm:"column:plugin_id;comment:'插件ID'" json:"plugin_id" form:"plugin_id"`
	FlowID   uint `gorm:"column:flow_id;comment:'流程ID'" json:"flow_id" form:"flow_id"`
}

func (c *DistributePlugin) AutoMigrate() error {
	err_ := errors.New("distribute_plugin")
	Once.Do(func() {
		orm := BootMS()
		err := errors.New("distribute_plugin")
		if !orm.Migrator().HasTable(&Plugin{}) {
			err = orm.AutoMigrate(&Plugin{})
		}
		if !orm.Migrator().HasTable(&PluginConfig{}) {
			err = orm.AutoMigrate(&PluginConfig{})
		}
		if !orm.Migrator().HasTable(&FlowPlugin{}) {
			err = orm.AutoMigrate(&FlowPlugin{})
		}
		if err != nil {
			err_ = err
			fmt.Println("AutoMigrate error:", err)
			// 处理错误
		} else {
			fmt.Println("AutoMigrate successful")
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
			fmt.Println("Create error:", err)
		} else {
			fmt.Println("Create successful")
		}
	})
	return err_
}

// 插件路由方法，为节点绑定执行插件
func (c *DistributePlugin) RouteApi(app foundation.Application) {
	router := app.MakeRoute()
	distributeCtrl := NewDeptController()

	//1、命令行新建一个插件
	//2、开发者通过设计，设计出该插件的一些选项和规则
	router.Post("api/plugin/product", distributeCtrl.Product)
	//3、为流程选择某些插件
	router.Post("api/flow/select_plugins", distributeCtrl.SelectPlugins)
	//4、获取系统中已有的插件
	router.Get("api/plugin/list", distributeCtrl.List)
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
