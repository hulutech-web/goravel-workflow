package requests

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/validation"
	"github.com/hulutech-web/goravel-workflow/models/common"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
)

type TemplateformRequest struct {
	ID                uint              `gorm:"column:id;primary_key;auto_increment;comment:'自增ID'" json:"id" form:"-"`
	Field             string            `gorm:"column:field;not null;default:'';comment:'表单字段英文名'" json:"field" form:"field"`
	FieldName         string            `gorm:"column:field_name;not null;default:'';comment:'表单字段中文名'" json:"field_name" form:"field_name"`
	FieldType         string            `gorm:"column:field_type;not null;default:'';comment:'表单字段类型'" json:"field_type" form:"field_type"`
	FieldValue        common.FieldValue `gorm:"column:field_value;type:text;comment:'表单字段值，select radio checkbox用'" json:"field_value" form:"field_value"`
	FieldDefaultValue string            `gorm:"column:field_default_value;type:text;comment:'表单字段默认值'" json:"field_default_value" form:"field_default_value"`
	FieldRules        common.Rule       `gorm:"column:field_rules;" json:"field_rules" form:"field_rules"`
	Sort              int               `gorm:"column:sort;not null;default:100;comment:'排序'" json:"sort" form:"sort"`
	TemplateID        uint              `gorm:"column:template_id;not null;default:0;comment:'模板ID'" json:"template_id" form:"template_id"`
}

func (r *TemplateformRequest) Authorize(ctx http.Context) error {
	return nil
}

func (r *TemplateformRequest) Rules(ctx http.Context) map[string]string {
	return map[string]string{
		"template_id": "required",
		"field":       "required|alpha_rule",
		"field_name":  "required",
		"field_type":  "required",
	}
}

func (r *TemplateformRequest) Messages(ctx http.Context) map[string]string {
	return map[string]string{
		"template_id.required": "模板ID不能为空",
		"field.required":       "表单字段英文名不能为空",
		"field.alpha_rule":     "表单字段英文名只能包含字母",
		"field_name.required":  "表单字段中文名不能为空",
		"field_type.required":  "表单字段类型不能为空",
	}
}

func (r *TemplateformRequest) Attributes(ctx http.Context) map[string]string {
	return map[string]string{
		"template_id":         "模板ID",
		"field":               "表单字段英文名",
		"field_name":          "表单字段中文名",
		"field_type":          "表单字段类型",
		"field_value":         "表单字段值",
		"field_default_value": "表单字段默认值",
		"field_rules":         "表单字段规则",
		"sort":                "排序",
	}
}

func (r *TemplateformRequest) PrepareForValidation(ctx http.Context, data validation.Data) error {
	if name, exist := data.Get("sort"); exist {
		data.Set("sort", cast.ToInt(name))
	}
	if val, exist := data.Get("field_rules"); exist {
		//将val转换为Rule类型
		r.FieldRules = common.Rule{}
		//	使用mapstruct将json字符串转换为Rule类型
		if err := mapstructure.Decode(val, &r.FieldRules); err != nil {
			return err
		}

	}
	return nil
}
