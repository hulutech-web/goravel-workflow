package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/hulutech-web/goravel-workflow/models/common"
	"gorm.io/gorm"
)

type TemplateForm struct {
	orm.Model
	Field             string            `gorm:"column:field;not null;default:'';comment:'表单字段英文名'" json:"field" form:"field"`
	FieldName         string            `gorm:"column:field_name;not null;default:'';comment:'表单字段中文名'" json:"field_name" form:"field_name"`
	FieldType         string            `gorm:"column:field_type;not null;default:'';comment:'表单字段类型'" json:"field_type" form:"field_type"`
	FieldValue        common.FieldValue `gorm:"column:field_value;type:text;comment:'表单字段值，select radio checkbox用'" json:"field_value" form:"field_value"`
	FieldDefaultValue string            `gorm:"column:field_default_value;type:text;comment:'表单字段默认值'" json:"field_default_value" form:"field_default_value"`
	FieldRules        common.Rule       `gorm:"column:field_rules;" json:"field_rules" form:"field_rules"`
	Sort              int               `gorm:"column:sort;not null;default:100;comment:'排序'" json:"sort" form:"sort"`
	TemplateID        uint              `gorm:"column:template_id;not null;default:0;comment:'模板ID'" json:"template_id" form:"template_id"`
	Template          Template
}

func (p *TemplateForm) TableName() string {
	return "templateforms"
}
func (p *TemplateForm) LoadTemplateForm(db *gorm.DB) error {
	return db.Preload("Template").First(p).Error
}
