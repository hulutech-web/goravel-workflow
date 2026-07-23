package official_plugins

import (
	"sync"

	"github.com/goravel/framework/contracts/database/driver"
	gormdriver "github.com/goravel/framework/database/driver"
	"github.com/goravel/framework/facades"
	gormio "gorm.io/gorm"
)

var (
	gormIns     *gormio.DB
	gormInsOnce sync.Once
)

func BootMS() *gormio.DB {
	gormInsOnce.Do(func() {
		driverCallback, exist := facades.Config().Get("database.connections.mysql.via").(func() (driver.Driver, error))
		if !exist || driverCallback == nil {
			return
		}
		drv, err := driverCallback()
		if err != nil {
			return
		}
		pool := drv.Pool()
		db, _, err := gormdriver.BuildGorm(facades.Config(), nil, pool, "mysql", nil)
		if err != nil {
			return
		}
		gormIns = db
	})
	return gormIns
}
