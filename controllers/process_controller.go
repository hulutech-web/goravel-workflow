package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
	httpfacades "github.com/hulutech-web/http_result"
	"github.com/spf13/cast"
	"goravel/app/http/controllers/common"
	"goravel/app/models"
	"strings"
)

type ProcessController struct {
	//Dependent services
}

func NewProcessController() *ProcessController {
	return &ProcessController{
		//Inject services
	}
}

func (r *ProcessController) Index(ctx http.Context) http.Response {
	return nil
}

func (r *ProcessController) Show(ctx http.Context) http.Response {
	return nil
}

func (r *ProcessController) Store(ctx http.Context) http.Response {
	flow_id := ctx.Request().InputInt("flow_id")
	left := ctx.Request().Input("left")
	top := ctx.Request().Input("top")
	tx, _ := facades.Orm().Query().Begin()
	var process models.Process
	var flow models.Flow
	tx.Model(&models.Flow{}).Where("id=?", flow_id).First(&flow)

	//步骤一
	process.FlowID = flow_id
	process.ProcessName = "新建流程"
	process.StyleWidth = 200
	process.StyleHeight = 48
	process.Style = fmt.Sprintf("width:200px;height:48px;line-height:30px;color:#78a300;left:%s;top:%s;", left, top)
	process.PositionLeft = left
	process.PositionTop = top
	if err := tx.Model(&models.Process{}).Create(&process); err != nil {
		tx.Rollback()
	}

	//步骤二
	jsMap := common.Plumb{}
	if flow.Jsplumb == "" {
		//添加属性
		jsMap.Total = 1
		jsMap.List = map[string]common.Node{}
		listMap := map[string]common.Node{}
		node := common.Node{
			ID:          cast.ToInt(process.ID),
			FlowId:      process.FlowID,
			ProcessName: process.ProcessName,
			ProcessTo:   "",
			Icon:        "",
			Style:       process.Style,
		}
		listMap[cast.ToString(process.ID)] = node
		jsMap.List = listMap
		strByte, _ := json.Marshal(jsMap)
		tx.Model(&models.Flow{}).Where("id=?", flow_id).Update("jsplumb", strByte)
		flow.IsPublish = false
		tx.Model(&models.Flow{}).Where("id=?", flow_id).Update(&flow)
		tx.Commit()
		return httpfacades.NewResult(ctx).Success("", http.Json{
			"id":           process.ID,
			"flow_id":      process.FlowID,
			"process_name": process.ProcessName,
			"process_to":   "",
			"icon":         "",
			"style":        process.Style,
		})
	} else {
		//jsMap的list属性为二维数组
		var jsMapTemp common.Plumb
		//将flow中的Jsplumb转换为jsMapTemp
		json.Unmarshal([]byte(flow.Jsplumb), &jsMapTemp)
		node := common.Node{
			ID:          cast.ToInt(process.ID),
			FlowId:      process.FlowID,
			ProcessName: process.ProcessName,
			ProcessTo:   "",
			Icon:        "",
			Style:       process.Style,
		}
		jsMapTemp.List[cast.ToString(process.ID)] = node
		jsMap = jsMapTemp
		//转换jsMap为json
		strByte, _ := json.Marshal(jsMap)
		flow.Jsplumb = string(strByte)
		flow.IsPublish = false
		tx.Model(&models.Flow{}).Where("id=?", flow_id).Update(&flow)
		tx.Commit()
		return httpfacades.NewResult(ctx).Success("", http.Json{
			"id":           process.ID,
			"flow_id":      process.FlowID,
			"process_name": process.ProcessName,
			"process_to":   "",
			"icon":         "",
			"style":        process.Style,
		})
	}
}

