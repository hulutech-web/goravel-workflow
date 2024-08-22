package controllers

import (
	"fmt"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/validation"
	"github.com/hulutech-web/goravel-workflow/controllers/common"
	"github.com/hulutech-web/goravel-workflow/models"
	"github.com/hulutech-web/goravel-workflow/services/workflow"
	"github.com/hulutech-web/goravel-workflow/services/workflow/official_plugins"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
	"reflect"
	"strings"
)

type EntryController struct {
	workflow         *workflow.Workflow
	dynamicValidator *common.DynamicValidator
}

func NewEntryController() *EntryController {
	return &EntryController{
		dynamicValidator: common.NewDynamicValidator(),
	}
}

func (r *EntryController) Create(ctx http.Context) http.Response {
	flow_id := ctx.Request().RouteInt("id")
	var flow models.Flow
	facades.Orm().Query().Model(&models.Flow{}).Where("id", flow_id).
		With("Template.TemplateForms").Find(&flow)
	return httpfacades.NewResult(ctx).Success("", flow)
}

func (r *EntryController) Index(ctx http.Context) http.Response {
	return nil
}

func (r *EntryController) Show(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var entry models.Entry
	facades.Orm().Query().Model(&models.Entry{}).With("EntryDatas").With("Flow.Template.TemplateForms").Where("id", id).Find(&entry)
	return httpfacades.NewResult(ctx).Success("", entry)
}

func (r *EntryController) EntryData(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var entrydata []models.EntryData
	var entry models.Entry
	query := facades.Orm().Query()
	query.Model(&models.Entry{}).Where("id=?", id).Find(&entry)
	//当时子流程时，需要查找当前流程的父流程
	query.Model(&models.EntryData{}).Where("entry_id=?", id).OrWhere("entry_id=?", entry.Pid).Find(&entrydata)

	last_flowlink := models.Flowlink{}
	query.Model(&models.Flowlink{}).Where("next_process_id=?", entry.ProcessID).
		Where("type=?", "Condition").Find(&last_flowlink)
	plugin_configs := official_plugins.PluginConfig{}
	//找上一个process
	query.Model(&official_plugins.PluginConfig{}).Where("process_id=?", last_flowlink.ProcessID).Find(&plugin_configs)
	return httpfacades.NewResult(ctx).Success("", http.Json{
		"entry":          entry,
		"entrydata":      entrydata,
		"plugin_configs": plugin_configs,
	})
}

