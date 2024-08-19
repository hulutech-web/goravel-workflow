package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRecursion(t *testing.T) {
	// 准备测试数据
	depts := []Dept{
		{DeptName: "Dept 1", Pid: 0},
		{DeptName: "Dept 2", Pid: 1},
		{DeptName: "Dept 3", Pid: 1},
		{DeptName: "Dept 4", Pid: 2},
	}

	d := &Dept{}
	expected := []Dept{
		{DeptName: "Dept 1", Pid: 0, Html: "", Level: 1},
		{DeptName: "Dept 2", Pid: 1, Html: "|---", Level: 2},
		{DeptName: "Dept 4", Pid: 2, Html: "|---|---", Level: 3},
		{DeptName: "Dept 3", Pid: 1, Html: "|---", Level: 2},
	}

	result := d.Recursion(depts, "|---", 0, 0)

	assert.Equal(t, expected, result, "Recursion result does not match expected")
}
