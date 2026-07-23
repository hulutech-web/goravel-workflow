package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"reflect"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/hulutech-web/goravel-workflow/controllers/common"
	"github.com/hulutech-web/goravel-workflow/models"
	"github.com/hulutech-web/goravel-workflow/services/workflow/official_plugins"
	"github.com/spf13/cast"
)

// Workflow concurrency type constants
const (
	ConcTypeSequential = 0 // 依次审批（默认）
	ConcTypeConsensus  = 1 // 会签：所有人通过才进入下一步
	ConcTypeAny        = 2 // 或签：一人通过即进入下一步，其余跳过
)

// Flowlink auditor special values
const (
	AuditorInitiator   = -1000 // 发起人
	AuditorDirector    = -1001 // 部门主管
	AuditorManager     = -1002 // 部门经理
	AuditorFormField   = -1003 // 从表单字段读取审批人
	AuditorDynamicExpr = -1004 // 动态表达式计算审批人
)

type Workflow struct {
	hooks map[string][]reflect.Value
	mutex sync.RWMutex
}

// Singleton is the Workflow singleton instance
var (
	baseWorkflowInstance *Workflow
	once                 sync.Once
)

// NewBaseWorkflow returns the singleton Workflow instance
func NewBaseWorkflow() *Workflow {
	once.Do(func() {
		baseWorkflowInstance = &Workflow{
			hooks: make(map[string][]reflect.Value),
		}
	})
	return baseWorkflowInstance
}

