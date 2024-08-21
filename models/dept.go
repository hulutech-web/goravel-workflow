package models

import (
	"github.com/goravel/framework/database/orm"
	"strings"
)

type Dept struct {
	orm.Model
	DeptName   string `gorm:"column:dept_name;not null;default:''" json:"dept_name" form:"dept_name"`
	Pid        uint   `gorm:"column:pid;not null;default:0" json:"pid" form:"pid"`
	DirectorID int    `gorm:"column:director_id;not null;default:0" json:"director_id" form:"derector_id"` // 部门主管
	ManagerID  int    `gorm:"column:manager_id;not null;default:0" json:"manager_id" form:"manager_id"`    // 部门经理
	Rank       int    `gorm:"column:rank;not null;default:1" json:"rank" form:"rank"`
	Html       string `gorm:"column:html;null;default:''" json:"html" form:"html"`
	Level      int    `gorm:"column:level;null;default:0" json:"level" form:"level"`
	Director   *Emp   `gorm:"foreignkey:DirectorID"` // 关联主管
	Manager    *Emp   `gorm:"foreignkey:ManagerID"`  // 关联经理
}

func (d *Dept) Recursion(models []Dept, html string, pid uint, level int) []Dept {
	var result []Dept
	for i, dept := range models {
		if dept.Pid == pid {
			dept.Html = strings.Repeat(html, level)
			dept.Level = level + 1
			result = append(result, dept)
			result = append(result, d.Recursion(append([]Dept{}, models[i+1:]...), html, dept.ID, level+1)...)
		}
	}
	return result
}
