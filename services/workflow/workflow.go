package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/hulutech-web/goravel-workflow/controllers/common"
	"github.com/hulutech-web/goravel-workflow/models"
	"github.com/spf13/cast"
	"reflect"
	"strings"
	"sync"
)

type Workflow struct {
	hooks map[string][]reflect.Value // 修改为 存储多个钩子函数
	mutex sync.Mutex
}

// Singleton 是 Workflow 的单例实例
var (
	baseWorkflowInstance *Workflow
	once                 sync.Once
)

// NewBaseWorkflow 单例工厂方法
func NewBaseWorkflow() *Workflow {
	once.Do(func() {
		baseWorkflowInstance = &Workflow{
			hooks: make(map[string][]reflect.Value),
		}
	})
	return baseWorkflowInstance
}

// RegisterHook 注册钩子方法
func (w *Workflow) RegisterHook(name string, method reflect.Value) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.hooks[name] = append(w.hooks[name], method)
	fmt.Printf("Registered hook: %s\n", name)
}

// NotifyStartThis 调用 NotifyStartThis 钩子
func (w *Workflow) NotifyStartThis(id uint) error {
	fmt.Printf("BaseWorkflow.NotifyStartThis :%d\n", id)

	return w.invokeHooks("NotifyStartThis", id)

}

// NotifyNextAuditor 调用 NotifyNextAuditor 钩子
func (w *Workflow) NotifyNextAuditor(id uint) error {
	fmt.Printf("BaseWorkflow.NotifyNextAuditor:%d\n", id)

	return w.invokeHooks("NotifyNextAuditor", id)

}

// invokeHooks 用于依次调用所有注册的钩子方法
func (w *Workflow) invokeHooks(hookName string, id uint) error {
	if hooks, ok := w.hooks[hookName]; ok {
		for _, hook := range hooks {
			w.mutex.Lock()
			defer w.mutex.Unlock()
			// 检查方法签名
			methodType := hook.Type()
			if methodType.NumIn() == 1 && methodType.In(0).Kind() == reflect.Uint {
				fmt.Printf("Calling %s...\n", hookName)
				hook.Call([]reflect.Value{reflect.ValueOf(id)})
				fmt.Printf("%s completed.\n", hookName)
			} else {
				return errors.New("Method signature mismatch or invalid hook for " + hookName)
			}
		}
	} else {
		return errors.New("Hook " + hookName + " not found.")
	}
	return nil
}