// RegisterHook registers a hook function by name
func (w *Workflow) RegisterHook(name string, method reflect.Value) {
	if w.hooks == nil {
		w.hooks = make(map[string][]reflect.Value)
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.hooks[name] = append(w.hooks[name], method)
}

// NotifySendOne calls the NotifySendOne hook
func (w *Workflow) NotifySendOne(id uint) error {
	if w == nil {
		return fmt.Errorf("workflow instance is nil")
	}
	w.invokeHooks("NotifySendOneHook", id)
	return nil
}

// NotifyNextAuditor calls the NotifyNextAuditor hook
func (w *Workflow) NotifyNextAuditor(id uint) error {
	if w == nil {
		return fmt.Errorf("workflow instance is nil")
	}
	w.invokeHooks("NotifyNextAuditorHook", id)
	return nil
}

// invokeHooks calls all registered hooks for the given name.
// Uses RWMutex for thread-safe concurrent reads.
func (w *Workflow) invokeHooks(hookName string, id uint) {
	w.mutex.RLock()
	hooks, ok := w.hooks[hookName]
	w.mutex.RUnlock()

	if !ok {
		fmt.Printf("Hook %s not found.\n", hookName)
		return
	}

	for _, hook := range hooks {
		methodType := hook.Type()
		if methodType.NumIn() == 1 && methodType.In(0).Kind() == reflect.Uint {
			fmt.Printf("Calling %s...\n", hookName)
			hook.Call([]reflect.Value{reflect.ValueOf(id)})
			fmt.Printf("%s completed.\n", hookName)
		} else {
			fmt.Printf("Method signature mismatch or invalid hook for %s.\n", hookName)
		}
	}
}

// skipRemainingConcurrentProcs marks all remaining pending procs as skipped in an any-sign step
func (w *Workflow) skipRemainingConcurrentProcs(query orm.Query, entryID uint, processID int, circle int) error {
	var pendingProcs []models.Proc
	if err := query.Model(&models.Proc{}).
		Where("entry_id=?", entryID).
		Where("process_id=?", processID).
		Where("circle=?", circle).
		Where("status=?", models.ProcStatusPending).
		Find(&pendingProcs); err != nil {
		return err
	}
	for i := range pendingProcs {
		pendingProcs[i].Status = models.ProcStatusSkipped
		if err := query.Model(&models.Proc{}).Where("id=?", pendingProcs[i].ID).Save(&pendingProcs[i]); err != nil {
			return err
		}
	}
	return nil
}

// checkConsensusComplete checks if all consensus approvers have finished (approved or rejected)
func (w *Workflow) checkConsensusComplete(query orm.Query, entryID uint, processID int, circle int) (allDone bool, hasRejection bool, err error) {
	totalProcs, err := query.Model(&models.Proc{}).Where("entry_id=?", entryID).Where("process_id=?", processID).Where("circle=?", circle).Count()
	if err != nil {
		return false, false, err
	}

	approvedProcs, err := query.Model(&models.Proc{}).Where("entry_id=?", entryID).Where("process_id=?", processID).Where("circle=?", circle).Where("status=?", models.ProcStatusApproved).Count()
	if err != nil {
		return false, false, err
	}

	rejectedProcs, err := query.Model(&models.Proc{}).Where("entry_id=?", entryID).Where("process_id=?", processID).Where("circle=?", circle).Where("status=?", models.ProcStatusRejected).Count()
	if err != nil {
		return false, false, err
	}

	hasRejection = rejectedProcs > 0
	allDone = (approvedProcs + rejectedProcs) >= totalProcs
	return allDone, hasRejection, nil
}

// createProcsForProcess creates Proc records for all auditors at a given process step
func (w *Workflow) createProcsForProcess(query orm.Query, entry *models.Entry, processID int, processName string, auditorIDs []int) error {
	if len(auditorIDs) == 0 {
		return errors.New("未找到审批人")
	}

	now := carbon.NewDateTime(carbon.Now())
	for _, empID := range auditorIDs {
		var emp models.Emp
		if err := query.Model(&models.Emp{}).Where("id=?", empID).With("Dept").First(&emp); err != nil {
			continue
		}
		proc := models.Proc{
			EntryID:     entry.ID,
			FlowID:      cast.ToInt(entry.FlowID),
			ProcessID:   processID,
			ProcessName: processName,
			EmpID:       cast.ToInt(emp.ID),
			EmpName:     emp.Name,
			DeptName:    emp.Dept.DeptName,
			Circle:      entry.Circle,
			Status:      models.ProcStatusPending,
			IsRead:      0,
			Concurrence: now,
		}
		if err := query.Model(&models.Proc{}).Create(&proc); err != nil {
			return err
		}
		w.NotifyNextAuditor(emp.ID)
	}
	return nil
}

// SetFirstProcessAuditor initializes approval tasks for the first process step
func (w *Workflow) SetFirstProcessAuditor(entry models.Entry, flowlink models.Flowlink) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var myFlowlink models.Flowlink
		var auditor_ids []int

		err := tx.Model(&models.Flowlink{}).Where("type != ?", "Condition").
			Where("process_id=?", flowlink.ProcessID).First(&myFlowlink)
		if err != nil {
			return err
		}

		var process_id int
		var process_name string
		if myFlowlink.ID == 0 {
			// First step has no designated approver, auto-approve and move to next
			proc := models.Proc{
				EntryID:     entry.ID,
				FlowID:      cast.ToInt(entry.FlowID),
				ProcessID:   cast.ToInt(flowlink.ProcessID),
				ProcessName: flowlink.Process.ProcessName,
				EmpID:       cast.ToInt(entry.EmpID),
				EmpName:     entry.Emp.Name,
				DeptName:    entry.Emp.Dept.DeptName,
				AuditorID:   cast.ToInt(entry.EmpID),
				AuditorName: entry.Emp.Name,
				AuditorDept: entry.Emp.Dept.DeptName,
				Status:      models.ProcStatusConsensus,
				Circle:      entry.Circle,
				Concurrence: carbon.NewDateTime(carbon.Now()),
			}
			if err := tx.Model(&models.Proc{}).Create(&proc); err != nil {
				return err
			}

			auditor_ids = w.GetProcessAuditorIds(entry, flowlink.NextProcessID)
			process_id = flowlink.NextProcessID
			process_name = flowlink.NextProcess.ProcessName
			entry.ProcessID = cast.ToUint(flowlink.NextProcessID)
		} else {
			auditor_ids = w.GetProcessAuditorIds(entry, cast.ToInt(flowlink.ProcessID))
			process_id = cast.ToInt(flowlink.ProcessID)
			process_name = flowlink.Process.ProcessName
			entry.ProcessID = cast.ToUint(flowlink.ProcessID)
		}

		// Query auditors using the outer transaction instead of creating a nested one
		var auditors_emps []models.Emp
		if err := tx.Model(&models.Emp{}).Where("id IN (?)", auditor_ids).With("Dept").Find(&auditors_emps); err != nil {
			return err
		}
		if len(auditors_emps) < 1 {
			return errors.New("下一步骤未找到审批人")
		}

		for _, emp := range auditors_emps {
			proc2 := models.Proc{
				EntryID:     entry.ID,
				FlowID:      cast.ToInt(entry.FlowID),
				ProcessID:   process_id,
				ProcessName: process_name,
				EmpID:       cast.ToInt(emp.ID),
				EmpName:     emp.Name,
				DeptName:    emp.Dept.DeptName,
				Circle:      entry.Circle,
				Status:      models.ProcStatusPending,
				IsRead:      0,
				Concurrence: carbon.NewDateTime(carbon.Now()),
			}
			if err := tx.Model(&models.Proc{}).Create(&proc2); err != nil {
				return err
			}
		}

		_, err = tx.Model(models.Entry{}).Where("id=?", entry.ID).Update("process_id", entry.ProcessID)
		return err
	})
}

