package seeders

import (
	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/facades"
)

type DatabaseSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *DatabaseSeeder) Signature() string {
	return "WorkflowSeeder"
}

// Run executes the seeder logic.
func (s *DatabaseSeeder) Run() error {
	return facades.Seeder().Call([]seeder.Seeder{
		&WorkflowFlowtypeSeeder{},
		&WorkflowEmpSeeder{},
		&WorkflowDeptSeeder{},
		&WorkflowUserSeeder{},
	})
}
