package official_plugins

import (
	"github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/facades"
	gormio "gorm.io/gorm"
)

// 申明一个MYSQL连接GormIns
var GormIns *gormio.DB

func BootMS() *gormio.DB {
	var gormImpl = gorm.NewGormImpl(facades.Config(), "mysql", db.NewConfigImpl(facades.Config(), "mysql"), gorm.NewDialectorImpl(facades.Config(), "mysql"))
	//	获取实例
	gormIns, _ := gormImpl.Make()
	GormIns = gormIns
	return GormIns
}