// GetProcessAuditorIds calculates approver IDs for a given process step
func (w *Workflow) GetProcessAuditorIds(entry models.Entry, next_process_id int) []int {
	var auditor_ids []int
	query := facades.Orm().Query()

	// Check Sys type first (special approver rules like initiator/director/manager)
	var sysFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).Where("type = ?", "Sys").Where("process_id=?", next_process_id).First(&sysFlowlink)

	if sysFlowlink.ID > 0 {
		switch sysFlowlink.Auditor {
		case "-1000":
			auditor_ids = append(auditor_ids, cast.ToInt(entry.EmpID))
		case "-1001":
			if entry.Emp.Dept.ID > 0 {
				auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.DirectorID))
			}
		case "-1002":
			if entry.Emp.Dept.ID > 0 {
				auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.ManagerID))
			}
		case "-1003":
			// Read approver ID from form field specified in approver_rule
			if sysFlowlink.ApproverRule != "" {
				var fieldValue string
				query.Model(&models.EntryData{}).Select("field_value").
					Where("entry_id=?", entry.ID).
					Where("field_name=?", sysFlowlink.ApproverRule).
					Pluck("field_value", &fieldValue)
				if fieldValue != "" {
					if id := cast.ToInt(fieldValue); id > 0 {
						auditor_ids = append(auditor_ids, id)
					}
				}
			}
		case "-1004":
			// Dynamic expression: use approver_rule as a mapping key
			if entry.Emp.Dept.ID > 0 {
				switch sysFlowlink.ApproverRule {
				case "director":
					auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.DirectorID))
				case "manager":
					auditor_ids = append(auditor_ids, cast.ToInt(entry.Emp.Dept.ManagerID))
				default:
					if id := cast.ToInt(sysFlowlink.ApproverRule); id > 0 {
						auditor_ids = append(auditor_ids, id)
					}
				}
			}
		default:
			if id := cast.ToInt(sysFlowlink.Auditor); id > 0 {
				auditor_ids = append(auditor_ids, id)
			}
		}
	} else {
		// Non-Sys type: Emp or Dept based assignment
		var empFlowlink models.Flowlink
		query.Model(&models.Flowlink{}).Where("type = ?", "Emp").Where("process_id=?", next_process_id).First(&empFlowlink)
		if empFlowlink.ID > 0 && empFlowlink.Auditor != "" {
			for _, idStr := range strings.Split(empFlowlink.Auditor, ",") {
				if id := cast.ToInt(strings.TrimSpace(idStr)); id > 0 {
					auditor_ids = append(auditor_ids, id)
				}
			}
		}

		var deptFlowlink models.Flowlink
		query.Model(&models.Flowlink{}).Where("type = ?", "Dept").Where("process_id=?", next_process_id).First(&deptFlowlink)
		if deptFlowlink.ID > 0 && deptFlowlink.Auditor != "" {
			var deptIDs []int
			for _, idStr := range strings.Split(deptFlowlink.Auditor, ",") {
				if id := cast.ToInt(strings.TrimSpace(idStr)); id > 0 {
					deptIDs = append(deptIDs, id)
				}
			}
			if len(deptIDs) > 0 {
				var directorIDs []int
				query.Model(&models.Dept{}).Select("director_id").Where("id IN (?)", deptIDs).Pluck("director_id", &directorIDs)
				auditor_ids = append(auditor_ids, directorIDs...)
			}
		}
	}

	return uniqueSlice(auditor_ids)
}

// uniqueSlice removes duplicates from an int slice while preserving order
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

// Transfer is the core workflow engine that handles approval routing
func (w *Workflow) Transfer(process_id int, user models.Emp, content string) error {
	query := facades.Orm().Query()

	// Resolve user to emp
	var emp models.Emp
	if err := query.Model(&models.Emp{}).With("Dept").Where("user_id=?", user.ID).First(&emp); err != nil {
		return err
	}

	// Find the current proc
	var proc models.Proc
	query.Model(&models.Proc{}).
		With("Entry.Emp.Dept").
		With("Entry.ParentEntry").
		Where("process_id=?", process_id).
		Where("emp_id=?", emp.ID).
		Where("status=?", models.ProcStatusPending).
		First(&proc)

	if proc.ID == 0 {
		return errors.New("未绑定员工，请设置员工绑定")
	}

	// Check if this process step has concurrency mode
	var currentFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).Where("process_id=?", proc.ProcessID).Where("type != ?", "Condition").First(&currentFlowlink)
	concurrencyType := currentFlowlink.ConcurrencyType

	// For consensus (会签): check if all approvers are done
	if concurrencyType == ConcTypeConsensus {
		allDone, hasRejection, err := w.checkConsensusComplete(query, proc.EntryID, proc.ProcessID, proc.Entry.Circle)
		if err != nil {
			return err
		}
		if allDone {
			if hasRejection {
				// One person rejected → reject the whole consensus step
				proc.Status = models.ProcStatusRejected
				proc.Content = content
				proc.AuditorID = cast.ToInt(emp.ID)
				proc.AuditorName = emp.Name
				query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(&proc)
				return w.handleRejectEntry(query, &proc)
			}
			// All approved → proceed to next step (fall through)
		} else {
			// Still waiting for other approvers — just mark this one approved
			proc.Status = models.ProcStatusApproved
			proc.Content = content
			proc.AuditorID = cast.ToInt(emp.ID)
			proc.AuditorName = emp.Name
			query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(&proc)
			return nil
		}
	}

	// For any-sign (或签): first approver wins, skip others
	if concurrencyType == ConcTypeAny {
		if err := w.skipRemainingConcurrentProcs(query, proc.EntryID, proc.ProcessID, proc.Entry.Circle); err != nil {
			return err
		}
	}

	// --- Normal transfer logic (sequential or after consensus/any-sign decision) ---

	// Check for conditional branches
	fkcount, err := query.Model(&models.Flowlink{}).Where("process_id=?", proc.ProcessID).Where("type=?", "Condition").Count()
	if err != nil {
		return err
	}

	if fkcount > 1 {
		return w.transferWithConditions(query, &proc, content, emp)
	}

	// No conditions — find the next flowlink
	var fklink models.Flowlink
	query.Model(&models.Flowlink{}).With("Process").With("NextProcess").
		Where("process_id=?", proc.ProcessID).Where("type != ?", "Condition").First(&fklink)

	if fklink.Process.ChildFlowID > 0 {
		return w.handleChildWorkflow(query, &proc, &fklink, content, emp)
	}

	if fklink.NextProcessID == -1 {
		return w.handleLastStep(query, &proc, &fklink, content, emp)
	}

	// Normal next step
	return w.handleNextStep(query, &proc, &fklink, content, emp)
}

