package models

import (
	"github.com/goravel/framework/database/orm"
)

type ProcAddSign struct {
	orm.Model
	EntryID     uint   `gorm:"column:entry_id;not null" json:"entry_id"`
	ProcID      uint   `gorm:"column:proc_id;not null" json:"proc_id"`
	SignType    string `gorm:"column:sign_type;not null;default:'before';comment:'前加签/后加签'" json:"sign_type"`
	SignEmpID   int    `gorm:"column:sign_emp_id;not null" json:"sign_emp_id"`
	SignEmpName string `gorm:"column:sign_emp_name;not null;default:''" json:"sign_emp_name"`
	Status      int    `gorm:"column:status;not null;default:0;comment:'0待处理 1已完成'" json:"status"`
}
