package seeders

import (
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
)

type DeptSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *DeptSeeder) Signature() string {
	return "DeptSeeder"
}

// Run executes the seeder logic.
func (s *DeptSeeder) Run() error {
	dept := models.Dept{}
	//1-总部
	query := facades.Orm().Query()
	query.Model(&dept).Create(&models.Dept{
		DeptName:   "总部",
		Pid:        0,
		ManagerID:  1,
		DirectorID: 1,
	})
	//2-技术部
	query.Model(&dept).Create(&models.Dept{
		DeptName:   "技术部",
		Pid:        1,
		Html:       "|-",
		ManagerID:  2,
		DirectorID: 3,
	})
	//3-财务部
	query.Model(&dept).Create(&models.Dept{
		DeptName:   "财务部",
		Pid:        1,
		Html:       "|-",
		ManagerID:  5,
		DirectorID: 6,
	})
	// 4-市场部
	query.Model(&dept).Create(&models.Dept{
		DeptName:   "市场部",
		Pid:        1,
		Html:       "|-",
		ManagerID:  8,
		DirectorID: 9,
	})
	//4-1市场部-销售部
	query.Model(&dept).Create(&models.Dept{
		DeptName:   "市场部-销售部",
		Pid:        4,
		Html:       "|-",
		ManagerID:  11,
		DirectorID: 12,
	})
	//4-2市场部-市场拓展部
	query.Model(&dept).Create(&models.Dept{
		DeptName:   "市场部-市场拓展部",
		Pid:        4,
		Html:       "|-",
		ManagerID:  14,
		DirectorID: 15,
	})
	return nil
}
