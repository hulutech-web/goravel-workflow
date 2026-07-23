# 审批流工作引擎 — 时序图

## 1. 发起流程

```
User → EntryController.Store()
  ├─ 查询 Flow + Template + TemplateForms
  ├─ DynamicValidator 校验表单数据
  ├─ 创建 Entry 记录 (status=0)
  ├─ Workflow.SetFirstProcessAuditor(entry, flowlink)
  │   ├─ 查询第一步的 Flowlink
  │   ├─ GetProcessAuditorIds() 计算审批人
  │   │   ├─ 检查 Auditor = -1000 (发起人)
  │   │   ├─ 检查 Auditor = -1001 (部门主管)
  │   │   ├─ 检查 Auditor = -1002 (部门经理)
  │   │   ├─ 检查 type=Emp (指定员工)
  │   │   └─ 检查 type=Dept (指定部门→查DirectorID)
  │   │   └─ uniqueSlice() 去重
  │   └─ 为每个审批人创建 Proc (status=0)
  ├─ 保存 EntryData (表单字段值)
  └─ 返回 Entry
```

## 2. 审批流转（核心路径）

```
用户 → EntryController.Pass() / ProcController.Pass()
  └─ Workflow.Pass(process_id, user, content)
      └─ Workflow.Transfer(process_id, user, content)
          ├─ 查询 Emp (通过 user_id)
          ├─ 查询 Proc (process_id=? AND emp_id=? AND status=0)
          │   └─ 找不到 → 返回错误"未绑定员工"
          ├─ 查询 Flowlink.ConcurrencyType
          │
          ├── ConcurrencyType = 1 (会签)
          │   ├─ checkConsensusComplete():
          │   │   ├─ 统计总Proc数
          │   │   ├─ 统计已通过Proc数
          │   │   └─ 统计已驳回Proc数
          │   ├─ 所有人未完成 → 标记当前Proc为已通过, 返回等待其他人
          │   └─ 所有人已完成:
          │       ├─ 有驳回 → handleRejectEntry() 驳回整个流程
          │       └─ 全通过 → 继续后续流转
          │
          ├── ConcurrencyType = 2 (或签)
          │   └─ skipRemainingConcurrentProcs():
          │       └─ 标记同一步骤下其他待处理Proc为已跳过(status=4)
          │
          └── ConcurrencyType = 0 (依次/默认)
              ├─ 查询 Flowlink (type=Condition) 判断是否有分支
              │
              ├── 有分支 (fkcount > 1)
              │   ├─ 查询 ProcessVar 获取条件字段
              │   ├─ 查询 EntryData 获取字段值
              │   ├─ 遍历 Condition Flowlink:
              │   │   ├─ Expression="1" → 无条件匹配
              │   │   └─ 解析 JSON → ProcessCondition[]
              │   │       ├─ 白名单校验操作符 (=, !=, >, <, >=, <=, like...)
              │   │       ├─ 转义值中的单引号
              │   │       └─ 正则校验字段名 (字母数字下划线)
              │   ├─ 构造参数化 SQL 查询 entrydatas
              │   ├─ 选中匹配的 Flowlink
              │   └─ GetProcessAuditorIds(NextProcessID)
              │
              ├── 无分支 (fkcount <= 1)
              │   ├─ 查询 Flowlink (NextProcessID)
              │   │
              │   ├── NextProcessID = -1 (最后一步)
              │   │   ├─ 更新 Entry status=9 (已完成)
              │   │   ├─ 如果有父流程且 ChildAfter=1 → 同时结束父流程
              │   │   ├─ 如果 ChildAfter=2 → 进入父流程下一步
              │   │   └─ NotifySendOne(发起人)
              │   │
              │   └── 还有下一步
              │       ├─ GetProcessAuditorIds(NextProcessID)
              │       ├─ 为每个审批人创建 Proc (status=0)
              │       ├─ NotifyNextAuditor(审批人)
              │       └─ 更新 Entry.process_id
              │
              └─ [通用步骤]
                  ├─ 标记当前 Proc status=1 (已通过)
                  ├─ 填充 AuditorID, AuditorName, Content
                  ├─ ExecPluginMethod("DistributePlugin", flowID, processID)
                  └─ triggerCC(entryID, flowID, processID, procID)
                      ├─ 查询 Process.cc_emp_ids
                      ├─ 查询对应的 Emp
                      └─ 为每个抄送人创建 CcRecord (status=0)
```

## 3. 驳回

### 3a. 普通驳回（UnPass — 驳回到上一步）

