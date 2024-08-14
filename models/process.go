package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

type Process struct {
	orm.Model
	FlowID           int       `gorm:"column:flow_id;not null;default:0;comment:'流程id'" json:"flow_id"`
	ProcessName      string    `gorm:"column:process_name;not null;default:'';comment:'步骤名称'" json:"process_name"`
	LimitTime        int       `gorm:"column:limit_time;not null;default:0;comment:'限定时间,单位秒'" json:"limit_time"`
	Type             string    `gorm:"column:type;not null;default:'operation';comment:'流程图显示操作框类型'" json:"type"`
	Icon             string    `gorm:"column:icon;default:'';comment:'流程图显示图标'" json:"icon,omitempty"`
	ProcessTo        string    `gorm:"column:process_to;not null;default:''" json:"process_to"`
	Style            string    `gorm:"column:style;type:text;" json:"style,omitempty"`
	StyleColor       string    `gorm:"column:style_color;not null;default:'#78a300'" json:"style_color"`
	StyleHeight      int       `gorm:"column:style_height;not null;default:30" json:"style_height"`
	StyleWidth       int       `gorm:"column:style_width;not null;default:30" json:"style_width"`
	PositionLeft     string    `gorm:"column:position_left;not null;default:'100px'" json:"position_left"`
	PositionTop      string    `gorm:"column:position_top;not null;default:'200px'" json:"position_top"`
	Position         int       `gorm:"column:position;not null;default:1;comment:'步骤位置：1正常步骤2：转入子流程0：第一步 当为2时 child_flow_id child_after child_back_process 可设置'" json:"position"`
	ChildFlowID      int       `gorm:"column:child_flow_id;not null;default:0;comment:'子流程id'" json:"child_flow_id"`
	ChildAfter       int       `gorm:"column:child_after;not null;default:2;comment:'子流程结束后 1.同时结束父流程 2.返回父流程'" json:"child_after"`
	ChildBackProcess int       `gorm:"column:child_back_process;not null;default:0;comment:'子流程结束后返回父流程进程'" json:"child_back_process"`
	Description      string    `gorm:"column:description;not null;default:'';comment:'步骤描述'" json:"description"`
	ProcessVars      []Process `gorm:"many2many:process_vars;" json:"process_vars"`
	Flow             Flow
}

// LoadEmp preloads the associated Emp for a Proc.
func (p *Process) LoadFlow(db *gorm.DB) error {
	return db.Preload("Flow").First(p).Error
}

func (p *Process) LoadProcessVars(db *gorm.DB) error {
	return db.Preload("Processvars").First(p).Error
}
