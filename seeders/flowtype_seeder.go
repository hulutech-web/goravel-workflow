package seeders

import (
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
)

type WorkflowFlowtypeSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *WorkflowFlowtypeSeeder) Signature() string {
	return "WorkflowFlowtypeSeeder"
}

// Run executes the seeder logic.
func (s *WorkflowFlowtypeSeeder) Run() error {
	var flowType models.Flowtype
	flowType.TypeName = "资金"
	facades.Orm().Query().Model(&flowType).Create(&flowType)
	return nil
}