```
用户 → ProcController.UnPass(proc_id, user, content)
  └─ Workflow.UnPass(proc_id, user, content)
      └─ Workflow.UnPassTo(proc_id, user, content, targetProcessID=0)
          ├─ 查询 Emp (通过 user_id)
          ├─ 查询 Proc (加载 Entry)
          ├─ 查找同一 Entry/Process/Circle 下 status=0 的任务
          ├─ 标记 todoProc status=-1 (已驳回)
          ├─ 更新 Entry status=-1
          ├─ 如果有父流程 (Pid > 0)
          │   └─ 同步更新父流程 status=-1
          └─ NotifySendOne(发起人)
```

### 3b. 驳回到指定节点（UnPassTo — 任意节点驳回）

```
用户 → ProcController.UnPass(proc_id, user, content, targetProcessID)
  └─ Workflow.UnPassTo(proc_id, user, content, targetProcessID)
      ├─ 查询 Emp (通过 user_id)
      ├─ 查询 Proc (加载 Entry)
      ├─ 查找目标步骤的 Proc:
      │   ├─ 先查同一 Entry/Circle 下待处理 Proc
      │   └─ 如果不存在 → 查所有 Proc (按 ID 倒序取最新)
      ├─ 如果目标 Proc 不存在 → 返回错误"目标审批节点不存在"
      ├─ 标记目标步骤之后(按ID排序)的所有Proc为已跳过(status=4)
      ├─ 重置目标Proc为待处理状态(status=0)
      ├─ 更新Entry status=0, ProcessID=目标步骤
      ├─ NotifyNextAuditor(目标审批人)
      └─ NotifySendOne(发起人)
```

## 4. 撤回

```
发起人 → ProcController.Revoke(entry_id, user)
  或 EntryController.Revoke(entry_id, user)
  └─ Workflow.Revoke(entryID, user) [事务]
      ├─ 查询 Entry (加载 Emp)
      │   └─ Entry不存在 → 返回"流程不存在"
      ├─ 校验: entry.EmpID == user.ID
      │   └─ 不是发起人 → 返回"只有发起人才能撤回"
      ├─ 校验: entry.Status == 0
      │   └─ 状态不允许 → 返回"当前流程状态不允许撤回"
      ├─ 查询所有 pending Procs (entry_id=? AND status=0)
      │   └─ 找到 auditor_id!=0 的任务 → 返回"流程已被处理"
      ├─ 更新 Entry status=-2 (已撤回)
      └─ 批量更新所有 pending Proc:
          ├─ status = -2
          ├─ auditor_id = 发起人ID
          └─ auditor_name = 发起人姓名
```

## 5. 加签

```
当前审批人 → ProcController.AddSign(entry_id, process_id, sign_emp_id, sign_type)
  └─ Workflow.AddSign(entryID, processID, signEmpID, signType, currentUser) [事务]
      ├─ 查询 Entry
      │   └─ 不存在或 status!=0 → 返回错误
      ├─ 查询当前用户的待办 Proc
      │   └─ 找不到 → 返回"未找到当前审批任务"
      ├─ 查询被加签的 Emp
      │   └─ 不存在 → 返回"被加签员工不存在"
      ├─ 创建 ProcAddSign 记录
      │   ├─ sign_type: "before"(前加签) / "after"(后加签)
      │   └─ status: 0 (待处理)
      └─ 创建新 Proc (给被加签人, status=0)
```

## 6. 转交

```
当前审批人 → ProcController.TransferProc(entry_id, proc_id, target_emp_id)
  └─ Workflow.TransferProc(entryID, procID, targetEmpID, currentUser) [事务]
      ├─ 查询目标 Proc
      │   └─ 不存在 → 返回"审批任务不存在"
      ├─ 校验: proc.EntryID == entryID
      │   └─ 不匹配 → 返回"审批任务与流程不匹配"
      ├─ 校验: proc.Status == 0
      │   └─ 已处理 → 返回"审批任务已处理，无法转交"
      ├─ 查询 Entry
      │   └─ status!=0 → 返回"流程状态不允许转交"
      ├─ 查询被转交人 Emp
      │   └─ 不存在 → 返回"被转交员工不存在"
      ├─ 创建新 Proc (给被转交人, status=0)
      └─ 标记原 Proc:
          ├─ status = 3 (已转交)
          ├─ auditor_id = 当前用户
          ├─ auditor_name = 当前用户
          └─ content = "已转交给{被转交人姓名}"
```

## 7. 评论

