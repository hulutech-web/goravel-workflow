package models

import (
	"github.com/goravel/framework/database/orm"
)

type CcRecord struct {
	orm.Model
	EntryID   uint `gorm:"column:entry_id;not null" json:"entry_id"`
	FlowID    uint `gorm:"column:flow_id;not null" json:"flow_id"`
	ProcessID uint `gorm:"column:process_id;not null" json:"process_id"`
	ProcID    uint `gorm:"column:proc_id;not null" json:"proc_id"`
	EmpID     int  `gorm:"column:emp_id;not null;comment:'抄送人ID'" json:"emp_id"`
	EmpName   string `gorm:"column:emp_name;not null;default:'';comment:'抄送人名称'" json:"emp_name"`
	Status    int  `gorm:"column:status;not null;default:0;comment:'0未读 1已读'" json:"status"`
}