func (r *ProcessController) Update(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	var process models.Process
	tx, _ := facades.Orm().Query().Begin()
	err := tx.Model(&models.Process{}).Where("id=?", id).First(&process)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程不存在", nil)
	}
	var processRequest common.ProcessRequest
	if err := ctx.Request().Bind(&processRequest); err != nil {
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
	}

	if processRequest.ProcessPosition == 9 {
		var count int64
		tx.Model(&models.Flowlink{}).Where("process_id=?", id).Count(&count)
		if count > 1 {
			return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "该节点是分支节点，不能设置为结束或起始步骤", nil)
		}
	}
	if processRequest.ProcessPosition == 0 {
		_, err := tx.Model(&models.Process{}).Where("flow_id=?", process.FlowID).Where("position", 0).Update("position", 1)
		tx.Model(&models.Process{}).Where("flow_id=?", process.FlowID).Update("position", 0)
		if err != nil {
			tx.Rollback()
			return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
		}
	}
	process.ProcessName = processRequest.ProcessName
	process.StyleColor = processRequest.StyleColor
	process.StyleHeight = processRequest.StyleHeight
	process.StyleWidth = processRequest.StyleWidth
	process.Style = fmt.Sprintf("width:%dpx;height:%dpx;line-height:30px;color:%s;left:%s;top:%s;",
		process.StyleWidth, process.StyleHeight, process.StyleColor, process.PositionLeft, process.PositionTop)
	process.Icon = processRequest.StyleIcon
	process.Position = processRequest.ProcessPosition
	process.ChildFlowID = processRequest.ChildFlowId
	process.ChildAfter = processRequest.ChildAfter
	process.ChildBackProcess = processRequest.ChildBackProcess
	if err := tx.Model(&models.Process{}).Where("id=?", id).Save(&process); err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
	}
	// 同步更新jsplumb json数据
	var flow models.Flow
	err = tx.Model(&models.Flow{}).Where("id=?", process.FlowID).With("Template.TemplateForms").First(&flow)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "流程不存在", nil)
	}
	jsMap := common.Plumb{}
	//flow.Jsplum解析为jsMap
	err = json.Unmarshal([]byte(flow.Jsplumb), &jsMap)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析数据错误", nil)
	}

	//需要将jsMap读取出来，然后再写回去
	for key, val := range jsMap.List {
		if key == cast.ToString(process.ID) {
			jsMap.List[key] = common.Node{
				ID:          cast.ToInt(process.ID),
				FlowId:      process.FlowID,
				ProcessTo:   val.ProcessTo,
				ProcessName: processRequest.ProcessName,
				Icon:        processRequest.StyleIcon,
				Style: fmt.Sprintf("width:%dpx;height:%dpx;line-height:30px;color:%s;left:%s;top:%s;",
					processRequest.StyleWidth, processRequest.StyleHeight, processRequest.StyleColor, process.PositionLeft, process.PositionTop),
			}
		}
	}

	jsplumbByte, err := json.Marshal(jsMap)

	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析数据错误", nil)
	}
	//更新流程图
	flow.Jsplumb = string(jsplumbByte)
	_, err = tx.Model(flow).Where("id=?", flow.ID).Update("jsplumb", flow.Jsplumb)
	_, err = tx.Model(flow).Where("id=?", flow.ID).Update("IsPublish", false)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
	}

	//更新步骤 流转条件 process_condition
	//根据ProcessCondition中的每一项分组，然后将每一组id相同的数据找出，将表达式合并为一个fmt.Sprintf("%s%s%s", condition.Field, condition.Operator, condition.Value)
	var conditionsMap map[int][]common.ProcessCondition
	if len(processRequest.ProcessCondition) > 0 {
		conditionsMap = groupConditionsById(processRequest.ProcessCondition)
	}
	//根据提交的conditionsMap更新process_var中的数据，如果有新数据，则新增，前提是只针对类型为int字段的数据，
	for _, conditions := range conditionsMap {
		for _, condition := range conditions {
			if condition.Field != "" {
				var exists_count int64
				facades.Orm().Query().Model(&models.ProcessVar{}).
					Where("flow_id=?", flow.ID).
					Where("process_id=?", id).
					Where("expression_field=?", condition.Field).Count(&exists_count)
				if exists_count == 0 {
					//新增一条
					var newProcessVar models.ProcessVar
					newProcessVar.FlowID = cast.ToInt(flow.ID)
					newProcessVar.ProcessID = id
					newProcessVar.ExpressionField = condition.Field
					facades.Orm().Query().Model(&models.ProcessVar{}).Create(&newProcessVar)

				}
			}
		}
	}

	for key, conditions := range conditionsMap {
		jsonStr, _ := json.Marshal(conditions)
		tx.Model(&models.Flowlink{}).Where("id=?", key).Update("expression", jsonStr)
	}

	//@袁浩：改，如果当前的processRequest.AutoPerson=="0",更新当前的步骤
	if processRequest.AutoPerson == "0" {
		tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("process_id=?", id).
			Where("type=?", "Condition").Update("auditor", processRequest.AutoPerson)
	}
	//权限处理
	if processRequest.AutoPerson != "0" {

		var fk models.Flowlink
		tx.Model(&fk).Where("flow_id=?", flow.ID).Where("process_id=?", id).Where("type=?", "Sys").First(&fk)
		if fk.ID != 0 {
			fk.Auditor = cast.ToString(processRequest.AutoPerson)
			_, err := tx.Model(&models.Flowlink{}).Where("id=?", fk.ID).Update("auditor", processRequest.AutoPerson)
			if err != nil {
				tx.Rollback()
				return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "数据错误", nil)
			}
		} else {
			tx.Model(&models.Flowlink{}).Create(&models.Flowlink{
				FlowID:        flow.ID,
				Type:          "Sys",
				ProcessID:     cast.ToUint(id),
				Auditor:       cast.ToString(processRequest.AutoPerson),
				NextProcessID: 0,
				Sort:          100,
			})
		}
		//更新当前flowlink的Audiitor

		//删除其他权限
		tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("process_id=?", id).
			Where("type!=?", "Condition").Where("type!=?", "Sys").Delete(&models.Flowlink{})
	} else {
		//指定部门
		if len(processRequest.RangeDeptIds) > 0 {
			var fkdept models.Flowlink
			tx.Model(&fkdept).Where("flow_id=?", flow.ID).Where("process_id=?", id).Where("type=?", "Dept").First(&fkdept)
			if fkdept.ID != 0 {
				//id组成的数组，然后转换为字符串
				auditor := ""
				for _, dept := range processRequest.RangeDeptIds {
					auditor += cast.ToString(dept) + ","
				}
				//取消最后一个,号
				auditor = strings.TrimSuffix(auditor, ",")
				fkdept.Auditor = auditor
				tx.Model(&models.Flowlink{}).Where("id=?", fkdept.ID).Update("auditor", fkdept.Auditor)
			} else {
				auditor := ""
				for _, dept := range processRequest.RangeDeptIds {
					auditor += cast.ToString(dept) + ","
				}
				tx.Model(&models.Flowlink{}).Create(&models.Flowlink{FlowID: flow.ID, Type: "Dept", ProcessID: cast.ToUint(id), Auditor: auditor, NextProcessID: 0, Sort: 100})
			}
		} else {
			//删除部门权限
			tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("process_id=?", id).
				Where("type=?", "Dept").Delete(&models.Flowlink{})
		}
		//	指定员工
		if len(processRequest.RangeEmpIds) > 0 {
			var fkemp models.Flowlink
			tx.Model(&fkemp).Where("flow_id=?", flow.ID).Where("process_id=?", id).Where("type=?", "Emp").First(&fkemp)
			if fkemp.ID != 0 {
				//id组成的数组，然后转换为字符串
				auditor := ""
				for _, emp := range processRequest.RangeEmpIds {
					auditor += cast.ToString(emp) + ","
				}
				auditor = strings.TrimSuffix(auditor, ",")
				fkemp.Auditor = auditor
				tx.Model(&models.Flowlink{}).Where("id=?", fkemp.ID).Update("auditor", fkemp.Auditor)
			} else {
				auditor := ""
				for _, emp := range processRequest.RangeEmpIds {
					auditor += cast.ToString(emp) + ","
				}
				auditor = strings.TrimSuffix(auditor, ",")
				tx.Model(&models.Flowlink{}).Create(&models.Flowlink{FlowID: flow.ID, Type: "Emp", ProcessID: cast.ToUint(id), Auditor: auditor, NextProcessID: 0, Sort: 100})
			}
		} else {
			//	删除
			tx.Model(&models.Flowlink{}).Where("flow_id=?", flow.ID).Where("process_id=?", id).Where("type=?", "Emp").Delete(&models.Flowlink{})
		}
	}
	tx.Commit()
	return httpfacades.NewResult(ctx).Success("保存成功", nil)
}
func groupConditionsById(conditions []common.ProcessCondition) map[int][]common.ProcessCondition {
	grouped := make(map[int][]common.ProcessCondition)
	for _, condition := range conditions {
		grouped[condition.Id] = append(grouped[condition.Id], condition)
	}
	return grouped
}