// transferWithConditions handles conditional branch routing
func (w *Workflow) transferWithConditions(query orm.Query, proc *models.Proc, content string, emp models.Emp) error {
	pvar := models.ProcessVar{}
	if err := query.Model(&models.ProcessVar{}).Where("process_id=?", proc.ProcessID).First(&pvar); err != nil {
		return err
	}

	flowlinks := []models.Flowlink{}
	query.Model(&models.Flowlink{}).Where("process_id=?", proc.ProcessID).Where("type=?", "Condition").Order("sort ASC").Find(&flowlinks)

	var matchedFlowlink models.Flowlink
	field := pvar.ExpressionField

	for _, m := range flowlinks {
		if m.Expression == "" {
			return errors.New("未设置流转条件，无法流转")
		}
		if m.Expression == "1" {
			matchedFlowlink = m
			break
		}

		processConditions := []common.ProcessCondition{}
		if err := json.Unmarshal([]byte(m.Expression), &processConditions); err != nil {
			continue
		}
		if len(processConditions) == 0 {
			continue
		}

		validOperators := map[string]bool{
			"=": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
			"like": true, "in": true, "not in": true, "between": true,
		}
		for _, cond := range processConditions {
			if cond.Field != field {
				return errors.New("没有该条件字段，请检查")
			}
			if !validOperators[strings.ToLower(cond.Operator)] {
				return errors.New("不支持的操作符")
			}
			cond.Value = strings.ReplaceAll(cond.Value, "'", "\\'")
			cond.Extra = strings.ReplaceAll(cond.Extra, "'", "\\'")
		}

		conditionSql := ""
		for _, cond := range processConditions {
			conditionSql += fmt.Sprintf(" `field_value` %s '%s' %s", cond.Operator, cond.Value, cond.Extra)
		}

		if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(field) {
			return errors.New("无效的字段名")
		}
		escapedField := strings.ReplaceAll(field, "'", "\\'")
		conditionSql = fmt.Sprintf("SELECT count(*) as number FROM entrydatas WHERE entry_id=? AND flow_id=? AND (%s) AND (`field_name`='%s')", conditionSql, escapedField)

		var resultCount struct{ Number int }
		if err := query.Raw(conditionSql, proc.EntryID, proc.FlowID).Scan(&resultCount).Error; err != nil {
			return errors.New("条件语法错误，请检查")
		}
		if resultCount.Number > 0 {
			matchedFlowlink = m
			break
		}
	}

	if matchedFlowlink.ID == 0 {
		return errors.New("未找到符合条件的流转条件，无法流转")
	}

	var withFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).With("NextProcess").Where("id=?", matchedFlowlink.ID).First(&withFlowlink)

	auditor_ids := w.GetProcessAuditorIds(proc.Entry, withFlowlink.NextProcessID)
	if len(auditor_ids) == 0 {
		return errors.New("未找到下一步骤审批人")
	}

	if err := w.createProcsForProcess(query, &proc.Entry, withFlowlink.NextProcessID, withFlowlink.NextProcess.ProcessName, auditor_ids); err != nil {
		return err
	}

	query.Model(&models.Entry{}).Where("id=?", proc.EntryID).Update("process_id", cast.ToUint(withFlowlink.NextProcessID))

	// Update parent entry child field if needed
	if proc.Entry.Pid > 0 {
		parentEntry := models.Entry{}
		query.Model(&models.Entry{}).Where("pid=?", proc.Entry.Pid).First(&parentEntry)
		if parentEntry.ID > 0 {
			query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Update("child", withFlowlink.NextProcessID)
		}
	}

	return w.finalizeProc(query, proc, content, emp, withFlowlink.NextProcessID)
}

