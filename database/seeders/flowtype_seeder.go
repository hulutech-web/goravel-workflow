package seeders

import (
	"github.com/goravel/framework/facades"
	"goravel/app/models"
)

type FlowtypeSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *FlowtypeSeeder) Signature() string {
	return "FlowtypeSeeder"
}

// Run executes the seeder logic.
func (s *FlowtypeSeeder) Run() error {
	var flowType models.Flowtype
	flowType.TypeName = "资金"
	facades.Orm().Query().Model(&flowType).Create(&flowType)
	return nil
}
