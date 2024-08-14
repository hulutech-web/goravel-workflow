package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

type ProcessVar struct {
	orm.Model
	ProcessID       int     `gorm:"column:process_id;not null" json:"process_id"`
	FlowID          int     `gorm:"column:flow_id;not null;comment:'流程id'" json:"flow_id"`
	ExpressionField string  `gorm:"column:expression_field;not null;comment:'条件表达式字段名称'" json:"expression_field"`
	Process         Process `gorm:"foreignKey:ProcessID;references:ID"`
}

func (e *ProcessVar) TableName() string {
	return "processvars"
}

func (p *ProcessVar) LoadProcess(db *gorm.DB) error {
	return db.Preload("Process").First(p).Error
}
