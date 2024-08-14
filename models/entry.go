package models

import (
	"github.com/goravel/framework/database/orm"
)

type Entry struct {
	orm.Model
	Title          string      `gorm:"column:title;not null;default:''" json:"title" form:"title"`
	FlowID         uint        `gorm:"column:flow_id;not null;default:0" json:"flow_id" form:"flow_id"`
	EmpID          uint        `gorm:"column:emp_id;not null;default:0" json:"emp_id" form:"emp_id"`
	ProcessID      uint        `gorm:"column:process_id;not null;default:0" json:"process_id" form:"process_id"`
	Circle         int         `gorm:"column:circle;not null;default:1" json:"circle" form:"circle"`
	Status         int         `gorm:"column:status;not_null" json:"status" form:"status"`
	Pid            int         `gorm:"column:pid;not null;default:0" json:"pid" form:"pid"`
	EnterProcessID int         `gorm:"column:enter_process_id;not null;default:0" json:"enter_process_id" form:"enter_process_id"`
	EnterProcID    int         `gorm:"column:enter_proc_id;not null;default:0" json:"enter_proc_id" form:"enter_proc_id"`
	Child          int         `gorm:"column:child;not null;default:0" json:"child" form:"child"`
	Flow           Flow        `gorm:"foreignKey:flow_id"` // 关联的Flow
	Emp            Emp         `gorm:"foreignKey:emp_id"`  // 关联的Emp
	Procs          []*Proc     // HasMany Proc
	Process        Process     `gorm:"foreignKey:process_id"` // 关联的Process
	EntryDatas     []EntryData // HasMany EntryData
	ParentEntry    *Entry      `gorm:"foreignKey:pid"`              // 关联的父Entry
	Children       []Entry     `gorm:"foreignKey:pid"`              // HasMany Entry, 级联删除
	EnterProcess   Process     `gorm:"foreignKey:enter_process_id"` // 关联的进入步骤Process
}