func (r *ProcessController) Destroy(ctx http.Context) http.Response {
	id := ctx.Request().RouteInt("id")
	flow_id := ctx.Request().InputInt("flow_id")
	var flow models.Flow
	tx, _ := facades.Orm().Query().Begin()
	tx.Model(&flow).Where("id=?", flow_id).First(&flow)
	tx.Model(&models.Flowlink{}).Where("flow_id=?", id).Where("id=?", id).Delete(&models.Flowlink{})
	tx.Model(&models.Flowlink{}).Where("flow_id=?", id).Where("next_process_id=?", id).Update("next_process_id", -1)
	tx.Model(&models.Process{}).Where("id=?", id).Delete(&models.Process{})
	jsMap := common.Plumb{}
	//flow.Jsplum解析为jsMap
	err := json.Unmarshal([]byte(flow.Jsplumb), &jsMap)
	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析数据错误", nil)
	}

	//需要将jsMap读取出来，然后再写回去
	for key, _ := range jsMap.List {
		if key == cast.ToString(id) {
			//	删除
			delete(jsMap.List, key)
		}
	}

	jsplumbByte, err := json.Marshal(jsMap)

	if err != nil {
		tx.Rollback()
		return httpfacades.NewResult(ctx).Error(http.StatusInternalServerError, "解析数据错误", nil)
	}
	//更新流程图
	flow.Jsplumb = string(jsplumbByte)
	tx.Model(&models.Flow{}).Where("id=?", flow.ID).Save(&flow)
	tx.Commit()
	return httpfacades.NewResult(ctx).Success("删除成功", nil)
}

