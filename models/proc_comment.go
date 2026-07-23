package models

import (
	"github.com/goravel/framework/database/orm"
)

type ProcComment struct {
	orm.Model
	EntryID uint   `gorm:"column:entry_id;not null" json:"entry_id"`
	ProcID  uint   `gorm:"column:proc_id;not null" json:"proc_id"`
	EmpID   int    `gorm:"column:emp_id;not null;comment:'发言人ID'" json:"emp_id"`
	EmpName string `gorm:"column:emp_name;not null;default:'';comment:'发言人名称'" json:"emp_name"`
	Content string `gorm:"column:content;type:text;not null;comment:'评论内容'" json:"content"`
	Status  int    `gorm:"column:status;not null;default:1;comment:'1正常 2删除'" json:"status"`
}
