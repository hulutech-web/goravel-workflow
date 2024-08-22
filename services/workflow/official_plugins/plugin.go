package official_plugins

import (
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
	PluginConfigs []PluginConfig `gorm:"foreignKey:PluginID;references:ID"`
	Flows         []*models.Flow `gorm:"many2many:flow_plugins"`
}

// flow_plugin中间表
type FlowPlugin struct {
	orm.Model
	PluginID uint `gorm:"column:plugin_id;comment:'插件ID'" json:"plugin_id" form:"plugin_id"`
	FlowID   uint `gorm:"column:flow_id;comment:'流程ID'" json:"flow_id" form:"flow_id"`
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
