package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/hulutech-web/goravel-workflow/services/workflow/official_plugins"
	"gorm.io/gorm"
)

type Flow struct {
	orm.Model
	FlowNo      string                     `gorm:"column:flow_no;not null" json:"flow_no" form:"flow_no"`
	FlowName    string                     `gorm:"column:flow_name;not null;default:''" json:"flow_name" form:"flow_name"`
	TemplateID  int                        `gorm:"column:template_id;not null;default:0" json:"template_id" form:"template_id"`
	Flowchart   string                     `gorm:"column:flowchart" json:"flowchart" form:"flowchart"`
	Jsplumb     string                     `gorm:"column:jsplumb;comment:'jsplumb流程图数据'" json:"jsplumb" form:"jsplumb"`
	TypeID      int                        `gorm:"column:type_id;not null;default:0" json:"type_id" form:"type_id"`
	IsPublish   bool                       `gorm:"column:is_publish;not null;default:0" json:"is_publish" form:"is_publish"`
	IsShow      bool                       `gorm:"column:is_show;not null;default:1" json:"is_show" form:"is_show"`
	Processes   []Process                  `gorm:"foreignKey:FlowID"`     // HasMany Process
	ProcessVars []ProcessVar               `gorm:"foreignKey:FlowID"`     // HasMany ProcessVar
	Template    Template                   `gorm:"foreignKey:TemplateID"` // BelongsTo Template
	Flowtype    Flowtype                   `gorm:"foreignKey:TypeID"`     // BelongsTo FlowType
	Plugins     []*official_plugins.Plugin `gorm:"many2many:flow_plugins"`
}

// LoadProcesses preloads the associated Processes for a Flow.
func (f *Flow) LoadProcesses(db *gorm.DB) error {
	return db.Preload("Processes").Find(f).Error
}

// LoadProcessVars preloads the associated ProcessVars for a Flow.
func (f *Flow) LoadProcessVars(db *gorm.DB) error {
	return db.Preload("ProcessVars").Find(f).Error
}

// LoadTemplate preloads the associated Template for a Flow.
func (f *Flow) LoadTemplate(db *gorm.DB) error {
	return db.Preload("Template").Find(f).Error
}

// LoadFlowType preloads the associated FlowType for a Flow.
func (f *Flow) LoadFlowType(db *gorm.DB) error {
	return db.Preload("FlowType").Find(f).Error
}
