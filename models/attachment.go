package models

import (
	"github.com/goravel/framework/database/orm"
)

type Attachment struct {
	orm.Model
	UserID uint   `gorm:"column:user_id;type:int(11)" form:"user_id" json:"user_id"`
	Name   string `gorm:"column:name;type:varchar(255);not null" json:"name" form:"name"`
	Path   string `gorm:"column:path;type:varchar(255);not null" json:"path" form:"path"`
	Ext    string `gorm:"column:ext;type:varchar(255);not null" json:"ext" form:"ext"`
}
