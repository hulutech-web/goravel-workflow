package requests

import (
	"github.com/goravel/framework/contracts/http"
)

type AddSignRequest struct {
	EntryID    int    `form:"entry_id" json:"entry_id"`
	ProcessID  int    `form:"process_id" json:"process_id"`
	SignEmpID  int    `form:"sign_emp_id" json:"sign_emp_id"`
	SignType   string `form:"sign_type" json:"sign_type"` // "before" or "after"
}

func (r *AddSignRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *AddSignRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id":    "required|integer|min:1",
		"process_id":  "required|integer|min:1",
		"sign_emp_id": "required|integer|min:1",
		"sign_type":   "required|in:before,after",
	}
}

func (r *AddSignRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id.required":   "流程ID不能为空",
		"sign_type.in":        "加签类型必须为 before 或 after",
	}
}
