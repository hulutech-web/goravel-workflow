package official_plugins

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/goravel/framework/database/orm"
)

// 分配任务插件
type DistributePlugin struct {
}

func NewDistributePlugin() *DistributePlugin {
	return &DistributePlugin{}
}

func (c *DistributePlugin) Register() string {
	fmt.Println("register distribute plugin called")
	// 分配数据插件
	return "distribute"
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

type DistributePluginConfig struct {
	orm.Model
	FlowID    uint `gorm:"column:flow_id" json:"flow_id" form:"flow_id"`
	ProcessID uint `gorm:"column:process_id" json:"process_id" form:"process_id"`
	Rules     Rule `gorm:"column:config" json:"config" form:"config"`
}

func (c *DistributePlugin) AutoMigrate() error {
	db := BootMS()
	return db.AutoMigrate(&DistributePluginConfig{})
}
