package seeders

import (
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
)

type WorkflowEmpSeeder struct {
}

// Signature The name and signature of the seeder.
func (s *WorkflowEmpSeeder) Signature() string {
	return "WorkflowEmpSeeder"
}

// Run executes the seeder logic.
func (s *WorkflowEmpSeeder) Run() error {
	emp := models.Emp{}
	password, _ := facades.Hash().Make("admin888")
	//1-总部
	query := facades.Orm().Query()
	query.Model(&emp).Create(&models.Emp{
		Name:     "董事长",
		WorkNo:   "0001",
		DeptID:   1,
		UserID:   1,
		Password: password,
	})
	//2-技术部
	query.Model(&emp).Create(&models.Emp{
		Name:     "技术部-技术经理",
		WorkNo:   "10001",
		DeptID:   2,
		UserID:   2,
		Password: password,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "技术部-技术主管",
		WorkNo:   "10002",
		DeptID:   2,
		Password: password,
		UserID:   3,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "技术部-技术员",
		WorkNo:   "10003",
		DeptID:   2,
		Password: password,
		UserID:   4,
	})
	//3-财务部
	query.Model(&emp).Create(&models.Emp{
		Name:     "财务部-经理",
		WorkNo:   "20001",
		DeptID:   3,
		Password: password,
		UserID:   5,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "财务部-主管",
		WorkNo:   "20002",
		DeptID:   3,
		Password: password,
		UserID:   6,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "财务部-财务员",
		WorkNo:   "20003",
		DeptID:   3,
		Password: password,
		UserID:   7,
	})
	// 4-市场部
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-经理",
		WorkNo:   "30001",
		Password: password,
		DeptID:   4,
		UserID:   8,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-主管",
		WorkNo:   "30002",
		DeptID:   4,
		Password: password,
		UserID:   9,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-总部员工1",
		WorkNo:   "30003",
		DeptID:   4,
		Password: password,
		UserID:   10,
	})
	//4-1市场部-销售部
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-销售部-经理",
		WorkNo:   "30011",
		DeptID:   5,
		Password: password,
		UserID:   11,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-销售部-主管",
		WorkNo:   "30012",
		DeptID:   5,
		Password: password,
		UserID:   12,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-销售部-员工1",
		WorkNo:   "30013",
		DeptID:   5,
		Password: password,
		UserID:   13,
	})
	//4-1市场部-扩展部
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-扩展部-经理",
		WorkNo:   "30021",
		DeptID:   6,
		Password: password,
		UserID:   14,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-扩展部-主管",
		WorkNo:   "30022",
		DeptID:   6,
		Password: password,
		UserID:   15,
	})
	query.Model(&emp).Create(&models.Emp{
		Name:     "市场部-扩展部-员工1",
		WorkNo:   "30023",
		DeptID:   6,
		Password: password,
		UserID:   16,
	})

	return nil
}
