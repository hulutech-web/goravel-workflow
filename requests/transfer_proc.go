package requests

import (
	"github.com/goravel/framework/contracts/http"
)

type TransferProcRequest struct {
	EntryID      int `form:"entry_id" json:"entry_id"`
	ProcID       int `form:"proc_id" json:"proc_id"`
	TargetEmpID  int `form:"target_emp_id" json:"target_emp_id"`
}

func (r *TransferProcRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *TransferProcRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id":     "required|integer|min:1",
		"proc_id":      "required|integer|min:1",
		"target_emp_id": "required|integer|min:1",
	}
}

func (r *TransferProcRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id.required":     "流程ID不能为空",
		"target_emp_id.required": "目标员工ID不能为空",
	}
}