// handleChildWorkflow creates or resumes a child workflow
func (w *Workflow) handleChildWorkflow(query orm.Query, proc *models.Proc, fklink *models.Flowlink, content string, emp models.Emp) error {
	child_entry := models.Entry{}
	query.Model(&models.Entry{}).
		Where("pid=?", proc.Entry.ID).
		Where("circle=?", proc.Entry.Circle).
		First(&child_entry)

	if child_entry.ID == 0 {
		newChild := models.Entry{
			Title:          proc.Entry.Title,
			FlowID:         cast.ToUint(fklink.Process.ChildFlowID),
			EmpID:          cast.ToUint(proc.Entry.EmpID),
			Status:         models.ProcStatusPending,
			Pid:            cast.ToInt(proc.Entry.ID),
			Circle:         proc.Entry.Circle,
			EnterProcessID: cast.ToInt(fklink.ProcessID),
			EnterProcID:    cast.ToInt(proc.ID),
		}
		query.Model(&models.Entry{}).Create(&newChild)

		query.Model(&models.Entry{}).Where("id=?", newChild.ID).
			With("Flow").With("Process").With("EnterProcess").With("Emp.Dept").First(&child_entry)
	} else {
		query.Model(&models.Entry{}).Where("id=?", child_entry.ID).
			With("Flow").With("Process").With("EnterProcess").With("Emp.Dept").First(&child_entry)
	}

	// Find child's first flowlink
	child_flowlink := models.Flowlink{}
	execSQL := "SELECT * FROM flowlinks AS f " +
		"WHERE f.flow_id = (SELECT child_flow_id FROM processes WHERE id = ? AND f.type = 'Condition' " +
		"AND EXISTS (SELECT * FROM processes AS p WHERE p.id = f.process_id AND p.position = 0) " +
		"ORDER BY f.sort ASC LIMIT 1);"
	query.Raw(execSQL, fklink.ProcessID).Scan(&child_flowlink)

	var resolvedFlowlink models.Flowlink
	query.Model(&models.Flowlink{}).Where("id=?", child_flowlink.ID).With("Process").With("NextProcess").First(&resolvedFlowlink)

	if err := w.SetFirstProcessAuditor(child_entry, resolvedFlowlink); err != nil {
		return err
	}

	query.Model(&models.Entry{}).Where("id=?", child_entry.Pid).Update("child", child_entry.ProcessID)

	return w.finalizeProc(query, proc, content, emp, cast.ToInt(fklink.ProcessID))
}

// handleLastStep completes the entry and handles parent workflow linkage
func (w *Workflow) handleLastStep(query orm.Query, proc *models.Proc, fklink *models.Flowlink, content string, emp models.Emp) error {
	procEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", proc.EntryID).First(&procEntry)
	procEntry.Status = models.ProcStatusApproved
	procEntry.ProcessID = fklink.ProcessID
	query.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)

	if proc.Entry.Pid > 0 {
		return w.handleParentAfterChildComplete(query, proc, content, emp)
	}

	return w.finalizeProc(query, proc, content, emp, cast.ToInt(fklink.ProcessID))
}

// handleParentAfterChildComplete handles parent workflow when child completes
func (w *Workflow) handleParentAfterChildComplete(query orm.Query, proc *models.Proc, content string, emp models.Emp) error {
	parentEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)

	if parentEntry.EnterProcess.ChildAfter == 1 {
		parentEntry.Status = models.ProcStatusApproved
		parentEntry.Child = 0
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
		w.NotifySendOne(proc.Entry.ID)
	} else if parentEntry.EnterProcess.ChildAfter == 2 {
		if parentEntry.EnterProcess.ChildBackProcess > 0 {
			w.goToProcess(query, &parentEntry, parentEntry.EnterProcess.ChildBackProcess)
		} else {
			parentFlowlink := models.Flowlink{}
			query.Model(&models.Flowlink{}).Where("process_id=?", parentEntry.EnterProcessID).Where("type != ?", "Condition").First(&parentFlowlink)
			if parentFlowlink.NextProcessID == -1 {
				parentEntry.Status = models.ProcStatusApproved
				parentEntry.Child = 0
				query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
				w.NotifySendOne(proc.Entry.EmpID)
			} else {
				w.goToProcess(query, &parentEntry, parentFlowlink.NextProcessID)
				parentEntry.ProcessID = cast.ToUint(parentFlowlink.NextProcessID)
				parentEntry.Status = models.ProcStatusPending
				query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
				w.NotifySendOne(cast.ToUint(proc.AuditorID))
			}
		}
		parentEntry.Child = 0
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
	}

	return nil
}

// handleNextStep creates procs for the next process and continues the workflow
func (w *Workflow) handleNextStep(query orm.Query, proc *models.Proc, fklink *models.Flowlink, content string, emp models.Emp) error {
	auditor_ids := w.GetProcessAuditorIds(proc.Entry, fklink.NextProcessID)
	if len(auditor_ids) == 0 {
		return errors.New("未找到下一步步骤审批人")
	}

	if err := w.createProcsForProcess(query, &proc.Entry, fklink.NextProcessID, fklink.NextProcess.ProcessName, auditor_ids); err != nil {
		return err
	}

	procEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", proc.Entry.ID).First(&procEntry)
	procEntry.ProcessID = cast.ToUint(fklink.NextProcessID)
	query.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)

	// Update parent entry child field if needed
	var parentEntry models.Entry
	query.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)
	if parentEntry.ID > 0 {
		parentEntry.Child = fklink.NextProcessID
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
	}

	return w.finalizeProc(query, proc, content, emp, fklink.NextProcessID)
}

// finalizeProc marks the current proc as approved and triggers plugins/CC
func (w *Workflow) finalizeProc(query orm.Query, proc *models.Proc, content string, emp models.Emp, nextProcessID int) error {
	pluginConfigs := []official_plugins.PluginConfig{}
	query.Model(official_plugins.PluginConfig{}).Where("process_id=?", proc.ProcessID).Find(&pluginConfigs)
	pluginConfigJSON, _ := json.Marshal(pluginConfigs)

	// Mark current proc as approved
	var currentProc models.Proc
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("process_id=?", proc.ProcessID).
		Where("circle=?", proc.Entry.Circle).
		Where("status=?", models.ProcStatusPending).
		First(&currentProc)
	if currentProc.ID > 0 {
		currentProc.Status = models.ProcStatusApproved
		currentProc.AuditorID = cast.ToInt(emp.ID)
		currentProc.AuditorName = emp.Name
		currentProc.DeptName = emp.Dept.DeptName
		currentProc.Content = content
		currentProc.Beizhu = string(pluginConfigJSON)
		currentProc.Concurrence = carbon.NewDateTime(carbon.Now())
		query.Model(&models.Proc{}).Where("id=?", currentProc.ID).Save(&currentProc)
	}

	w.ExecPluginMethod("DistributePlugin", cast.ToUint(proc.FlowID), cast.ToUint(proc.ProcessID))
	w.triggerCC(proc.EntryID, cast.ToUint(proc.FlowID), cast.ToUint(proc.ProcessID), proc.ID)

	return nil
}

