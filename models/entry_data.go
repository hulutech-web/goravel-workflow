package models

import (
	"github.com/goravel/framework/database/orm"
)

type EntryData struct {
	orm.Model
	EntryID     int    `gorm:"column:entry_id;not null;default:0" form:"entry_id" json:"entry_id"`
	FlowID      int    `gorm:"column:flow_id;not null;default:0" form:"flow_id" json:"flow_id"`
	FieldName   string `gorm:"column:field_name;not null;default:''" form:"field_name" json:"field_name"`
	FieldValue  string `gorm:"column:field_value" json:"field_value" form:"field_value" json:"field_value"`
	FieldRemark string `gorm:"column:field_remark;not null;default:''" form:"field_remark" json:"field_remark"`
}

func (e *EntryData) TableName() string {
	return "entrydatas"
}