func (r *ProcessController) Attribute(ctx http.Context) http.Response {
	id := ctx.Request().QueryInt("id")
	process := models.Process{}
	tx := facades.Orm().Query()
	tx.Model(&models.Process{}).Where("id=?", id).First(&process)

	//1- //当前步骤的下一步操作
	next_process := []models.Flowlink{}
	tx.Model(&models.Flowlink{}).Where("process_id=?", process.ID).
		Where("flow_id=?", process.FlowID).Where("type=?", "Condition").With("Process").
		With("NextProcess").Find(&next_process)
	next_process_ids := []int{}
	tx.Model(&models.Flowlink{}).Where("process_id=?", process.ID).
		Where("flow_id=?", process.FlowID).Where("type=?", "Condition").With("Process").
		With("NextProcess").Pluck("next_process_id", &next_process_ids)
	beixuan_process := []models.Flowlink{}
	tx.Model(&models.Flowlink{}).Where("flow_id=?", process.FlowID).
		Where("type=?", "Condition").Where("process_id !=?", process.ID).
		Where("process_id not in (?)", next_process_ids).With("Process").With("NextProcess").Find(&beixuan_process)

	//	2-流程模板 表单字段
	flow := models.Flow{}

	fields := []models.TemplateForm{}
	tx.Model(&models.Flow{}).Where("id=?", process.FlowID).With("Template").First(&flow)
	if flow.Template.ID != 0 {
		tfId := flow.Template.ID
		tx.Model(&models.TemplateForm{}).Where("template_id=?", tfId).Find(&fields)
	}

	//3-当前选择员工
	select_emps := []models.Emp{}
	auditor_emp_flowlink := models.Flowlink{}
	tx.Model(&models.Flowlink{}).Where("process_id=?", process.ID).
		Where("type=?", "Emp").Select("auditor").First(&auditor_emp_flowlink)
	//depts按照,拆分
	empsSlice := []string{}
	for _, emp := range strings.Split(auditor_emp_flowlink.Auditor, ",") {
		empsSlice = append(empsSlice, emp)
	}
	tx.Model(&models.Emp{}).Where("id in (?)", empsSlice).Find(&select_emps)
	//4 -flowlinks
	flowlink := models.Flowlink{}
	sys := "0"
	tx.Model(&models.Flowlink{}).Where("process_id = ?", process.ID).Where("flow_id=?", process.FlowID).
		Where("type=?", "Sys").First(&flowlink)
	if flowlink.Auditor != "" {
		sys = flowlink.Auditor
	}

	// 5-部门
	select_depts := []models.Dept{}
	auditor_dept_flowlink := models.Flowlink{}
	tx.Model(&models.Flowlink{}).Where("type=?", "Dept").Where("process_id=?", process.ID).
		Select("auditor").First(&auditor_dept_flowlink)
	//depts按照,拆分
	deptsSlice := []string{}
	for _, dept := range strings.Split(auditor_dept_flowlink.Auditor, ",") {
		deptsSlice = append(deptsSlice, dept)
	}
	tx.Model(&models.Dept{}).Where("id in (?)", deptsSlice).Find(&select_depts)

	// 6-flow
	flows := []models.Flow{}
	tx.Model(&models.Flow{}).Where("is_publish=?", 1).Where("id!=?", process.FlowID).Find(&flows)

	processes := []models.Process{}
	tx.Model(&models.Process{}).Where("flow_id=?", process.FlowID).Find(&processes)
	var count int64
	var can_child bool
	tx.Model(&models.Flowlink{}).Where("process_id=?", process.ID).Where("type=?", "Condition").
		Count(&count)
	if count == 1 {
		can_child = true
	}
	return httpfacades.NewResult(ctx).Success("", http.Json{
		"process":         process,
		"next_process":    next_process,
		"beixuan_process": beixuan_process,
		"fields":          fields,
		"select_emps":     select_emps,
		"sys":             sys,
		"select_depts":    select_depts,
		"flows":           flows,
		"processes":       processes,
		"can_child":       can_child,
	})
}

