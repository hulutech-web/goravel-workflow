package requests

import (
	"github.com/goravel/framework/contracts/http"
)

type RevokeEntryRequest struct {
	EntryID int `form:"entry_id" json:"entry_id"`
}

func (r *RevokeEntryRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *RevokeEntryRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id": "required|integer|min:1",
	}
}

func (r *RevokeEntryRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id.required": "流程ID不能为空",
	}
}
