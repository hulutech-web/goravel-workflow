package common

import (
	"fmt"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"goravel/app/models"
	"reflect"
	"slices"
	"strings"
)

type DynamicValidator struct {
}

func NewDynamicValidator() *DynamicValidator {
	return &DynamicValidator{}
}

func (r *DynamicValidator) DynamicValidate(flow_id int) (map[string]string, map[string]string) {
	var flow models.Flow
	facades.Orm().Query().Model(&models.Flow{}).Where("id", flow_id).First(&flow)
	template := models.Template{}
	facades.Orm().Query().Model(&models.Template{}).Where("id=?", flow.TemplateID).First(&template)
	if template.ID == 0 {
		return make(map[string]string), nil
	}
	template_forms := []models.TemplateForm{}

	facades.Orm().Query().Model(&models.TemplateForm{}).Where("template_id", template.ID).Find(&template_forms)
	var validateMap = make(map[string]string)
	var messageMap = make(map[string]string)
	ruleSlice := []string{"required", "string", "uint", "min_len", "max_len", "max", "min", "ne", "date", "file", "image", "number", "email", "slice"}
	for _, template_form := range template_forms {
		if template_form.FieldRules != nil {
			for _, rule := range template_form.FieldRules {
				//	如果ruleSlice中存在rule，则添加到validateMap中
				//最终形成的结构类似于required|uint|
				var truthRule string
				if slices.Contains(ruleSlice, rule.RuleName) {
					if rule.RuleName == "min_len" || rule.RuleName == "max_len" || rule.RuleName == "max" || rule.RuleName == "min" || rule.RuleName == "ne" {
						truthRule += fmt.Sprintf("%s:%s|", rule.RuleName, rule.RuleValue)
					} else {

						truthRule += fmt.Sprintf("%s|", rule.RuleName)
					}
					if rule.RuleName == "file" {
						//实际上是一个文本类型
						truthRule += fmt.Sprintf("%s|", "string")
					}
					if rule.RuleName == "required" {
						messageMap[fmt.Sprintf("%s.%s", template_form.Field, rule.RuleName)] = fmt.Sprintf("%s%s%s", "错误", rule.RuleTitle, rule.RuleValue)
					} else {
						messageMap[fmt.Sprintf("%s.%s", template_form.Field, rule.RuleName)] = fmt.Sprintf("%s[%s]%s", "错误", rule.RuleTitle, rule.RuleValue)
					}
					//将truthRule最后一个|去掉
					validateMap[template_form.Field] += truthRule
				}
			}
		}
	}
	//去掉validateMap中每个value的最后一个|
	for key, val := range validateMap {
		validateMap[key] = strings.TrimRight(val, "|")
	}
	return validateMap, messageMap
}

// 如果提交数据为int64类型，将其转换为int类型
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