func (r *ProcessController) Condition(ctx http.Context) http.Response {
	flow_id := ctx.Request().InputInt("flow_id")
	process_id := ctx.Request().InputInt("process_id")
	next_process_id := ctx.Request().InputInt("next_process_id")
	//当前流程
	flowlink := models.Flowlink{}
	tx, _ := facades.Orm().Query().Begin()
	tx.Model(&models.Flowlink{}).Where("process_id=?", process_id).Where("next_process_id=?", next_process_id).
		Where("flow_id=?", flow_id).Where("type=?", "Condition").FirstOrFail(&flowlink)
	flow := models.Flow{}
	tx.Model(&models.Flow{}).With("Template.TemplateForms").Where("id=?", flow_id).First(&flow)
	//$day > 3  AND
	// $sex == 女
	fieldsArr := []string{}
	for _, form := range flow.Template.TemplateForms {
		//form.field form.field_name
		cleanedExpression := strings.Replace(flowlink.Expression, "$", "", -1)

		if strings.Contains(cleanedExpression, form.Field) {
			// 新建一个replaceStr
			replaceStr := strings.Replace(cleanedExpression, form.Field, form.FieldName, -1)
			fieldsArr = append(fieldsArr, replaceStr)
		}
	}
	res := make(map[int]interface{})
	if len(fieldsArr) > 0 {
		res[flowlink.NextProcessID] = map[string]interface{}{
			"desc":   fieldsArr[0],
			"option": "",
		}
	} else {
		res[flowlink.NextProcessID] = map[string]interface{}{
			"desc":   []string{},
			"option": "",
		}
	}

	return httpfacades.NewResult(ctx).Success("", res)
}
