package models

import (
	"github.com/goravel/framework/database/orm"
	"gorm.io/gorm"
)

type Flowtype struct {
	orm.Model
	TypeName string `gorm:"column:type_name;not null;default:''" json:"type_name"`
	Flows    []Flow `gorm:"-"`
}

// LoadFlowsForType preloads the associated Flows for a FlowType.
func (ft *Flowtype) LoadFlowsForType(db *gorm.DB) error {
	return db.Preload("Flows").Find(ft).Error
}