func (r *EntryController) Store(ctx http.Context) http.Response {
	//添加发起节点
	flow_id := ctx.Request().InputInt("flow_id")
	var user models.Emp
	facades.Auth(ctx).User(&user)

	flowlink := models.Flowlink{}
	facades.Orm().Query().Table("flowlinks").Where("flowlinks.flow_id=?", cast.ToUint(flow_id)).Where("flowlinks.type=?", "Condition").Join("left join processes on flowlinks.id=processes.id").
		Where("processes.position=?", 0).Order("sort  ASC").Find(&flowlink)
	dbSql := fmt.Sprintf("SELECT * "+
		"FROM `flowlinks` "+
		"WHERE `flow_id` = %d "+
		"  AND `type` = 'Condition' "+
		"  AND EXISTS ("+
		"    SELECT 1 "+
		"    FROM `processes` "+
		"   WHERE `flowlinks`.`process_id` = `processes`.`id` "+
		"      AND `processes`.`position` = 0"+
		"  ) "+
		"ORDER BY `sort` ASC "+
		"LIMIT 1;", flow_id)
	facades.Orm().Query().Raw(dbSql).Scan(&flowlink)
	var withFlowlink models.Flowlink
	facades.Orm().Query().Model(&models.Flowlink{}).Where("id=?", flowlink.ID).
		With("Process").With("NextProcess").Find(&withFlowlink)
	//校验提交的数据
	validRule, validMsg := r.dynamicValidator.DynamicValidate(flow_id)
	validator, err := facades.Validation().Make(r.dynamicValidator.DynamicValidateField(ctx), validRule, validation.Messages(validMsg))
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, err.Error(), "")
	}
	if validator.Fails() {
		return httpfacades.NewResult(ctx).ValidError("", validator.Errors().All())
	}
	query := facades.Orm().Query()
	var entry models.Entry
	entry.Title = ctx.Request().Input("title")
	entry.FlowID = cast.ToUint(flow_id)
	entry.EmpID = user.ID
	entry.Circle = 1
	entry.Status = 0
	err = query.Model(&models.Entry{}).Create(&entry)

	var withEntry models.Entry
	query.Model(&models.Entry{}).Where("id=?", entry.ID).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").
		Find(&withEntry)
	//进程初始化
	//第一步看是否指定审核人

	err = r.workflow.SetFirstProcessAuditor(withEntry, withFlowlink)

	//向entrydata中插入数据
	for key, val := range ctx.Request().All() {
		if key == "title" || key == "flow_id" {
			continue
		} else {
			//判断val的类型，如果是[]string,则转换为解析为字符串

			if reflect.TypeOf(val).Kind() == reflect.Slice {
				var sliceStr []string
				//将val解析为sliceStr
				for _, v := range val.([]interface{}) {
					sliceStr = append(sliceStr, cast.ToString(v))
				}
				var newVal string
				newVal = strings.Join(sliceStr, ",")
				var entryData models.EntryData
				entryData.FlowID = cast.ToInt(flow_id)
				entryData.EntryID = cast.ToInt(entry.ID)
				entryData.FieldName = key
				entryData.FieldValue = newVal
				query.Model(&models.EntryData{}).Create(&entryData)
			} else {
				var entryData models.EntryData
				entryData.FlowID = cast.ToInt(flow_id)
				entryData.EntryID = cast.ToInt(entry.ID)
				entryData.FieldName = key
				entryData.FieldValue = cast.ToString(val)
				query.Model(&models.EntryData{}).Create(&entryData)
			}
		}
	}
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, err.Error(), "")
	}
	//流程表单数据插入，需要goravel的验证规则
	return httpfacades.NewResult(ctx).Success("发起成功", entry)
}

func (r *EntryController) Update(ctx http.Context) http.Response {
	return nil
}

func (r *EntryController) Destroy(ctx http.Context) http.Response {
	return nil
}

// 重发
func (r *EntryController) Resend(ctx http.Context) http.Response {
	entry_id := ctx.Request().Input("entry_id")
	entry := models.Entry{}
	query := facades.Orm().Query()
	query.Model(&models.Entry{}).Where("id=?", entry_id).Where("status=?", -1).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").
		Find(&entry)

	flow := models.Flow{}

	query.Model(&models.Flow{}).Where("id=?", entry.FlowID).Where("is_publish=?", true).Find(&flow)
	if flow.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程未发布，请检查", "")
	}
	var flowlink models.Flowlink

	sql := fmt.Sprintf("SELECT * FROM `flowlinks` WHERE `flow_id` = %d "+
		"AND EXISTS (SELECT 1 FROM `processes` WHERE `processes`.`id` = `flowlinks`.`process_id` AND `processes`.`position` = 0) ORDER BY `sort` ASC LIMIT 1;", entry.FlowID)
	query.Raw(sql).Scan(&flowlink)
	if flowlink.ID == 0 {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "节点关系错误，请检查", "")
	}
	var withFlowlink models.Flowlink
	facades.Orm().Query().Model(&models.Flowlink{}).Where("id=?", flowlink.ID).
		With("Process").With("NextProcess").Find(&withFlowlink)
	//零值更新
	var map_entry = make(map[string]interface{})
	map_entry["circle"] = entry.Circle + 1
	map_entry["child"] = 0
	map_entry["status"] = 0
	query.Model(&models.Entry{}).Where("id=?", entry.ID).Update(map_entry)
	newEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", entry.ID).With("Flow").With("Emp.Dept").With("Procs").With("EnterProcess").Find(&newEntry)

	err := r.workflow.SetFirstProcessAuditor(newEntry, withFlowlink)
	if err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "系统错误，请检查", "")
	}
	return httpfacades.NewResult(ctx).Success("重发成功", entry)
}