func (w *Workflow) SetFirstProcessAuditor(entry models.Entry, flowlink models.Flowlink) error {
	return facades.Orm().Transaction(func(tx orm.Transaction) error {
		var myFlowlink models.Flowlink

		var auditor_ids []int
		err := tx.Model(&models.Flowlink{}).Where("type != ?", "Condition").
			Where("process_id=?", flowlink.ProcessID).First(&myFlowlink)

		var process_id int
		var process_name string
		if myFlowlink.ID == 0 {

			//第一步未指定审核人 自动进入下一步操作
			var proc models.Proc
			proc.FlowID = cast.ToInt(entry.FlowID)
			proc.ProcessID = cast.ToInt(flowlink.ProcessID)
			proc.ProcessName = flowlink.Process.ProcessName
			proc.EmpID = cast.ToInt(entry.EmpID)
			proc.EmpName = entry.Emp.Name
			proc.DeptName = entry.Emp.Dept.DeptName
			proc.AuditorID = cast.ToInt(entry.EmpID)
			proc.AuditorName = entry.Emp.Name
			proc.AuditorDept = entry.Emp.Dept.DeptName
			proc.Status = 9
			proc.Circle = entry.Circle
			proc.Concurrence = carbon.NewDateTime(carbon.Now())
			proc.EntryID = entry.ID
			err = facades.Orm().Transaction(func(tx2 orm.Transaction) error {
				err = tx2.Model(&models.Proc{}).Create(&proc)
				return err
			})

			auditor_ids = w.GetProcessAuditorIds(entry, flowlink.NextProcessID)
			process_id = flowlink.NextProcessID
			process_name = flowlink.NextProcess.ProcessName
			entry.ProcessID = cast.ToUint(flowlink.NextProcessID)
		} else {

			auditor_ids = w.GetProcessAuditorIds(entry, cast.ToInt(flowlink.NextProcessID))
			process_id = cast.ToInt(flowlink.ProcessID)
			process_name = flowlink.Process.ProcessName
			entry.ProcessID = cast.ToUint(flowlink.ProcessID)

		}
		//步骤流转
		//步骤审核人
		var auditors_emps []models.Emp
		err = facades.Orm().Transaction(func(tx4 orm.Transaction) error {
			tx4.Model(&models.Emp{}).Where("id IN (?)", auditor_ids).With("Dept").Find(&auditors_emps)
			if len(auditors_emps) < 1 {
				return errors.New("下一步骤未找到审批人")
			}
			return nil
		})
		for _, emp := range auditors_emps {
			var proc2 models.Proc
			proc2.EntryID = entry.ID
			proc2.FlowID = cast.ToInt(entry.FlowID)
			proc2.ProcessID = process_id
			proc2.ProcessName = process_name
			proc2.EmpID = cast.ToInt(emp.ID)
			proc2.EmpName = emp.Name
			proc2.DeptName = emp.Dept.DeptName
			proc2.Status = 0
			proc2.Circle = entry.Circle
			proc2.Concurrence = carbon.NewDateTime(carbon.Now())
			facades.Orm().Query().Model(&models.Proc{}).Create(&proc2)
		}

		facades.Orm().Query().Model(models.Entry{}).Where("id=?", entry.ID).Save(&entry)
		return nil
	})
}

func (w *Workflow) GetProcessAuditorIds(entry models.Entry, next_process_id int) []int {
	var auditor_ids []int
	var flowlink models.Flowlink
	query := facades.Orm().Query()
	query.Model(&models.Flowlink{}).Where("type = ?", "Sys").Where("process_id=?", next_process_id).First(&flowlink)
	if flowlink.ID > 0 {
		if flowlink.Auditor == "-1000" {
			//发起人
			auditor_ids = append(auditor_ids, cast.ToInt(entry.EmpID))
		}
		if flowlink.Auditor == "-1001" {
			//发起人部门主管
			if entry.Emp.Dept.ID == 0 {
				return auditor_ids
			}
			auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.DirectorID))
		}
		if flowlink.Auditor == "-1002" {
			//发起人部门经理
			if entry.Emp.Dept.ID == 0 {
				return auditor_ids
			}
			auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.ManagerID))
		}
	} else {
		//	concurrent 并行
		//	1、指定员工
		concurrent_emp_flowlink := models.Flowlink{}
		query.Model(&models.Flowlink{}).Where("type = ?", "Emp").Where("process_id=?", next_process_id).First(&concurrent_emp_flowlink)
		if concurrent_emp_flowlink.ID > 0 {
			Auditor_ids := []string{}
			//按照,分割concurrent_flowlink.Auditor
			Auditor_ids = strings.Split(concurrent_emp_flowlink.Auditor, ",")
			for _, id := range Auditor_ids {
				auditor_ids = append(auditor_ids, cast.ToInt(id))
			}
		}
		//	2、指定部门
		concurrent_dept_flowlink := models.Flowlink{}
		query.Model(&models.Flowlink{}).Where("type = ?", "Emp").Where("process_id=?", next_process_id).
			First(&concurrent_dept_flowlink)

		if concurrent_dept_flowlink.ID > 0 {
			dept_id_strs := []string{}
			//按照,分割concurrent_flowlink.Auditor
			dept_id_strs = strings.Split(concurrent_emp_flowlink.Auditor, ",")
			dept_ids := []int{}
			for _, id := range dept_id_strs {
				dept_ids = append(dept_ids, cast.ToInt(id))
			}
			emp_ids := []int{}
			query.Model(&models.Emp{}).Where("dept_id IN (?)", dept_ids).Pluck("id", &emp_ids)
			for _, id := range emp_ids {
				auditor_ids = append(auditor_ids, id)
			}
		}
		//	3、指定角色，待完善
	}
	ret_auditor_ids := uniqueSlice(auditor_ids)
	//	对auditor_ids去重
	return ret_auditor_ids

}

