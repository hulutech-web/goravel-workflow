package official_plugins

import (
	"github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/facades"
	gormio "gorm.io/gorm"
	"sync"
)

var (
	once sync.Once
)

// 申明一个MYSQL连接GormIns
var gormIns *gormio.DB

func BootMS() *gormio.DB {
	once.Do(func() {
		facades.Config().Add("APP_DEBUG", false)
		var gormImpl = gorm.NewGormImpl(facades.Config(), "mysql",
			db.NewConfigImpl(facades.Config(), "mysql"),
			gorm.NewDialectorImpl(facades.Config(), "mysql"))
		gormIns, _ = gormImpl.Make()
	})
	return gormIns
}
