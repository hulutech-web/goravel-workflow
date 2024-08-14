package rules

import (
	"github.com/goravel/framework/contracts/validation"
	"strings"
)

type AlphaRule struct {
}

// Signature The name of the rule.
func (receiver *AlphaRule) Signature() string {
	return "alpha_rule"
}

// Passes Determine if the validation rule passes.
func (receiver *AlphaRule) Passes(data validation.Data, val any, options ...any) bool {
	if val == "" {
		return true
	}
	for _, ch := range []rune(val.(string)) {
		if !('a' <= ch && ch <= 'z') && !('A' <= ch && ch <= 'Z') && !(strings.Contains("_", string(ch))) {
			return false
		}
	}
	return true
}

// Message Get the validation error message.
func (receiver *AlphaRule) Message() string {
	return "字符 :attribute 必须是英文字符或者包含下划线"
}