// handleRejectEntry handles entry-level rejection (mark entry as rejected, notify initiator)
func (w *Workflow) handleRejectEntry(query orm.Query, proc *models.Proc) error {
	procEntry := models.Entry{}
	query.Model(&models.Entry{}).Where("id=?", proc.EntryID).First(&procEntry)
	procEntry.Status = models.ProcStatusRejected
	query.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)

	if proc.Entry.Pid > 0 {
		parentEntry := models.Entry{}
		query.Model(&models.Entry{}).Where("id=?", proc.Entry.Pid).First(&parentEntry)
		parentEntry.Child = proc.ProcessID
		parentEntry.Status = models.ProcStatusRejected
		query.Model(&models.Entry{}).Where("id=?", parentEntry.ID).Save(&parentEntry)
	}

	w.NotifySendOne(proc.Entry.EmpID)
	return nil
}

// goToProcess creates approval tasks at a specified process step
func (w *Workflow) goToProcess(query orm.Query, entry *models.Entry, processID int) error {
	auditor_ids := w.GetProcessAuditorIds(*entry, processID)
	if len(auditor_ids) == 0 {
		return errors.New("未找到审批人")
	}

	var processName string
	query.Model(&models.Process{}).Where("id=?", processID).Select("process_name").First(&processName)

	return w.createProcsForProcess(query, entry, processID, processName, auditor_ids)
}

// Pass is an alias for Transfer
func (w *Workflow) Pass(process_id int, user models.Emp, content string) error {
	return w.Transfer(process_id, user, content)
}

// UnPass rejects the current approval task and sends it back to the previous step
func (w *Workflow) UnPass(proc_id int, user models.Emp, content string) error {
	return w.UnPassTo(proc_id, user, content, 0)
}

// UnPassTo rejects to a specific target process step (0 = previous step)
func (w *Workflow) UnPassTo(proc_id int, user models.Emp, content string, targetProcessID int) error {
	query := facades.Orm().Query()
	var emp models.Emp
	query.Model(&models.Emp{}).Where("user_id=?", user.ID).First(&emp)

	var proc models.Proc
	query.Model(&models.Proc{}).Where("id=?", proc_id).With("Entry").First(&proc)
	if proc.ID == 0 {
		return errors.New("审批任务不存在")
	}

	if targetProcessID > 0 {
		return w.rejectToNode(query, &proc, emp, content, targetProcessID)
	}

	return w.rejectPreviousStep(query, &proc, emp, content)
}

// rejectPreviousStep implements the original reject-to-previous-step logic
func (w *Workflow) rejectPreviousStep(query orm.Query, proc *models.Proc, emp models.Emp, content string) error {
	var todoProc models.Proc
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("process_id=?", proc.ProcessID).
		Where("circle=?", proc.Entry.Circle).
		Where("status=?", models.ProcStatusPending).
		First(&todoProc)

	if todoProc.ID == 0 {
		return w.handleRejectEntry(query, proc)
	}

	todoProc.Status = models.ProcStatusRejected
	todoProc.AuditorID = cast.ToInt(emp.ID)
	todoProc.AuditorName = emp.Name
	todoProc.AuditorDept = emp.Dept.DeptName
	todoProc.Content = content
	todoProc.IsRead = 1
	todoProc.Concurrence = carbon.NewDateTime(carbon.Now())
	query.Model(&models.Proc{}).Where("id=?", todoProc.ID).Save(&todoProc)

	return w.handleRejectEntry(query, proc)
}