```
用户 → ProcController.AddComment(entry_id, proc_id, content)
  └─ Workflow.AddComment(entryID, procID, empID, empName, content)
      └─ 创建 ProcComment 记录 (status=1 正常)

用户 → ProcController.GetComments(entry_id)
  └─ Workflow.GetComments(entryID)
      └─ 查询 entry_id=? AND status=1 的所有评论 (按 ID 升序)
```

## 8. 抄送列表

```
用户 → CcController.List(user)
  ├─ 查询 Emp (通过 user_id)
  └─ 查询该 EmpID 的所有 CcRecord (按 ID 降序)

用户 → CcController.GetEntryCC(entry_id)
  └─ 查询该 entry_id 的所有 CcRecord (按 ID 升序)
```

## 9. 子流程

```
父流程 Transfer() 检测到 ChildFlowID > 0
  ├─ 查询是否存在子流程 Entry (pid=父EntryID AND circle=父Circle)
  │   ├─ 存在 → 直接使用
  │   └─ 不存在 → 创建新子流程 Entry
  │       ├─ Title = 父流程标题
  │       ├─ FlowID = ChildFlowID
  │       ├─ EmpID = 父流程发起人
  │       ├─ Pid = 父EntryID
  │       └─ EnterProcessID = ChildFlowID
  ├─ 查询子流程起始 Flowlink
  ├─ Workflow.SetFirstProcessAuditor(child_entry, child_flowlink)
  │   └─ 同第1步"发起流程"
  └─ 更新父Entry.child = 子流程当前步骤
```

## 10. 父子流程联动

```
子流程完成 (Transfer 中 NextProcessID=-1)
  └─ 检查父流程 Entry.Pid > 0
      └─ 检查父流程 EnterProcess.ChildAfter
          ├─ ChildAfter = 1 → 同时结束父流程
          │   ├─ 更新父Entry status=9, child=0
          │   └─ NotifySendOne(发起人)
          │
          └─ ChildAfter = 2 → 返回父流程
              ├─ 检查 ChildBackProcess > 0
              │   └─ 是 → goToProcess(父Entry, ChildBackProcess)
              │           ├─ GetProcessAuditorIds()
              │           ├─ 创建 Proc 给审批人
              │           └─ 更新父Entry.process_id
              │
              └─ 否 → 进入父流程下一步
                  ├─ 查询父流程 Flowlink (type=Condition)
                  ├─ 如果 NextProcessID=-1 → 直接结束父流程
                  └─ 否则 → goToProcess(父Entry, NextProcessID)
```

## 11. 定时超时检查

```
定时任务触发: workflow:timeout-check (通过系统cron定期执行)
  ├─ 查询所有 status=0 的 Proc
  ├─ 遍历每个 Proc:
  │   ├─ 查询对应 Process.LimitTime (秒)
  │   ├─ 如果 LimitTime <= 0 → 跳过
  │   └─ 计算 Concurrence 与当前时间的差值 (秒)
  │   └─ 如果超过 LimitTime:
  │       ├─ 标记 Proc status=-1, content="超时未处理，系统自动驳回"
  │       ├─ 更新 Entry status=-1
  │       └─ 打印日志记录超时详情
  └─ 返回
```

## 关键常量速查

| 状态值 | 含义 |
|--------|------|
| Entry.Status = 0 | 进行中 |
| Entry.Status = 9 | 已完成 |
| Entry.Status = -1 | 已驳回 |
| Entry.Status = -2 | 已撤回 |
| Proc.Status = 0 | 待处理 |
| Proc.Status = 1 | 已通过 |
| Proc.Status = -1 | 已驳回 |
| Proc.Status = -2 | 已撤回 |
| Proc.Status = 3 | 已转交 |
| Proc.Status = 4 | 已跳过（或签中未被选中的审批人） |
| Proc.Status = 9 | 会签通过（第一步无指定审核人时自动通过） |
| Flowlink.Auditor = -1000 | 发起人 |
| Flowlink.Auditor = -1001 | 部门主管 |
| Flowlink.Auditor = -1002 | 部门经理 |
| Flowlink.Auditor = -1003 | 从表单字段读取审批人 |
| Flowlink.Auditor = -1004 | 动态表达式计算审批人 |
| Flowlink.ConcurrencyType = 0 | 依次审批 |
| Flowlink.ConcurrencyType = 1 | 会签：所有人通过 |
| Flowlink.ConcurrencyType = 2 | 或签：一人通过其余跳过 |
| Process.Position = 0 | 第一步 |
| Process.Position = 1 | 正常步骤 |
| Process.Position = 2 | 转入子流程 |
| Process.ChildAfter = 1 | 子流程结束后同时结束父流程 |
| Process.ChildAfter = 2 | 子流程结束后返回父流程 |
