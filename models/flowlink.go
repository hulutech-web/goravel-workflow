package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

type Flowlink struct {
	orm.Model
	Type       string `gorm:"column:type;not null;comment:'Condition:表示步骤流转\\nRole:当前步骤操作人'"`                             // 类型：Condition或Role
	Auditor    string `gorm:"column:auditor;not null;default:'0';comment:'审批人 系统自动 指定人员 指定部门 指定角色\\ntype=Condition时不启用'"` // 审批人设置
	Expression string `gorm:"column:expression;not null;default:'';comment:'条件判断表达式\\n为1表示true，通过的话直接进入下一步骤'"`            // 条件判断表达式
	Sort       int    `gorm:"column:sort;not null;comment:'条件判断顺序'"`                                                      // 判断顺序

	FlowID        uint    `gorm:"column:flow_id"`                         // 流程ID
	ProcessID     uint    `gorm:"column:process_id"`                      // 当前步骤ID
	NextProcessID int     `gorm:"column:next_process_id;default:2"`       // 下一步骤ID
	Process       Process `gorm:"foreignKey:ProcessID;references:id"`     // HasOne Process
	NextProcess   Process `gorm:"foreignKey:NextProcessID;references:id"` // HasOne NextProcess
	Flow          Flow    `gorm:"foreignKey:FlowID;references:id"`        // BelongsTo Flow
}

func (fl *Flowlink) LoadProcess(db *gorm.DB) error {
	return db.Preload("Process").Find(fl).Error
}

func (fl *Flowlink) LoadProcesses(db *gorm.DB) error {
	return db.Preload("Processes").Find(fl).Error
}

func (fl *Flowlink) LoadNextProcess(db *gorm.DB) error {
	return db.Preload("NextProcess").Find(fl).Error
}