// 辅助函数，从slice中去重
func uniqueSlice(slice []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, value := range slice {
		if _, ok := seen[value]; !ok {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}

// 流转
func (w *Workflow) Transfer(process_id int, user models.Emp, content string) error {
	tx, _ := facades.Orm().Query().Begin()
	var emp models.Emp
	facades.Orm().Query().Model(&models.Emp{}).With("Dept").Where("user_id=?", user.ID).First(&emp)
	var proc models.Proc
	tx.Model(&models.Proc{}).With("Entry.Emp.Dept").Where("process_id=?", process_id).
		Where("emp_id=?", emp.ID).Where("status=?", 0).First(&proc)
	if proc.ID == 0 {
		return errors.New("未绑定员工，请设置员工绑定")
	}
	var fkcount int64
	tx.Model(&models.Flowlink{}).Where("process_id=?", proc.ProcessID).Where("type=?", "Condition").Count(&fkcount)

	if fkcount > 1 {
		//	情况一：有条件
		pvar := models.ProcessVar{}
		tx.Model(&models.ProcessVar{}).Where("process_id=?", process_id).First(&pvar)
		var field_value string
		tx.Model(&models.EntryData{}).Select("field_value").
			Where("entry_id=?", proc.EntryID).
			Where("field_name=?", pvar.ExpressionField).Pluck("field_value", &field_value)

		flowlinks := []models.Flowlink{}
		tx.Model(&models.Flowlink{}).Where("process_id=?", proc.ProcessID).
			Where("type=?", "Condition").Find(&flowlinks)
		var flowlink models.Flowlink //满足条件的flowlink
		field := pvar.ExpressionField
		for _, m := range flowlinks {
			if m.Expression == "" {
				return errors.New("未设置流转条件，无法流转，请联系流程设置人员")
			}

			if m.Expression == "1" {
				flowlink = m
				break
			} else {
				//m.Expression
				type ResultCount struct {
					Number int `json:"number"`
				}
				var resultCount ResultCount
				processConditions := []common.ProcessCondition{}
				json.Unmarshal([]byte(m.Expression), &processConditions)
				if len(processConditions) > 0 {
					//检查语法错误(使用mysql数条件表达式
					conditionSql := ""
					for _, condition := range processConditions {
						if condition.Field != field {
							return errors.New("没有该条件字段，请检查")
						} else {
							conditionSql += fmt.Sprintf(" `field_value` %s %s %s", condition.Operator, condition.Value, condition.Extra)
						}
					}
					conditionSql = fmt.Sprintf("SELECT count(*) as number FROM entrydatas WHERE entry_id=%d and flow_id=%d and (%s) and (`field_name`='%s')",
						proc.EntryID, proc.FlowID, conditionSql, field)
					//还需要条件entry_id和flow_id
					err := facades.Orm().Query().Raw(conditionSql).Scan(&resultCount)
					if err != nil {
						return errors.New("条件语法错误，请检查")
					}
					if resultCount.Number > 0 {
						flowlink = m
						break
					}
				}
			}
		}
		if flowlink.ID == 0 {
			return errors.New("未找到符合条件的流转条件，无法流转")
		}
		var withFlowlink models.Flowlink
		facades.Orm().Query().Model(&models.Flowlink{}).With("NextProcess").Where("id=?", flowlink.ID).First(&withFlowlink)
		auditor_ids := w.GetProcessAuditorIds(proc.Entry, withFlowlink.NextProcessID)
		if len(auditor_ids) == 0 {
			return errors.New("未找到下一步骤审批人")
		}
		auditors := []models.Emp{}
		tx.Model(&models.Emp{}).Where("id IN (?)", auditor_ids).With("Dept").Find(&auditors)
		if len(auditors) == 0 {
			return errors.New("未找到下一步骤审批人")
		}
		curr_time := carbon.NewDateTime(carbon.Now())
		for _, auditor := range auditors {
			tx.Model(&models.Proc{}).Create(&models.Proc{
				EntryID:     proc.EntryID,
				FlowID:      cast.ToInt(proc.FlowID),
				ProcessID:   withFlowlink.NextProcessID,
				ProcessName: withFlowlink.NextProcess.ProcessName,
				EmpID:       cast.ToInt(auditor.ID),
				EmpName:     auditor.Name,
				DeptName:    auditor.Dept.DeptName,
				Circle:      proc.Entry.Circle,
				Status:      0,
				IsRead:      0,
				Concurrence: curr_time,
			})
			//通知下一个审批人
			//通知发起人，被驳回
			baseWorkflowInstance.NotifyNextAuditor(auditor.ID)
		}
		procEntry := models.Entry{}
		tx.Model(&models.Entry{}).Where("id=?", proc.EntryID).FirstOrFail(&procEntry)
		procEntry.ProcessID = cast.ToUint(flowlink.NextProcessID)
		tx.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)
		//判断是否存在父进程
		if proc.Entry.Pid > 0 {
			proc2Entry := models.Entry{}
			tx.Model(&models.Entry{}).Where("id=?", proc.EntryID).FirstOrFail(&proc2Entry)
			partentEntry := models.Entry{}
			tx.Model(&models.Entry{}).Where("pid=?", proc.ID).FirstOrFail(&partentEntry)
			partentEntry.Child = flowlink.NextProcessID
			tx.Model(&models.Entry{}).Where("id=?", partentEntry.ID).Save(&partentEntry)
		}
	} else {
		fklink := models.Flowlink{}
		tx.Model(&models.Flowlink{}).With("Process").With("NextProcess").Where("process_id=?", proc.ProcessID).
			Where("type=?", "Condition").First(&fklink)
		if fklink.Process.ChildFlowID > 0 {
			// 创建子流程
			child_entry := models.Entry{}
			tx.Model(&models.Entry{}).
				Where("pid=?", proc.Entry.ID).
				Where("circle=?", proc.Entry.Circle).FirstOrFail(&child_entry)
			if child_entry.ID == 0 {
				newChildEntry := models.Entry{}
				newChildEntry.Title = proc.Entry.Title
				newChildEntry.FlowID = cast.ToUint(fklink.Process.ChildFlowID)
				newChildEntry.EmpID = cast.ToUint(emp.ID)
				newChildEntry.Status = 0
				newChildEntry.Pid = cast.ToInt(proc.Entry.ID)
				newChildEntry.Circle = proc.Entry.Circle
				newChildEntry.EnterProcessID = cast.ToInt(fklink.ProcessID)
				newChildEntry.EnterProcID = cast.ToInt(proc.ID)
				tx.Model(&models.Entry{}).Create(&newChildEntry)
			}

			child_flowlink := models.Flowlink{}
			tx.Model(&models.Flowlink{}).
				Where("flow_id=?", fklink.Process.ChildFlowID).With("Process").
				Where("type=?", "Condition").Where("Process", func(query orm.Query) orm.Query {
				return query.Where("position=?", 0)
			}).Order("sort ASC").First(&child_flowlink)
			err := w.SetFirstProcessAuditor(child_entry, child_flowlink)
			if err != nil {
				return err
			}
			child_parent := models.Entry{}
			tx.Model(&models.Entry{}).Where("id=?", child_entry.Pid).First(&child_parent)
			child_parent.Child = cast.ToInt(child_entry.ProcessID)
			tx.Model(&models.Entry{}).Where("id=?", child_parent.ID).Save(&child_parent)
		} else {
			if fklink.NextProcessID == -1 {
				//最后一步
				tx.Model(&models.Entry{}).Where("id=?", proc.EntryID).Update(models.Entry{
					Status:    9,
					ProcessID: fklink.ProcessID,
				})

				if proc.Entry.Pid > 0 {
					if proc.Entry.EnterProcess.ChildAfter == 1 {
						//同时结束父流程
						parentEntry := models.Entry{}
						tx.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)
						map_entry := make(map[string]interface{})
						map_entry["status"] = 9
						map_entry["child"] = 0
						tx.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Update(&map_entry)
					} else {
						//	进入设置的父流程步骤
						if proc.Entry.EnterProcess.ChildBackProcess > 0 {
							w.goToProcess(*proc.Entry.ParentEntry, proc.Entry.EnterProcess.ChildBackProcess)
							proc.Entry.ParentEntry.ProcessID = cast.ToUint(proc.Entry.EnterProcess.ChildBackProcess)
						} else {
							//默认进入父流程步骤下一步
							parentFlowlink := models.Flowlink{}
							tx.Model(&models.Flowlink{}).Where("process_id=?", proc.Entry.EnterProcessID).Where("type=?", "Condition").First(&parentFlowlink)
							if parentFlowlink.NextProcessID == -1 {
								parentEntry := models.Entry{}
								tx.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)
								map_entry := make(map[string]interface{})
								map_entry["process_id"] = cast.ToUint(proc.Entry.EnterProcess.ChildBackProcess)
								map_entry["status"] = 9
								map_entry["child"] = 0
								tx.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Update(&map_entry)

								var notifyProc models.Proc
								tx.Model(&models.Proc{}).Where("id=?", proc.ID).FirstOrFail(&notifyProc)
							} else {
								w.goToProcess(*proc.Entry.ParentEntry, parentFlowlink.NextProcessID)
								proc.Entry.ParentEntry.ProcessID = cast.ToUint(parentFlowlink.NextProcessID)
								parentEntry := models.Entry{}
								tx.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)
								map_entry := make(map[string]interface{})
								map_entry["process_id"] = parentFlowlink.NextProcessID
								map_entry["status"] = 0
								tx.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Update(&map_entry)

							}
						}
						pentry := models.Entry{}
						tx.Model(&models.Entry{}).Where("id=?", proc.Entry.ParentEntry.ID).First(&pentry)
						map_entry := make(map[string]interface{})
						map_entry["child"] = 0
						tx.Model(&models.Entry{}).Where("id=?", pentry.ID).Save(&map_entry)

					}
				} else {
					var notifyProc models.Proc
					tx.Model(&models.Proc{}).Where("id=?", proc.ID).FirstOrFail(&notifyProc)
				}
				w.NotifyStartThis(proc.EntryID)

			} else {
				auditor_ids := w.GetProcessAuditorIds(proc.Entry, fklink.NextProcessID)
				auditors := []models.Emp{}
				tx.Model(&models.Emp{}).Where("id in (?)", auditor_ids).With("Dept").Find(&auditors)
				if len(auditors) < 1 {
					return errors.New("未找到下一步步骤审批人")
				}
				for _, auditor := range auditors {
					tx.Model(&models.Proc{}).Create(&models.Proc{
						EntryID:     proc.Entry.ID,
						FlowID:      cast.ToInt(proc.FlowID),
						ProcessID:   cast.ToInt(fklink.NextProcessID),
						ProcessName: fklink.NextProcess.ProcessName,
						EmpID:       cast.ToInt(auditor.ID),
						EmpName:     auditor.Name,
						Content:     content,
						DeptName:    auditor.Dept.DeptName,
						Circle:      proc.Entry.Circle,
						Concurrence: carbon.NewDateTime(carbon.Now()),
						Status:      0,
						IsRead:      0,
					})
					//通知下一个审批人
					w.NotifyNextAuditor(auditor.ID)
				}
				tx.Model(&models.Entry{}).Where("id=?", proc.Entry.ID).Update("process_id", cast.ToUint(fklink.NextProcessID))
				//	判断是否存在父进程
				var parentEntry models.Entry
				tx.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).FirstOrFail(&parentEntry)
				if parentEntry.Pid > 0 {
					parentEntry.Child = cast.ToInt(fklink.NextProcessID)
					tx.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
				}
			}
		}
	}

	tx.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("process_id=?", proc.ProcessID).
		Where("circle=?", proc.Entry.Circle).
		Where("status=?", 0).Update(models.Proc{
		Status:      1,
		AuditorID:   cast.ToInt(emp.ID),
		AuditorName: emp.Name,
		DeptName:    emp.Dept.DeptName,
		Content:     content,
		Concurrence: carbon.NewDateTime(carbon.Now()),
	})

	tx.Commit()

	return nil
}

