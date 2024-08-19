package feature

import (
	tests "github.com/goravel/framework/testing"
	"github.com/hulutech-web/goravel-workflow/models"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type EmpTestSuite struct {
	suite.Suite
	tests.TestCase
	DB *gorm.DB
}

func TestEmpTestSuite(t *testing.T) {
	suite.Run(t, new(EmpTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *EmpTestSuite) SetupTest() {
	var err error
	s.DB, err = getGormDBConnection()
	if err != nil {
		s.FailNow(err.Error())
	}
}

// TearDownTest will run after each test in the suite.
func (s *EmpTestSuite) TearDownTest() {
	sqlDB, err := s.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (s *EmpTestSuite) TestEmpIndex() {
	empList := []models.Emp{}
	s.DB.Find(&empList)
	s.True(true)
}
func (s *EmpTestSuite) TestEmpStore() {
	//创建
	s.DB.Create(&models.Emp{
		Name:     "test",
		WorkNo:   "0009",
		Email:    "test@test.com",
		Password: "password",
	})
	s.True(true)
}

func (s *EmpTestSuite) TestEmpUpdate() {
	//更新
	emp := models.Emp{}
	emp.Name = "test"
	s.DB.Where("name=?", emp.Name).Update("name", "test_name")
	s.True(true)
}
func (s *EmpTestSuite) TestEmpShow() {
	//查询
	emp := models.Emp{}
	s.DB.First(&emp, "workno = ?", "30001")
}

func (s *EmpTestSuite) TestEmpDestroy() {
	//删除
	s.DB.Where("workno=?", "0009").Delete(&models.Emp{})
}

func (s *EmpTestSuite) TestEmpBind() {
	//绑定用户
	s.DB.Model(&models.Emp{}).Where("id=?", 17).Update("user_id", 20)
}
