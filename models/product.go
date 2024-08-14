package models

import (
	"github.com/goravel/framework/database/orm"
)

type Product struct {
	orm.Model
	Name          string  `gorm:"column:name;type:varchar(255);not null" form:"name" json:"name"`
	Special       string  `gorm:"column:special;type:varchar(255);not null" form:"special" json:"special"`
	Dimension     string  `gorm:"column:dimension;type:varchar(255);not null" form:"dimension" json:"dimension"`
	Quantity      int     `gorm:"column:quantity;type:int(10);not null" form:"quantity" json:"quantity"`
	Unit          string  `gorm:"column:unit;type:varchar(255);not null" form:"unit" json:"unit"`
	UnitPrice     float64 `gorm:"column:unit_price;type:float(10,2);not null" form:"unit_price" json:"unit_price"`
	DiscountPrice float64 `gorm:"column:discount_price;type:float(10,2);not null" form:"discount_price" json:"discount_price"`
	Amount        float64 `gorm:"column:amount;type:float(10,2);not null" form:"amount" json:"amount"`
	Description   string  `gorm:"column:description;type:varchar(255);" form:"description" json:"description"`
	ImageURL      string  `gorm:"column:image_url;type:varchar(255);" form:"image_url" json:"image_url"`
}