func (w *Workflow) goToProcess(entry models.Entry, processID int) error {
	auditor_ids := w.GetProcessAuditorIds(entry, processID)
	auditors := []models.Emp{}
	err := facades.Orm().Query().Model(&models.Emp{}).With("Dept").Where("id in (?)", auditor_ids).Find(&auditors)
	if err != nil {
		return err
	}
	if len(auditors) < 1 {
		return errors.New("未找到下一步步骤审批人")
	}
	current_time := carbon.NewDateTime(carbon.Now())
	processName := ""
	err = facades.Orm().Query().Model(&models.Process{}).Where("id=?", processID).Select("ProcessName").
		Scan(&processName)
	if err != nil {
		return err
	}
	for _, auditor := range auditors {
		err = facades.Orm().Query().Model(&models.Proc{}).Create(&models.Proc{
			EntryID:     entry.ID,
			FlowID:      cast.ToInt(entry.FlowID),
			ProcessID:   cast.ToInt(processID),
			ProcessName: processName,
			EmpID:       cast.ToInt(auditor.ID),
			EmpName:     auditor.Name,
			DeptName:    auditor.Dept.DeptName,
			Circle:      entry.Circle,
			Status:      0,
			IsRead:      0,
			Concurrence: current_time,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Workflow) Pass(process_id int, user models.Emp, content string) error {
	return w.Transfer(process_id, user, content)
}

func (w *Workflow) UnPass(proc_id int, user models.Emp, content string) {
	var proc models.Proc
	query := facades.Orm().Query()
	var emp models.Emp
	query.Model(&models.Emp{}).Where("user_id=?", user.ID).First(&emp)
	query.Model(&models.Proc{}).Where("id=?", proc_id).With("Entry").First(&proc)
	todoProc := models.Proc{}
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("process_id=?", proc.ProcessID).
		Where("circle=?", proc.Entry.Circle).
		Where("status=?", 0).First(&todoProc)
	todoProc.Status = 1
	todoProc.Beizhu = "审批人不同意"
	todoProc.AuditorID = cast.ToInt(emp.ID)
	todoProc.AuditorName = user.Name
	todoProc.AuditorDept = user.Dept.DeptName
	todoProc.Concurrence = carbon.NewDateTime(carbon.Now())
	todoProc.Content = content
	todoProc.IsRead = 1
	todoProc.Status = -1
	query.Model(&models.Proc{}).Where("id=?", todoProc.ID).Save(&todoProc)
	query.Model(&models.Entry{}).Where("id=?", proc.EntryID).Update("status", -1)
	if proc.Entry.Pid > 0 {
		var parentEntry models.Entry
		query.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).FirstOrFail(&parentEntry)
		parentEntry.Child = proc.ProcessID
		parentEntry.Status = -1
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
	}
	w.NotifyStartThis(proc.Entry.EmpID)

}
