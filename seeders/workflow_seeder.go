package seeders

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/facades"
)

type WorkflowDatabaseSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *WorkflowDatabaseSeeder) Signature() string {
	return "WorkflowDatabaseSeeder"
}

// Run executes the seeder logic.
func (s *WorkflowDatabaseSeeder) Run() error {
	return facades.Seeder().Call([]seeder.Seeder{
		&WorkflowFlowtypeSeeder{},
		&WorkflowEmpSeeder{},
		&WorkflowDeptSeeder{},
		&WorkflowUserSeeder{},
	})
}
