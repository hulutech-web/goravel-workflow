package official_plugins

import (
	"testing"
)

func TestAutoMigrate(t *testing.T) {
	db := BootMS()
	db.AutoMigrate(&DistributePluginConfig{})
}