// rejectToNode implements reject-to-arbitrary-node logic
func (w *Workflow) rejectToNode(query orm.Query, proc *models.Proc, emp models.Emp, content string, targetProcessID int) error {
	// Find all pending procs for this entry
	var allPendingProcs []models.Proc
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("circle=?", proc.Entry.Circle).
		Where("status=?", models.ProcStatusPending).
		Find(&allPendingProcs)

	// Find the target proc
	var targetProc models.Proc
	found := false
	for i := range allPendingProcs {
		if allPendingProcs[i].ProcessID == targetProcessID {
			targetProc = allPendingProcs[i]
			found = true
			break
		}
	}

	if !found {
		// Target process doesn't have a pending proc — search all procs
		query.Model(&models.Proc{}).
			Where("entry_id=?", proc.EntryID).
			Where("circle=?", proc.Entry.Circle).
			Where("process_id=?", targetProcessID).
			Order("id DESC").
			First(&targetProc)

		if targetProc.ID == 0 {
			return errors.New("目标审批节点不存在")
		}

		// Mark all procs between target and current as skipped
		var currentProcs []models.Proc
		query.Model(&models.Proc{}).
			Where("entry_id=?", proc.EntryID).
			Where("circle=?", proc.Entry.Circle).
			Where("status IN (?, ?)", models.ProcStatusPending, models.ProcStatusApproved).
			Order("id ASC").
			Find(&currentProcs)

		skipped := false
		for i := range currentProcs {
			if skipped {
				currentProcs[i].Status = models.ProcStatusSkipped
				query.Model(&models.Proc{}).Where("id=?", currentProcs[i].ID).Save(&currentProcs[i])
			}
			if currentProcs[i].ProcessID == targetProcessID {
				skipped = true
			}
		}

		// Reset target proc to pending
		targetProc.Status = models.ProcStatusPending
		targetProc.IsRead = 0
		targetProc.Concurrence = carbon.NewDateTime(carbon.Now())
		query.Model(&models.Proc{}).Where("id=?", targetProc.ID).Save(&targetProc)

		// Update entry to point to target process
		procEntry := models.Entry{}
		query.Model(&models.Entry{}).Where("id=?", proc.EntryID).First(&procEntry)
		procEntry.Status = models.ProcStatusPending
		procEntry.ProcessID = cast.ToUint(targetProcessID)
		query.Model(&models.Entry{}).Where("id=?", procEntry.ID).Save(&procEntry)

		w.NotifyNextAuditor(uint(targetProc.EmpID))
		w.NotifySendOne(proc.Entry.EmpID)
		return nil
	}

	// Target proc exists and is pending — mark current procs as skipped
	var currentProcs []models.Proc
	query.Model(&models.Proc{}).
		Where("entry_id=?", proc.EntryID).
		Where("circle=?", proc.Entry.Circle).
		Where("status IN (?, ?)", models.ProcStatusPending, models.ProcStatusApproved).
		Order("id ASC").
		Find(&currentProcs)

	skipped := false
	for i := range currentProcs {
		if skipped {
			currentProcs[i].Status = models.ProcStatusSkipped
			query.Model(&models.Proc{}).Where("id=?", currentProcs[i].ID).Save(&currentProcs[i])
		}
		if currentProcs[i].ProcessID == targetProcessID {
			skipped = true
		}
	}

	// Mark the rejecting proc as skipped
	proc.Status = models.ProcStatusSkipped
	proc.Content = content
	proc.AuditorID = cast.ToInt(emp.ID)
	proc.AuditorName = emp.Name
	query.Model(&models.Proc{}).Where("id=?", proc.ID).Save(&proc)

	w.NotifySendOne(proc.Entry.EmpID)
	return nil
}

// ExecPluginMethod executes a plugin
func (w *Workflow) ExecPluginMethod(plugin_name string, flowID uint, processID uint) error {
	ctor := GetCollectorIns()
	return ctor.DoPluginsExec(plugin_name, flowID, processID)
}

// Revoke allows the initiator to withdraw a pending entry
func (w *Workflow) Revoke(entryID uint, user models.Emp) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var entry models.Entry
		tx.Model(&models.Entry{}).Where("id=?", entryID).With("Emp").First(&entry)
		if entry.ID == 0 {
			return errors.New("流程不存在")
		}
		if cast.ToInt(entry.EmpID) != cast.ToInt(user.ID) {
			return errors.New("只有发起人才能撤回流程")
		}
		if entry.Status != models.ProcStatusPending {
			return errors.New("当前流程状态不允许撤回")
		}

		var pendingProcs []models.Proc
		tx.Model(&models.Proc{}).Where("entry_id=?", entryID).Where("status=?", models.ProcStatusPending).Find(&pendingProcs)
		for _, p := range pendingProcs {
			if p.AuditorID != 0 {
				return errors.New("流程已被处理，无法撤回")
			}
		}

		entry.Status = models.ProcStatusRevoked
		tx.Model(&models.Entry{}).Where("id=?", entryID).Save(&entry)

		for _, p := range pendingProcs {
			p.Status = models.ProcStatusRevoked
			p.AuditorID = cast.ToInt(user.ID)
			p.AuditorName = user.Name
			tx.Model(&models.Proc{}).Where("id=?", p.ID).Save(&p)
		}
		return nil
	})
}

