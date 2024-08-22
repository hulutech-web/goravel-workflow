package official_plugins

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/goravel/framework/database/orm"
	"github.com/hulutech-web/goravel-workflow/models"
)

type Plugin struct {
	orm.Model
	Name          string         `gorm:"column:name;unique;comment:'插件名称'" json:"name" form:"name"`
	Version       string         `gorm:"column:version;comment:'版本号'" json:"version" form:"version"`
	Status        int            `gorm:"column:status;comment:'状态'" json:"status" form:"status"`
	Description   string         `gorm:"column:description;comment:'描述'" json:"description" form:"description"`
	Author        string         `gorm:"column:author;comment:'作者'" json:"author" form:"author"`
	IsDesigned    int            `gorm:"column:is_designed;comment:'是否完成设计'" json:"is_designed" form:"is_designed"`
	PluginConfigs []PluginConfig `gorm:"foreignKey:PluginID;references:ID"`
	Flows         []*models.Flow `gorm:"many2many:flow_plugins"`
}

// flow_plugin中间表
type FlowPlugin struct {
	orm.Model
	PluginID uint `gorm:"column:plugin_id;comment:'插件ID'" json:"plugin_id" form:"plugin_id"`
	FlowID   uint `gorm:"column:flow_id;comment:'流程ID'" json:"flow_id" form:"flow_id"`
}

// 为某一个流程中的某一个步骤添加规则
type PluginConfig struct {
	orm.Model
	PluginID  uint `gorm:"column:plugin_id;comment:'插件ID'" json:"plugin_id" form:"plugin_id"`
	FlowID    uint `gorm:"column:flow_id" json:"flow_id" form:"flow_id"`
	ProcessID uint `gorm:"column:process_id" json:"process_id" form:"process_id"`
	FieldID   int  `gorm:"column:field_id;comment:'对应template_form中的字段field对应的id'" json:"field_id" form:"field_id"`
	Rules     Rule `gorm:"column:rules" json:"config" form:"rules"`
}

type RuleItem struct {
	RuleID    int    `json:"rule_id" form:"rule_id"`       //部门id
	RuleLabel string `json:"rule_label" form:"rule_label"` //部门名称
	RuleValue int    `json:"rule_value" form:"rule_value"` //部门值
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
