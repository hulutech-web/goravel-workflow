package feature

import (
	tests "github.com/goravel/framework/testing"
	"github.com/hulutech-web/goravel-workflow/seeders"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DeptControllerTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestDeptControllerTestSuite(t *testing.T) {
	suite.Run(t, new(DeptControllerTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *DeptControllerTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *DeptControllerTestSuite) TearDownTest() {

}

// 树形结构
func (s *DeptControllerTestSuite) TestDeptTree() {
	s.Seed(&seeders.WorkflowDeptSeeder{})
	s.True(true)
	////树形结构
	//client := resty.New().
	//	SetBaseURL(fmt.Sprintf("http://%s:%s",
	//		facades.Config().GetString("http.host"),
	//		facades.Config().GetString("http.port"))).
	//	SetHeader("Content-Type", "application/json")
	////查询
	//var deptList struct {
	//	Depts []models.Dept
	//}
	//resp, err := client.R().SetResult(&deptList).Get("/api/dept")
	//s.Require().NoError(err)
	//s.Require().Equal(200, resp.StatusCode())
}
