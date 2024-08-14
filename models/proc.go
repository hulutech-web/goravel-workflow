package models

import (
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/support/carbon"
	"gorm.io/gorm"
)

type Proc struct {
	orm.Model
	EntryID     uint            `gorm:"column:entry_id;not null" json:"entry_id" form:"entry_id"`
	FlowID      int             `gorm:"column:flow_id;not null;comment:'流程id'" json:"flow_id" form:"flow_id"`
	ProcessID   int             `gorm:"column:process_id;not null;comment:'当前步骤'" json:"process_id" form:"process_id"`
	ProcessName string          `gorm:"column:process_name;not null;default:'';comment:'当前步骤名称'" json:"process_name" form:"process_name"`
	EmpID       int             `gorm:"column:emp_id;not null;comment:'审核人'" json:"emp_id" form:"emp_id"`
	EmpName     string          `gorm:"column:emp_name;default:null;comment:'审核人名称'" json:"emp_name" form:"emp_name"`
	DeptName    string          `gorm:"column:dept_name;default:null;comment:'审核人部门名称'" json:"dept_name" form:"dept_name"`
	AuditorID   int             `gorm:"column:auditor_id;not null;default:0;comment:'具体操作人'" json:"auditor_id" form:"auditor_id"`
	AuditorName string          `gorm:"column:auditor_name;not null;default:'';comment:'操作人名称'" json:"auditor_name" form:"auditor_name"`
	AuditorDept string          `gorm:"column:auditor_dept;not null;default:'';comment:'操作人部门'" json:"auditor_dept" form:"auditor_dept"`
	Status      int             `gorm:"column:status;not null;comment:'当前处理状态 0待处理 9通过 -1驳回\n0：处理中\n-1：驳回\n9：会签'" json:"status" form:"status"`
	Content     string          `gorm:"column:content;default:null;comment:'批复内容'" json:"content" form:"content"`
	IsRead      int             `gorm:"column:is_read;not null;default:0;comment:'是否查看'" json:"is_read" form:"is_read"`
	IsReal      bool            `gorm:"column:is_real;not null;default:1;comment:'审核人和操作人是否同一人'" json:"is_real" form:"is_real"`
	Circle      int             `gorm:"column:circle;not null;default:1" json:"circle" form:"circle"`
	Beizhu      string          `gorm:"column:beizhu;type:text;comment:'备注'" json:"beizhu" form:"beizhu"`
	Concurrence carbon.DateTime `gorm:"column:concurrence;not null;default:0;comment:'并行查找解决字段， 部门 角色 指定 分组用'" json:"concurrence" form:"concurrence"`
	Emp         Emp             `gorm:"foreignKey:EmpID"`                                                  // 关联的Emp
	Entry       Entry           `gorm:"foreignKey:EntryID"`                                                // 关联的Entry
	Process     Process         `gorm:"foreignKey:ProcessID"`                                              // 关联的Process
	Flow        Flow            `gorm:"foreignKey:FlowID"`                                                 // 关联的Flow
	SubProcs    []Proc          `gorm:"foreignkey:EntryID;constraint:OnUpdate:CASCADE,OnDelete:NO ACTION"` // HasMany Proc
}

// LoadEmp preloads the associated Emp for a Proc.
func (p *Proc) LoadEmp(db *gorm.DB) error {
	return db.Preload("Emp").First(p).Error
}

// LoadEntry preloads the associated Entry for a Proc.
func (p *Proc) LoadEntry(db *gorm.DB) error {
	return db.Preload("Entry").First(p).Error
}

// LoadProcess preloads the associated Process for a Proc.
func (p *Proc) LoadProcess(db *gorm.DB) error {
	return db.Preload("Process").First(p).Error
}

// LoadFlow preloads the associated Flow for a Proc.
func (p *Proc) LoadFlow(db *gorm.DB) error {
	return db.Preload("Flow").First(p).Error
}

// LoadSubProcs preloads the associated SubProcs (child processes) for a Proc.
func (p *Proc) LoadSubProcs(db *gorm.DB) error {
	return db.Preload("SubProcs", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "entry_id") // Optionally select specific fields.
	}).First(p).Error
}
