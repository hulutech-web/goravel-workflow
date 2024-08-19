package feature

import (
	tests "github.com/goravel/framework/testing"
	"github.com/hulutech-web/goravel-workflow/models"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
)

type DeptTestSuite struct {
	suite.Suite
	tests.TestCase
	DB *gorm.DB
}

func TestDeptTestSuite(t *testing.T) {
	suite.Run(t, new(DeptTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *DeptTestSuite) SetupTest() {
	var err error
	s.DB, err = getGormDBConnection()
	if err != nil {
		s.FailNow(err.Error())
	}
}

// TearDownTest will run after each test in the suite.
func (s *DeptTestSuite) TearDownTest() {
	sqlDB, err := s.DB.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func (s *DeptTestSuite) TestDeptIndex() {
	DeptList := []models.Dept{}
	s.DB.Find(&DeptList)
	s.True(true)
}

func (s *DeptTestSuite) TestDeptStore() {
	//创建
	s.DB.Create(&models.Dept{
		DeptName:   "客服部",
		Pid:        0,
		DirectorID: 0,
		ManagerID:  0,
		Rank:       1,
		Html:       "|-",
		Level:      1,
	})
	s.True(true)
}

func (s *DeptTestSuite) TestDeptUpdate() {
	//更新
	Dept := models.Dept{}
	Dept.DeptName = "客服部"
	s.DB.Where("dept_name=?", Dept.DeptName).Update("dept_name", "客服部1")
	s.True(true)
}

func (s *DeptTestSuite) TestDeptShow() {
	//查询
	Dept := models.Dept{}
	s.DB.First(&Dept, "dept_name like ?", "客服部")
	s.True(true)

}

func (s *DeptTestSuite) TestDeptDestroy() {
	//删除
	s.DB.Where("dept_name=?", "客服部1").Delete(&models.Dept{})
	s.True(true)

}

// 绑定经理
func (s *DeptTestSuite) TestDeptBindManager() {
	//绑定经理
	s.DB.Model(&models.Dept{}).Where("id=?", 17).Update("manager_id", 20)
	s.True(true)

}

// 绑定主管
func (s *DeptTestSuite) TestDeptBindDirector() {
	//绑定主管
	s.DB.Model(&models.Dept{}).Where("id=?", 17).Update("director_id", 20)
	s.True(true)

}

// 树形结构
func (s *DeptTestSuite) TestDeptTree() {
	//树形结构
	var depts []models.Dept
	s.DB.Find(&depts)
	deptInstance := models.Dept{}

	deptInstance.Recursion(depts, "|---", 0, 0)
	s.True(true)

}
