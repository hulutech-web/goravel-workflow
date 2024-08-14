package models

import (
	"github.com/goravel/framework/database/orm"
)

type Template struct {
	orm.Model
	TemplateName  string `gorm:"column:template_name;not null;default:''" json:"template_name"`
	TemplateForms []TemplateForm
}
