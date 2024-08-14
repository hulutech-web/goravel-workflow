package seeders

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
)

type WorkflowUserSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *WorkflowUserSeeder) Signature() string {
	return "WorkflowUserSeeder"
}

// Run executes the seeder logic.
func (s *WorkflowUserSeeder) Run() error {
	emps := []models.Emp{}
	facades.Orm().Query().Model(&models.Emp{}).Find(&emps)
	for _, emp := range emps {
		var user models.User
		password, _ := facades.Hash().Make("admin888")
		user.Name = emp.Name
		user.Mobile = emp.WorkNo
		user.Password = password
		user.IdNumber = gofakeit.CreditCardNumber(nil)
		facades.Orm().Query().Model(&models.User{}).Create(&user)
	}
	return nil
}
