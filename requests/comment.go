package requests

import (
	"github.com/goravel/framework/contracts/http"
)

type CommentRequest struct {
	EntryID int    `form:"entry_id" json:"entry_id"`
	ProcID  int    `form:"proc_id" json:"proc_id"`
	Content string `form:"content" json:"content"`
}

func (r *CommentRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *CommentRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"entry_id": "required|integer|min:1",
		"content":  "required|string|max:500",
	}
}

func (r *CommentRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"content.max": "评论内容不能超过500字",
	}
}
