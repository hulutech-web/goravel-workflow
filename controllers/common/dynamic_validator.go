package common

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/hulutech-web/goravel-workflow/models"
)

type DynamicValidator struct {
}

func NewDynamicValidator() *DynamicValidator {
	return &DynamicValidator{}
}

func (r *DynamicValidator) DynamicValidate(flow_id int) (map[string]any, map[string]any) {
	var flow models.Flow
	facades.Orm().Query().Model(&models.Flow{}).Where("id", flow_id).Find(&flow)
	template := models.Template{}
	facades.Orm().Query().Model(&models.Template{}).Where("id=?", flow.TemplateID).Find(&template)
	if template.ID == 0 {
		return make(map[string]any), nil
	}
	template_forms := []models.TemplateForm{}
	facades.Orm().Query().Model(&models.TemplateForm{}).Where("template_id", template.ID).Find(&template_forms)

	var validateMap = make(map[string]any)
	var messageMap = make(map[string]any)

	// v1.18 全部合法规则名（snake_case）+ 旧规则名兼容映射
	allRules := []string{
		// 必填
		"required", "required_if", "required_unless", "required_with", "required_with_all",
		"required_without", "required_without_all", "required_if_accepted", "required_if_declined",
		// 存在性
		"filled", "present", "present_if", "present_unless", "present_with", "present_with_all",
		"missing", "missing_if", "missing_unless", "missing_with", "missing_with_all",
		// 接受/拒绝
		"accepted", "accepted_if", "declined", "declined_if",
		// 禁止
		"prohibited", "prohibited_if", "prohibited_unless", "prohibited_if_accepted", "prohibited_if_declined", "prohibits",
		// 类型
		"string", "integer", "int", "uint", "numeric", "boolean", "bool", "float", "array", "list", "slice", "map",
		// 大小
		"size", "min", "max", "between", "gt", "gte", "lt", "lte",
		// 数值
		"digits", "digits_between", "decimal", "multiple_of", "min_digits", "max_digits",
		// 字符串格式
		"alpha", "alpha_num", "alpha_dash", "ascii", "email", "url", "active_url", "ip",
		"ipv4", "ipv6", "mac_address", "mac", "json", "uuid", "uuid3", "uuid4", "uuid5", "ulid",
		"hex_color", "regex", "not_regex", "lowercase", "uppercase",
		// 字符串内容
		"starts_with", "doesnt_start_with", "ends_with", "doesnt_end_with", "contains", "doesnt_contain", "confirmed",
		// 比较
		"same", "different", "eq", "ne", "in", "not_in", "in_array", "in_array_keys",
		// 日期
		"date", "date_format", "date_equals", "before", "before_or_equal", "after", "after_or_equal", "timezone",
		// 排除
		"exclude", "exclude_if", "exclude_unless", "exclude_with", "exclude_without",
		// 数组/数据库
		"distinct", "required_array_keys", "exists", "unique",
		// 控制
		"bail", "nullable", "sometimes",
	}
	ruleAlias := map[string]string{
		"min_len": "min",
		"max_len": "max",
		"number":  "numeric",
	}

	// 需要参数的规则（带冒号值）
	paramRules := map[string]bool{
		"min": true, "max": true, "ne": true, "size": true, "between": true,
		"gt": true, "gte": true, "lt": true, "lte": true,
		"digits": true, "digits_between": true, "decimal": true, "multiple_of": true,
		"min_digits": true, "max_digits": true,
		"date_format": true, "before": true, "before_or_equal": true,
		"after": true, "after_or_equal": true,
		"starts_with": true, "doesnt_start_with": true, "ends_with": true, "doesnt_end_with": true, "contains": true,
		"same": true, "in": true, "not_in": true,
		"regex": true, "not_regex": true,
		"exists": true, "unique": true,
		"required_if": true, "required_if_accepted": true, "required_if_declined": true,
		"required_unless": true, "required_with": true, "required_with_all": true,
		"required_without": true, "required_without_all": true,
		"present_if": true, "present_unless": true, "present_with": true, "present_with_all": true,
		"missing_if": true, "missing_unless": true, "missing_with": true, "missing_with_all": true,
		"prohibited_if": true, "prohibited_unless": true,
		"prohibited_if_accepted": true, "prohibited_if_declined": true,
		"prohibits": true,
		"exclude_if": true, "exclude_unless": true, "exclude_with": true, "exclude_without": true,
		"in_array": true, "in_array_keys": true,
		"accepted_if": true, "declined_if": true,
	}

	// 需要额外追加 string 规则的（文件类）
	autoAppendString := map[string]bool{"file": true, "image": true}

	for _, tf := range template_forms {
		if tf.FieldRules == nil {
			continue
		}
		var rulesParts []string
		var messages []string
		for _, rule := range tf.FieldRules {
			checkName := rule.RuleName
			if alias, ok := ruleAlias[rule.RuleName]; ok {
				checkName = alias
			} else if !slices.Contains(allRules, checkName) {
				continue
			}

			if paramRules[checkName] {
				if rule.RuleValue != "" {
					rulesParts = append(rulesParts, fmt.Sprintf("%s:%s", checkName, rule.RuleValue))
				}
			} else {
				rulesParts = append(rulesParts, checkName)
			}

			if autoAppendString[checkName] {
				rulesParts = append(rulesParts, "string")
			}

			// 构建错误消息
			msgPrefix := "错误"
			if checkName == "required" {
				messages = append(messages, fmt.Sprintf("%s%s%s", msgPrefix, rule.RuleTitle, rule.RuleValue))
			} else {
				messages = append(messages, fmt.Sprintf("%s[%s]%s", msgPrefix, rule.RuleTitle, rule.RuleValue))
			}
		}

		if len(rulesParts) > 0 {
			validateMap[tf.Field] = strings.Join(rulesParts, "|")
			for i, msg := range messages {
				if i < len(rulesParts) {
					key := fmt.Sprintf("%s.%s", tf.Field, parseRuleKey(rulesParts[i]))
					messageMap[key] = msg
				}
			}
		}
	}

	return validateMap, messageMap
}

// parseRuleKey 从 "rule:value" 或 "rule" 提取规则名作为 messageMap key
func parseRuleKey(part string) string {
	if idx := strings.Index(part, ":"); idx > 0 {
		return part[:idx]
	}
	return part
}

// DynamicValidateField converts int64 values to int for validation
func (r *DynamicValidator) DynamicValidateField(ctx http.Context) map[string]any {
	result := map[string]any{}
	requests := ctx.Request().All()
	for key, val := range requests {
		atype := reflect.TypeOf(val)
		if atype.Name() == "float64" {
			result[key] = int(val.(float64))
		} else {
			result[key] = val
		}
	}
	return result
}