// AddSign adds an additional approver to the current approval task
func (w *Workflow) AddSign(entryID uint, processID int, signEmpID int, signType string, currentUser models.Emp) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var entry models.Entry
		tx.Model(&models.Entry{}).Where("id=?", entryID).First(&entry)
		if entry.ID == 0 {
			return errors.New("流程不存在")
		}
		if entry.Status != models.ProcStatusPending {
			return errors.New("流程状态不允许加签")
		}

		var targetProc models.Proc
		tx.Model(&models.Proc{}).
			Where("entry_id=?", entryID).
			Where("process_id=?", processID).
			Where("emp_id=?", currentUser.ID).
			Where("status=?", models.ProcStatusPending).
			First(&targetProc)
		if targetProc.ID == 0 {
			return errors.New("未找到当前审批任务")
		}

		var signEmp models.Emp
		tx.Model(&models.Emp{}).Where("id=?", signEmpID).With("Dept").First(&signEmp)
		if signEmp.ID == 0 {
			return errors.New("被加签员工不存在")
		}

		sign := models.ProcAddSign{
			EntryID:     entryID,
			ProcID:      targetProc.ID,
			SignType:    signType,
			SignEmpID:   signEmpID,
			SignEmpName: signEmp.Name,
			Status:      models.ProcStatusPending,
		}
		tx.Model(&models.ProcAddSign{}).Create(&sign)

		newProc := models.Proc{
			EntryID:     entry.ID,
			FlowID:      cast.ToInt(entry.FlowID),
			ProcessID:   targetProc.ProcessID,
			ProcessName: targetProc.ProcessName,
			EmpID:       cast.ToInt(signEmp.ID),
			EmpName:     signEmp.Name,
			DeptName:    signEmp.Dept.DeptName,
			Circle:      targetProc.Circle,
			Status:      models.ProcStatusPending,
			IsRead:      0,
			Concurrence: carbon.NewDateTime(carbon.Now()),
		}
		tx.Model(&models.Proc{}).Create(&newProc)
		return nil
	})
}

// TransferProc transfers the current approval task to another employee
func (w *Workflow) TransferProc(entryID uint, procID uint, targetEmpID int, currentUser models.Emp) error {
	return facades.Orm().Transaction(func(tx orm.Query) error {
		var targetProc models.Proc
		tx.Model(&models.Proc{}).Where("id=?", procID).First(&targetProc)
		if targetProc.ID == 0 {
			return errors.New("审批任务不存在")
		}
		if cast.ToInt(targetProc.EntryID) != cast.ToInt(entryID) {
			return errors.New("审批任务与流程不匹配")
		}
		if targetProc.Status != models.ProcStatusPending {
			return errors.New("审批任务已处理，无法转交")
		}

		var entry models.Entry
		tx.Model(&models.Entry{}).Where("id=?", entryID).First(&entry)
		if entry.Status != models.ProcStatusPending {
			return errors.New("流程状态不允许转交")
		}

		var targetEmp models.Emp
		tx.Model(&models.Emp{}).Where("id=?", targetEmpID).With("Dept").First(&targetEmp)
		if targetEmp.ID == 0 {
			return errors.New("被转交员工不存在")
		}

		newProc := models.Proc{
			EntryID:     entry.ID,
			FlowID:      cast.ToInt(entry.FlowID),
			ProcessID:   targetProc.ProcessID,
			ProcessName: targetProc.ProcessName,
			EmpID:       cast.ToInt(targetEmp.ID),
			EmpName:     targetEmp.Name,
			DeptName:    targetEmp.Dept.DeptName,
			Circle:      targetProc.Circle,
			Status:      models.ProcStatusPending,
			IsRead:      0,
			Concurrence: carbon.NewDateTime(carbon.Now()),
		}
		tx.Model(&models.Proc{}).Create(&newProc)

		targetProc.Status = models.ProcStatusTransferred
		targetProc.AuditorID = cast.ToInt(currentUser.ID)
		targetProc.AuditorName = currentUser.Name
		targetProc.Content = "已转交给" + targetEmp.Name
		tx.Model(&models.Proc{}).Where("id=?", procID).Save(&targetProc)

		return nil
	})
}

// AddComment adds a comment to an entry
func (w *Workflow) AddComment(entryID uint, procID uint, empID int, empName string, content string) error {
	comment := models.ProcComment{
		EntryID: entryID,
		ProcID:  procID,
		EmpID:   empID,
		EmpName: empName,
		Content: content,
		Status:  1,
	}
	return facades.Orm().Query().Model(&models.ProcComment{}).Create(&comment)
}

// GetComments retrieves all comments for an entry
func (w *Workflow) GetComments(entryID uint) ([]models.ProcComment, error) {
	var comments []models.ProcComment
	err := facades.Orm().Query().
		Model(&models.ProcComment{}).
		Where("entry_id=? AND status=?", entryID, 1).
		Order("id asc").
		Find(&comments)
	return comments, err
}

// triggerCC creates CC records after a proc is approved
func (w *Workflow) triggerCC(entryID, flowID, processID, procID uint) {
	var ccEmpIDs []string
	facades.Orm().Query().
		Model(&models.Process{}).
		Where("id=?", processID).
		Pluck("cc_emp_ids", &ccEmpIDs)

	if len(ccEmpIDs) == 0 || ccEmpIDs[0] == "" {
		return
	}

	var empIDs []int
	for _, idStr := range strings.Split(ccEmpIDs[0], ",") {
		if id := cast.ToInt(strings.TrimSpace(idStr)); id > 0 {
			empIDs = append(empIDs, id)
		}
	}
	if len(empIDs) == 0 {
		return
	}

	var emps []models.Emp
	facades.Orm().Query().Model(&models.Emp{}).Where("id IN (?)", empIDs).Find(&emps)

	for _, emp := range emps {
		record := models.CcRecord{
			EntryID:   entryID,
			FlowID:    flowID,
			ProcessID: processID,
			ProcID:    procID,
			EmpID:     cast.ToInt(emp.ID),
			EmpName:   emp.Name,
			Status:    0,
		}
		facades.Orm().Query().Model(&models.CcRecord{}).Create(&record)
	}
}
