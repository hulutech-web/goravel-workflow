package requests

import (
	"github.com/goravel/framework/contracts/http"
)

type CcRequest struct {
	EntryID int `form:"entry_id" json:"entry_id"`
}

func (r *CcRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *CcRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id": "required|integer|min:1",
	}
}

func (r *CcRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id.required": "流程ID不能为空",
	}
}
