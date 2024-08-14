package seeders

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
)

type UserSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *UserSeeder) Signature() string {
	return "UserSeeder"
}

// Run executes the seeder logic.
func (s *UserSeeder) Run() error {
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
