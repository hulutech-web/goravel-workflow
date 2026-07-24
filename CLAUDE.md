# goravel-workflow — Workflow Engine Extension Pack

This is a workflow engine library built on top of `github.com/goravel/framework`. It provides complete approval workflow capabilities including consensus signing (会签), any-sign (或签), rejection, addition/transfer of signers, comments, CC, and timeout checking.

## Directory Structure

```
goravel-workflow/
├── workflow.go              # Entry: ServiceProvider registers routes, migrations, config, commands
├── service_provider.go      # Goravel service provider
├── config/workflow.go       # Config file template (published during setup)
├── contracts/workflow.go    # Workflow interface definition
├── migrations/              # DB migration files (11+ workflow tables)
├── seeders/                 # Seeder data (dept, emp, flow type initialization)
├── models/                  # Data models (Entry, Proc, Flow, Flowlink, Process, etc.)
├── services/workflow/       # Core workflow engine (~1115 lines: Transfer/Pass/UnPass, etc.)
├── services/workflow/official_plugins/  # Plugin system (DistributePlugin, etc.)
├── controllers/             # API controllers (auth, entry, proc, flow, template, etc.)
├── routes/api.go            # API route registration
├── commands/                # Artisan commands (publish, timeout-check, plugin)
├── middleware/jwt.go        # JWT middleware
├── requests/                # Request validation structs
├── rules/                   # Custom validation rules
└── docs/                    # Workflow logic documentation (mermaid diagrams)
```

## Architecture

Core model chain: **Flow → Process → Entry → Proc**

- **Flow** — A workflow definition (e.g., "leave request", "expense approval")
- **Process** — Steps within a flow
- **Flowlink** — Connections between processes (who approves, concurrency type)
- **Entry** — A specific workflow instance (e.g., "John's leave request #123")
- **Proc** — An individual approval task within an entry
- **EntryData** — Form field values submitted with an entry

### Data Models

#### Flow (`models/flow.go`)

| Field | Description |
|-------|-------------|
| `FlowNo` | Flow number |
| `FlowName` | Flow name |
| `TemplateID` | Associated form template ID |
| `IsPublish` | Whether published |
| `Jsplumb` | Flowchart JSON (jsPlumb visualization) |

#### Process (`models/process.go`)

| Field | Description |
|-------|-------------|
| `ProcessName` | Step name |
| `Position` | Step position: 0=first step, 1=normal step, 2=enter child workflow |
| `ChildFlowID` | Child flow ID (when position=2) |
| `ChildAfter` | After child completes: 1=end parent too, 2=return to parent |
| `ChildBackProcess` | Parent step ID to return to (when ChildAfter=2) |
| `LimitTime` | Time limit in seconds |
| `CcEmpIDs` | CC recipient IDs (comma-separated) |
| `AutoPerson` | Auto-approver settings |

#### Flowlink (`models/flowlink.go`)

| Field | Description |
|-------|-------------|
| `Type` | Type: `Condition`=conditional routing, `Emp`=specific employee, `Dept`=specific dept, `Sys`=system auto |
| `Auditor` | Approver setting (see special values below) |
| `ApproverRule` | Approver rule: for -1003 stores form field name, for -1004 stores expression mapping key |
| `ConcurrencyType` | Concurrency mode: 0=sequential, 1=consensus, 2=any-sign |
| `Expression` | Condition expression (JSON), "1" means unconditional |
| `Sort` | Evaluation order |
| `NextProcessID` | Next step ID (-1 = last step) |

#### Entry (`models/entry.go`)

| Field | Description |
|-------|-------------|
| `Title` | Instance title |
| `FlowID` | Belonging flow |
| `EmpID` | Initiator ID |
| `Status` | Status (see status codes below) |
| `Pid` | Parent entry ID (child workflow marker) |
| `Circle` | Round number (increments on resend) |
| `ProcessID` | Current step |
| `EnterProcessID` | Entry step ID |
| `EnterProcID` | Entry task ID |
| `Child` | Child workflow current step |

#### Proc (`models/proc.go`)

| Field | Description |
|-------|-------------|
| `EntryID` | Belonging instance |
| `ProcessID` | Current step |
| `EmpID` | Assigned approver |
| `Status` | Status (see status codes below) |
| `Content` | Approval content |
| `Concurrence` | Creation time (for parallel approval lookup) |
| `IsRead` | Whether viewed |
| `IsReal` | Whether approver == operator |
| `UnpassTargetProcessID` | Target step for rejection |

#### Auxiliary Models
- **ProcComment** — Comments/discussions during approval
- **CcRecord** — CC records created after approval
- **ProcAddSign** — Addition signer records (before/after)
- **EntryData** — Form field data
- **ProcessVar** — Step variables (for conditional evaluation)

## Special Values & Status Codes

### Flowlink.Auditor Special Values
| Value | Meaning |
|-------|---------|
| -1000 | Initiator themselves |
| -1001 | Department director |
| -1002 | Department manager |
| -1003 | Read approver from form field (ApproverRule stores field name) |
| -1004 | Dynamic expression (ApproverRule stores mapping key) |

### Flowlink.ConcurrencyType
| Value | Meaning |
|-------|---------|
| 0 | Sequential (default) |
| 1 | Consensus (会签) — all approvers must approve |
| 2 | Any-sign (或签) — first approver wins, others skipped |

### Flowlink.Type
| Value | Meaning |
|-------|---------|
| Condition | Conditional routing |
| Emp | Specific employee |
| Dept | Specific department |
| Sys | System auto |

### Entry.Status
| Value | Meaning |
|-------|---------|
| 0 | In progress |
| 9 | Completed |
| -1 | Rejected |
| -2 | Revoked |

### Proc.Status
| Value | Meaning |
|-------|---------|
| 0 | Pending |
| 1 | Approved |
| -1 | Rejected |
| -2 | Revoked |
| 3 | Transferred |
| 4 | Skipped (any-sign) |
| 9 | Consensus approved (auto-approved first step) |

### Process.Position
| Value | Meaning |
|-------|---------|
| 0 | First step |
| 1 | Normal step |
| 2 | Enter child workflow |

### Process.ChildAfter
| Value | Meaning |
|-------|---------|
| 1 | End parent workflow when child completes |
| 2 | Return to parent workflow when child completes |

## Integration Steps

### 1. Register ServiceProvider
In target project's `config/app.go`:
```go
import "your/module/path/goravel-workflow"

func init() {
    "providers": []foundation.ServiceProvider{
        // ... existing providers
        &workflow.ServiceProvider{},
    }
}
```

### 2. Publish Resources
```bash
go run . artisan vendor:publish --package=your-module-path/goravel-workflow
```
Optional tags: `--tag=migrations`, `--tag=config`, `--tag=seeders`

### 3. Database Setup
The package provides migrations for workflow-specific tables (flows, flowtypes, processes, entries, procs, etc.). It does **NOT** create `users`, `depts`, or `emps` tables — those must exist in the host application.

Run migrations:
```bash
go run . artisan migrate --seed
```

### 4. Implement User Hook Interface
Add a `Workflow` field to your User model, then implement two methods:

```go
type User struct {
    orm.Model
    Name     string
    WorkNo   string
    Password string
    // ... other fields
    Workflow *workflow.Workflow
    orm.SoftDeletes
}

// Called when entry is rejected or process completes
func (u *User) NotifySendOne(id uint) error {
    // Send notification to initiator
    return nil
}

// Called when next approver is assigned after approval
func (u *User) NotifyNextAuditor(id uint) error {
    // Send notification to next approver
    return nil
}
```

### 5. Configure Model Mapping
After publishing, edit `config/workflow.go`:
```go
"Dept": "Department", // Model name for departments in your app
"Emp":  "User",       // Model name for employees in your app
```

### 6. Register Hooks
In your `AppServiceProvider.Boot()`:
```go
wf := workflow.NewBaseWorkflow()
user := &models.User{Workflow: wf}
wf.RegisterHook("NotifySendOneHook", reflect.ValueOf(user.NotifySendOne))
wf.RegisterHook("NotifyNextAuditorHook", reflect.ValueOf(user.NotifyNextAuditor))
```

### 7. Check Route Conflicts
The package registers many `/api/*` routes. Check for conflicts before starting.

## Core Business Logic

### 1. Create Entry (发起流程)

**Entry:** `EntryController.Store()`

1. Query Flow + Template + TemplateForms by `flow_id`
2. Find first Flowlink (`position=0`, ordered by sort)
3. Validate form data via `DynamicValidator` (rules from Template)
4. Create Entry record (`status=0`)
5. Call `SetFirstProcessAuditor(entry, flowlink)`:
   - Query Flowlinks for this step (type != Condition)
   - If none found (ID=0), auto-create a Proc with status=9 (approved) and advance to next step
   - Calculate approver IDs via `GetProcessAuditorIds()`
   - Query Emp records for each approver
   - Create Proc tasks (status=0) for each approver
   - Update Entry.ProcessID
6. Save EntryData records for each form field
7. Return Entry

### 2. Approval Transfer (审批流转) — Core Engine

**Entry:** `ProcController.Pass()` → `Workflow.Pass()` → `Workflow.Transfer(process_id, user, content)`

1. Resolve user to Emp (via user_id)
2. Find current pending Proc (process_id + emp_id + status=0)
3. Query Flowlink.ConcurrencyType:

   - **Consensus (会签, type=1):**
     - Count total procs, approved procs, rejected procs
     - If not all done → mark current as approved, return (wait for others)
     - If all done and any rejected → reject entire entry (`handleRejectEntry`)
     - If all done and all approved → continue

   - **Any-sign (或签, type=2):**
     - Mark all other pending procs at same step as skipped (status=4)
     - Continue

   - **Sequential (依次, type=0):**
     - Continue normally

4. Check for conditional branches (Count Flowlinks where type=Condition):

   **With conditions (fkcount > 1):**
   - Query ProcessVar for condition field name
   - Query EntryData for field values
   - Iterate Condition Flowlinks:
     - Expression="1" → unconditional match
     - Parse JSON → ProcessCondition[] with whitelist validation (=, !=, >, <, >=, <=, like, in, not in, between)
     - Escape single quotes in values
     - Regex-validate field names (alphanumeric + underscore only)
     - Build parameterized SQL query against entrydatas
     - First matching Flowlink wins

   **Without conditions (fkcount <= 1):**
   - Query Flowlink for NextProcessID

5. Handle NextProcessID:

   **-1 (last step):**
   - Update Entry status=9 (completed)
   - If has parent workflow:
     - ChildAfter=1 → end parent too, notify initiator
     - ChildAfter=2 → return to parent (goToProcess or next step)

   **ChildFlowID > 0 (child workflow):**
   - Find or create child Entry (pid=parent entry id)
   - Initialize child's first Flowlink
   - Call `SetFirstProcessAuditor(child_entry)`
   - Update parent Entry.child field

   **Normal next step:**
   - Calculate approver IDs via `GetProcessAuditorIds()`
   - Create new Proc tasks (status=0)
   - Notify next approvers (NotifyNextAuditor)
   - Update Entry.process_id

6. Mark current Proc as approved (status=1)
7. Execute plugins: `ExecPluginMethod("DistributePlugin", flowID, processID)`
8. Trigger CC: `triggerCC()` — creates CcRecord for each cc_emp_ids

### 3. Rejection (驳回)

**a) Simple Reject (UnPass — reject to previous step):**
1. Find pending Proc at same Entry/Process/Circle
2. If none found → reject entire entry directly
3. Mark todoProc status=-1 (rejected)
4. Update Entry status=-1
5. If has parent → sync parent status=-1
6. Notify initiator (NotifySendOne)

**b) Reject to Specific Node (UnPassTo — arbitrary node rejection):**
1. Find target Proc at targetProcessID:
   - First check pending procs at that step
   - If not found → query all procs (order by id DESC)
2. If target doesn't exist → error
3. Mark all procs between target and current as skipped (status=4)
4. Reset target Proc to pending (status=0)
5. Update Entry: status=0, ProcessID=target
6. Notify target approver + initiator

### 4. Withdrawal (撤回)

**Entry:** Only the initiator can withdraw their own pending entry.

1. Transaction begins
2. Verify: Entry exists, user == initiator, Entry.status == 0
3. Verify: No proc has been handled yet (all pending procs have auditor_id=0)
4. Update Entry status=-2 (revoked)
5. Batch update all pending Procs: status=-2, auditor_id/name = initiator
6. Transaction commits

### 5. Add Signer (加签)

**Entry:** Current approver can add extra approvers.

1. Transaction begins
2. Verify: Entry exists and status=0
3. Find current user's pending Proc at this step
4. Query target Emp
5. Create ProcAddSign record (sign_type: "before"/"after")
6. Create new Proc for the added signer (status=0)
7. Transaction commits

### 6. Transfer (转交)

**Entry:** Current approver transfers their task to someone else.

1. Transaction begins
2. Verify: Proc exists, belongs to entry, status=0 (pending)
3. Verify: Entry status=0
4. Query target Emp
5. Create new Proc for transferee (status=0)
6. Mark original Proc: status=3 (transferred), content="已转交给{name}"
7. Transaction commits

### 7. Comments (评论)

- `AddComment`: Create ProcComment record (status=1)
- `GetComments`: Query all comments for entry_id, ordered by ID ASC

### 8. CC (抄送)

Triggered automatically after `Transfer()` completes:
1. Query Process.CcEmpIDs
2. Query corresponding Emp records
3. Create CcRecord (status=0 unread) for each

### 9. Child Workflows (子流程)

When a Process has Position=2 and ChildFlowID > 0:
1. Find existing child Entry (pid=parent AND circle=parent.circle) or create new one
2. Query child's first Flowlink (position=0, no Condition type)
3. Call `SetFirstProcessAuditor(child_entry)` — same as initial creation
4. Update parent Entry.child = child's current process

**Parent-child linkage on child completion:**
- ChildAfter=1 → Parent also ends (status=9), notify initiator
- ChildAfter=2 → Return to parent:
  - ChildBackProcess > 0 → goToProcess(parent, ChildBackProcess)
  - Otherwise → parent's next step via Flowlink

### 10. Timeout Check (超时检查)

Scheduled command: `workflow:timeout-check` (run via system cron periodically)
1. Query all pending Procs (status=0)
2. For each: check Process.LimitTime (seconds)
3. If LimitTime > 0 and elapsed > LimitTime:
   - Mark Proc status=-1, content="超时未处理，系统自动驳回"
   - Update Entry status=-1
   - Log details

## Approver Calculation Logic (GetProcessAuditorIds)

For a given process step, calculate approver IDs:

1. Query Flowlink with type=Sys (system auto):
   - Auditor="-1000" → add initiator's emp_id
   - Auditor="-1001" → add entry.Emp.Dept.DirectorID
   - Auditor="-1002" → add entry.Emp.Dept.ManagerID
   - Auditor="-1003" → read ApproverRule as field name, query EntryData for value
   - Auditor="-1004" → use ApproverRule as mapping key ("director"/"manager"/numeric ID)
   - Other numeric → add directly

2. If no Sys type, query type=Emp:
   - Split comma-separated IDs, add each

3. If no Emp type, query type=Dept:
   - Split comma-separated dept IDs
   - Query each dept's DirectorID
   - Add all directors

4. Deduplicate via `uniqueSlice()` and return

## Transaction Boundaries

| Operation | Transaction |
|-----------|-------------|
| `Transfer()` | Non-transactional (individual sub-ops may use transactions) |
| `Revoke()` | Full transaction (Entry + all pending Procs) |
| `AddSign()` | Full transaction (ProcAddSign + new Proc) |
| `TransferProc()` | Full transaction (new Proc + mark old) |
| `SetFirstProcessAuditor()` | Uses outer transaction (no nested) |

## Security & Engineering

### SQL Injection Prevention
- Operator whitelist: only `=, !=, >, <, >=, <=, like, in, not in, between`
- Single quotes escaped in values
- Field names validated via regex: `^[a-zA-Z0-9_]+$`
- Parameterized queries with `?` placeholders

### Concurrency Safety
- `sync.RWMutex` protects hooks map (concurrent reads allowed)
- Nil-safe: Workflow methods return errors instead of panicking on nil receiver

### Performance
- Composite indexes on: procs(status, entry_id), procs(emp_id, status), entrydatas(entry_id, field_name), entry(status), entry(flow_id, status)

## Key Services

### Workflow Engine (`services/workflow/workflow.go`)
- `NewBaseWorkflow()` — Singleton instance (sync.Once)
- `RegisterHook(name, method)` — Register callback hooks
- `Transfer(processID, user, content)` — Approve and move to next step
- `Pass(processID, user, content)` — Alias for Transfer
- `UnPass(procID, user, content)` — Reject to previous step
- `UnPassTo(procID, user, content, targetProcessID)` — Reject to specific node
- `Revoke(entryID, user)` — Initiator withdraws their own pending entry
- `AddSign(entryID, processID, signEmpID, signType, currentUser)` — Add extra approver
- `TransferProc(entryID, procID, targetEmpID, currentUser)` — Transfer approval to someone else
- `AddComment(entryID, procID, empID, empName, content)` — Add comment
- `GetComments(entryID)` — Get all comments for an entry
- `ExecPluginMethod(pluginName, flowID, processID)` — Execute plugin

### Controllers (`controllers/`)
All under `/api` prefix with JWT middleware. Key endpoints:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/auth/login` | POST | Admin login |
| `/api/h5/login` | POST | H5 login |
| `/api/captcha/get` | GET | Get captcha |
| `/api/captcha/validate` | POST | Validate captcha |
| `/api/upload` | POST | File upload |
| `/api/dept` | CRUD | Department management |
| `/api/emp` | CRUD | Employee management |
| `/api/flow` | CRUD | Flow definition management |
| `/api/flow/publish` | POST | Publish flow |
| `/api/flow/{id}/entry` | GET | Create entry form |
| `/api/entry` | POST | Submit entry |
| `/api/entry/{id}` | GET | Show entry |
| `/api/entry/resend` | POST | Resend entry |
| `/api/entry/revoke` | POST | Revoke entry |
| `/api/pass` | POST | Approve |
| `/api/unpass` | POST | Reject |
| `/api/revoke` | POST | Revoke (alias) |
| `/api/addsign` | POST | Add signer |
| `/api/transfer` | POST | Transfer |
| `/api/comment` | POST | Add comment |
| `/api/comments/{entry_id}` | GET | Get comments |
| `/api/cc/list` | GET | CC list |
| `/api/cc/entry/{entry_id}` | GET | Entry CC records |
| `/api/process/attribute` | GET | Process attributes |
| `/api/process/con` | POST | Process conditions |
| `/api/process/list` | POST | Process list |

### Commands
- `workflow:publish` — Publish workflow resources
- `workflow:timeout-check` — Check overdue pending tasks
- `plugin` — Plugin management

## Important Notes

1. **Global App singleton**: `service_provider.go` sets `var App foundation.Application` globally. This is fine for Goravel apps but be aware if testing in isolation.

2. **Models expect specific fields**: The `Emp` model expects columns `name`, `workno`, `email`, `password`, `dept_id`, `leave`, `user_id`. The `Dept` model expects `director_id`, `manager_id`. Ensure your host app's tables match these.

3. **Config init() conflict**: Both `config/workflow.go` in the package and the published config call `config.Add("workflow", ...)`. The published version takes precedence once copied.

4. **Plugin system**: `official_plugins/` provides a `DistributePlugin` mechanism. Plugins are registered per-process and executed on approval via `ExecPluginMethod()`.

5. **Child workflows**: The engine supports nested/child workflows via `Flowlink.Process.ChildFlowID`. When a process has a child flow ID, a new Entry is created with `pid` pointing to the parent.

6. **Route naming**: The package uses standard Goravel resource routing (`router.Resource`). Check for conflicts with existing routes in the host application.

7. **JWT middleware**: All workflow routes require JWT authentication via `middleware.Jwt()`. Ensure the host app has JWT configured.

## Debugging & Troubleshooting (from docs/)

### System Architecture

```mermaid
graph TB
    subgraph "前端层"
        A[Web/App] --> B[API Router]
    end
    
    subgraph "控制器层 controllers/"
        B --> C1[EntryController 发起/查询/重发/撤回]
        B --> C2[ProcController 审批/驳回/加签/转交/评论]
        B --> C3[CcController 抄送列表/查询]
    end
    
    subgraph "服务层 services/workflow/"
        C1 --> W[Workflow Singleton 单例模式 + 钩子系统]
        C2 --> W
        C3 -.-> CC[triggerCC 抄送触发器]
        
        W --> H1[SetFirstProcessAuditor 初始化审批链]
        W --> H2[Transfer 核心流转引擎]
        W --> H3[UnPass/UnPassTo 驳回/驳回到指定节点]
        W --> H4[Revoke 撤回]
        W --> H5[AddSign 加签]
        W --> H6[TransferProc 转交]
        W --> H7[AddComment/GetComments 评论]
        W --> H8[getProcessAuditorIds 计算审批人]
        
        H2 --> S1[并签模式检查 会签/或签/依次]
        H2 --> S2[条件分支解析 白名单+参数化SQL]
        H2 --> S3[子流程处理 ChildFlowID]
        H2 --> S4[父子联动 ChildAfter]
        
        W --> P[PluginCollector 插件执行]
    end
    
    subgraph "数据层 models/"
        W --> DB[(SQLite/MySQL)]
        
        DB --> M1[Flow 流程定义]
        DB --> M2[Process 步骤定义]
        DB --> M3[Flowlink 流转关系]
        DB --> M4[Entry 流程实例]
        DB --> M5[Proc 审批任务]
        DB --> M6[EntryData 表单数据]
        DB --> M7[ProcComment 评论]
        DB --> M8[CcRecord 抄送记录]
        DB --> M9[ProcAddSign 加签记录]
        DB --> M10[ProcessVar 条件变量]
    end
    
    subgraph "辅助模块"
        W --> Aux1[DynamicValidator 动态校验]
        W --> Aux2[PluginConfig 插件配置]
        W --> Aux3[timeout_check 定时检查]
    end
    
    style W fill:#f9f,stroke:#333,stroke-width:4px
    style H2 fill:#ff9,stroke:#333,stroke-width:3px
    style DB fill:#9cf,stroke:#333,stroke-width:2px
```

### Transfer Core Decision Tree

```mermaid
graph TD
    Start([用户点击审批]) --> QueryProc[查询当前Proc任务 process_id + emp_id + status=0]
    QueryProc --> CheckProc{Proc存在?}
    CheckProc -->|否| Err1[返回错误:未绑定员工]
    CheckProc -->|是| CheckConcurrency[查询Flowlink.ConcurrencyType]
    
    CheckConcurrency --> ConcType{并发类型?}
    ConcType -.->|0=依次| CountCondition[统计Condition类型Flowlink数量]
    
    ConcType -->|1=会签| CheckConsensus[checkConsensusComplete]
    CheckConsensus --> AllDone{所有人完成?}
    AllDone -->|否| MarkApproved[标记当前Proc已通过,等待其他人]
    AllDone -->|是且有驳回| HandleRejectEntry[驳回整个流程]
    AllDone -->|是且全通过| CountCondition
    
    ConcType -->|2=或签| SkipOthers[skipRemainingConcurrentProcs标记其余为已跳过]
    SkipOthers --> CountCondition
    
    CountCondition --> Branch{fkcount > 1?}
    Branch -->|有条件分支| HasBranch[有条件分支]
    Branch -->|无分支/最后一步| NoBranch[无条件分支]
    
    NoBranch --> CheckLast{NextProcessID等于-1?}
    CheckLast -->|是| FinishEntry[更新Entry status=9 已完成]
    FinishEntry --> CheckParent{有父流程?}
    CheckParent -->|是| CheckChildAfter{ChildAfter值等于?}
    CheckParent -->|否| CommonStep[标记Proc status=1]
    
    CheckChildAfter -->|1| EndBoth[同时结束父流程 NotifySendOne]
    CheckChildAfter -->|2| BackToParent[返回父流程]
    BackToParent --> CheckBackProc{ChildBackProcess?>0?}
    CheckBackProc -->|是| GoToProcess[goToProcess 创建新审批任务]
    CheckBackProc -->|否| ParentNext[进入父流程下一步]
    
    CheckLast -->|否| CreateChild{ChildFlowID > 0?}
    CreateChild -->|是| ChildEntry[查找/创建子流程Entry]
    ChildEntry --> ChildInit[SetFirstProcessAuditor 初始化子流程]
    ChildEntry --> UpdateChild[更新父Entry.child字段]
    CreateChild -->|否| CalcAuditors[GetProcessAuditorIds 计算下一步审批人]
    
    HasBranch --> QueryPVar[查询ProcessVar 获取条件字段名]
    QueryPVar --> QueryEData[查询EntryData 获取字段值]
    QueryPVar --> QueryFLinks[查询所有Condition Flowlink]
    
    QueryFLinks --> LoopConditions[遍历每个条件]
    LoopConditions --> CheckExpr{Expression等于1?}
    CheckExpr -->|是| MatchFound[匹配成功]
    CheckExpr -->|否| ParseJSON[解析JSON为ProcessCondition数组]
    
    ParseJSON --> ValidateOps{操作符白名单校验}
    ValidateOps -->|失败| Err2[返回错误:不支持的操作符]
    ValidateOps -->|通过| EscapeValue[转义值中的单引号]
    EscapeValue --> ValidateField{字段名正则校验}
    ValidateField -->|失败| Err3[返回错误:无效字段名]
    ValidateField -->|通过| BuildSQL[构造参数化SQL查询]
    
    BuildSQL --> ExecSQL[执行SQL: SELECT count FROM entrydatas WHERE ...]
    ExecSQL --> CheckResult{resultCount.Number > 0?}
    CheckResult -->|是| MatchFound
    CheckResult -->|否| NextCondition{还有条件?}
    NextCondition -->|是| LoopConditions
    NextCondition -->|否| Err4[返回错误:未找到匹配条件]
    
    MatchFound --> CalcAuditors
    GoToProcess --> CreateProcs[为每个审批人创建Proc status=0]
    ParentNext --> CreateProcs
    ChildInit --> CreateProcs
    CalcAuditors --> CreateProcs
    
    CreateProcs --> NotifyNext[NotifyNextAuditor 通知下一审批人]
    NotifyNext --> UpdateEntry[更新Entry.process_id]
    UpdateEntry --> MarkDone[标记当前Proc status=1]
    MarkDone --> ExecPlugin[ExecPluginMethod DistributePlugin]
    ExecPlugin --> TriggerCC[triggerCC 创建抄送记录]
    TriggerCC --> End([完成])
    
    Err1 --> End
    Err2 --> End
    Err3 --> End
    Err4 --> End
    EndBoth --> End
    
    style Start fill:#9f9,stroke:#333
    style End fill:#f9f,stroke:#333
    style HasBranch fill:#ff9,stroke:#f90
    style NoBranch fill:#9ff,stroke:#09f
    style ValidateOps fill:#f99,stroke:#900
    style BuildSQL fill:#ff9,stroke:#900
    style CreateProcs fill:#9cf,stroke:#333
    style TriggerCC fill:#fc9,stroke:#333
```

### State Machine

```mermaid
stateDiagram-v2
    [*] --> 待发起: 用户选择流程
    
    待发起 --> 进行中: Store() 创建Entry status=0 初始化Proc
    
    进行中 --> 已完成: Transfer() 最后一步 Entry.status=9
    
    进行中 --> 已驳回: UnPass() 驳回 Entry.status=-1 Proc.status=-1
    
    进行中 --> 已撤回: Revoke() 撤回 Entry.status=-2 Proc.status=-2
    
    进行中 --> 进行中: Pass() 审批通过 创建新Proc 更新Entry.process_id
    
    进行中 --> 进行中: UnPassTo(targetProcessID) 驳回到指定节点
    
    进行中 --> 进行中: AddSign() 加签 新增Proc
    
    进行中 --> 进行中: TransferProc() 转交 原Proc.status=3 新Proc.status=0
    
    进行中 --> 进行中: 会签模式 所有人通过后继续
    
    进行中 --> 进行中: 或签模式 一人通过后其余跳过
    
    进行中 --> 进行中: 条件分支匹配 根据Expression选择Flowlink
    
    进行中 --> 子流程: ChildFlowID>0 创建子Entry Pid=父EntryID
    
    子流程 --> 进行中: 子流程完成 ChildAfter=1或2
    
    子流程 --> 子流程: 子流程内部流转 同主流程逻辑
    
    state 进行中 {
        [*] --> 待处理: Proc.status=0
        待处理 --> 已通过: Pass()/Transfer()
        待处理 --> 已驳回: UnPass()
        待处理 --> 已撤回: Revoke()
        待处理 --> 已转交: TransferProc()
        已通过 --> [*]
        已驳回 --> [*]
        已撤回 --> [*]
        已转交 --> [*]
    }
    
    已完成 --> 重发: Resend() Circle++ status=0
    已驳回 --> 重发
    重发 --> 进行中
    
    note right of 进行中
        可并发操作:
        - 添加评论
        - 查看抄送
        - 发起加签(当前任务)
        - 发起转交(当前任务)
    end note
    
    note right of 子流程
        父子关联:
        - Entry.Pid = 父EntryID
        - Entry.Circle = 父Circle
        - 子流程完成后
          根据ChildAfter决定
          父流程走向
    end note
```

### Approver Calculation Logic

```mermaid
graph TD
    Start([GetProcessAuditorIds entry, next_process_id]) --> QuerySys[查询Flowlink type=Sys AND process_id=?]
    
    QuerySys --> CheckSys{Sys记录存在?}
    
    CheckSys -->|是| CheckAuditor{Auditor值等于?}
    CheckSys -->|否| QueryEmpDept[查询type=Emp和type=Dept]
    
    CheckAuditor -->|-1000| AddInitiator[添加发起人emp_id entry.EmpID]
    CheckAuditor -->|-1001| AddDirector[添加部门主管 entry.Emp.Dept.DirectorID]
    CheckAuditor -->|-1002| AddManager[添加部门经理 entry.Emp.Dept.ManagerID]
    CheckAuditor -->|-1003| QueryForm[从EntryData查询ApproverRule字段值]
    CheckAuditor -->|-1004| DynamicExpr[根据ApproverRule映射键计算]
    CheckAuditor -->|其他| AddDirect[直接添加指定ID]
    
    AddInitiator --> Dedup
    AddDirector --> Dedup
    AddManager --> Dedup
    QueryForm --> Dedup
    DynamicExpr --> Dedup
    AddDirect --> Dedup
    
    QueryEmpDept --> CheckEmp{type为Emp存在?}
    CheckEmp -->|是| SplitEmp[分割逗号ID列表 添加到结果]
    CheckEmp -->|否| CheckDept[检查type=Dept]
    
    SplitEmp --> Dedup
    CheckDept -->|是| SplitDept[分割部门ID列表]
    CheckDept -->|否| ReturnResult
    
    SplitDept --> QueryDeptDir[查询各部门DirectorID]
    QueryDeptDir --> AddDirectors[添加所有director_id]
    AddDirectors --> Dedup
    
    Dedup{uniqueSlice去重} --> ReturnResult[返回去重后的 auditor_ids数组]
    ReturnResult --> End([结束])
    
    style Start fill:#9f9,stroke:#333
    style End fill:#f9f,stroke:#333
    style Dedup fill:#ff9,stroke:#f90
```

### Child Workflow & Parent-Child Linkage

```mermaid
graph LR
    subgraph ParentFlow["父流程 Entry_Parent"]
        P1[步骤A: 部门主管审批 Proc_1 status=0]
        P2[步骤B: 总经理审批 Proc_2 status=0]
        P3[步骤C: 子流程入口 position=2, ChildFlowID=5]
    end
    
    subgraph ChildFlow["子流程 Entry_Child Pid=Entry_Parent"]
        C1[步骤1: 专员填写 Proc_3 status=0]
        C2[步骤2: 专员主管审批 Proc_4 status=0]
        C3[步骤3: 完成]
    end
    
    P3 -->|ChildFlowID=5| C1
    C3 -->|子流程完成| CheckCA{ChildAfter值等于?}
    
    CheckCA -->|1| EndBoth[同时结束父流程 Entry_Parent.status=9]
    CheckCA -->|2| CheckBack{ChildBackProcess?>0?}
    
    CheckBack -->|是| GoBack[回到父流程指定步骤 goToProcess]
    CheckBack -->|否| ParentNext[进入父流程下一步 查询Flowlink]
    
    EndBoth --> PF_end[父流程完成]
    GoBack --> PF_mid[父流程继续]
    ParentNext --> PF_mid
    
    PF_end --> Final([全部结束])
    PF_mid --> Continue[父流程继续流转]
    Continue --> Final
    
    style P3 fill:#ff9,stroke:#f90
    style C3 fill:#9ff,stroke:#09f
    style CheckCA fill:#f9f,stroke:#909
```

### Data Flow Overview

```mermaid
graph LR
    subgraph 输入
        A1[用户提交表单]
        A2[用户点击审批]
        A3[用户点击驳回]
        A4[用户点击撤回]
        A5[用户点击加签]
        A6[用户点击转交]
        A7[用户发送评论]
    end
    
    subgraph 处理
        B1[DynamicValidator]
        B2[Transfer引擎:含会签/或签]
        B3[UnPass/UnPassTo 驳回逻辑]
        B4[Revoke事务]
        B5[AddSign事务]
        B6[TransferProc事务]
        B7[AddComment]
        B8[timeout_check 超时检查]
    end
    
    subgraph 数据写入
        C1[Entry]
        C2[Proc]
        C3[EntryData]
        C4[ProcAddSign]
        C5[CcRecord]
        C6[ProcComment]
    end
    
    subgraph 通知输出
        D1[NotifyNextAuditor]
        D2[NotifySendOne]
        D3[Plugin回调]
    end
    
    A1 --> B1 --> C1 & C2 & C3
    A2 --> B2 --> C2 & C1
    A3 --> B3 --> C2 & C1
    A4 --> B4 --> C1 & C2
    A5 --> B5 --> C2 & C4
    A6 --> B6 --> C2
    A7 --> B7 --> C6
    
    B2 --> D1 & D2 & D3
    B3 --> D2
    B4 --> D2
    
    style B2 fill:#ff9,stroke:#f90,stroke-width:3px
    style C1 fill:#9cf,stroke:#333
    style C2 fill:#9cf,stroke:#333
    style D1 fill:#9f9,stroke:#333
    style D2 fill:#9f9,stroke:#333
```

### Request Lifecycle Sequence

```mermaid
sequenceDiagram
    participant U as 用户
    participant C as Controller
    participant W as Workflow
    participant DB as 数据库
    participant N as 通知钩子
    
    Note over U,DB: 场景1: 发起流程
    U->>C: POST /entry/store {flow_id, title, fields...}
    C->>DB: 查询Flow + Template
    C->>DB: 查询第一步Flowlink (position=0)
    C->>C: DynamicValidator校验表单
    C->>DB: INSERT Entry (status=0)
    C->>W: SetFirstProcessAuditor(entry, flowlink)
    W->>DB: 查询Flowlink type!=Condition
    W->>W: GetProcessAuditorIds()
    W->>DB: 查询Emp列表
    loop 每个审批人
        W->>DB: INSERT Proc (status=0)
    end
    W-->>C: 返回
    C->>DB: INSERT EntryData (每个字段)
    C-->>U: 返回Entry ID
    
    Note over U,DB: 场景2: 审批通过
    U->>C: POST /proc/pass {process_id, content}
    C->>DB: 查询Emp (user_id→id)
    C->>DB: 查询Proc (process_id+emp_id+status=0)
    C->>W: Transfer(process_id, user, content)
    W->>DB: 统计Condition Flowlink数量
    
    alt 有条件分支 (fkcount>1)
        W->>DB: 查询ProcessVar
        W->>DB: 查询EntryData
        W->>DB: 查询所有Condition Flowlink
        loop 每个条件
            W->>W: 白名单校验操作符
            W->>W: 转义值+正则校验字段
            W->>DB: 参数化SQL查询entrydatas
            alt 匹配成功
                W->>DB: 选中该Flowlink
            end
        end
    else 无条件分支
        W->>DB: 查询Flowlink
        alt NextProcessID=-1 (最后一步)
            W->>DB: UPDATE Entry status=9
            alt 有父流程
                W->>DB: 检查ChildAfter
                W->>N: NotifySendOne(发起人)
            end
        else ChildFlowID>0 (子流程)
            W->>DB: 查找/创建子Entry
            W->>W: SetFirstProcessAuditor(child)
            W->>DB: 更新父Entry.child
        else 正常下一步
            W->>W: GetProcessAuditorIds()
            W->>DB: 查询Emp列表
            loop 每个审批人
                W->>DB: INSERT Proc (status=0)
                W->>N: NotifyNextAuditor(emp_id)
            end
            W->>DB: UPDATE Entry process_id
        end
    end
    
    W->>DB: UPDATE Proc status=1
    W->>DB: 查询PluginConfig
    W->>W: ExecPluginMethod(DistributePlugin)
    W->>W: triggerCC(entryID, flowID, processID, procID)
    W->>DB: 查询Process.cc_emp_ids
    loop 每个抄送人
        W->>DB: INSERT CcRecord (status=0)
    end
    W-->>C: 返回
    C-->>U: 返回"审批成功"
    
    Note over U,DB: 场景3: 撤回
    U->>C: POST /entry/revoke {entry_id}
    C->>DB: 查询Emp
    C->>W: Revoke(entryID, user) [事务开始]
    W->>DB: 查询Entry+Emp
    W->>W: 校验: 发起人? 状态=0?
    W->>DB: 查询pending Procs
    W->>W: 校验: auditor_id=0?
    alt 校验通过
        W->>DB: UPDATE Entry status=-2
        loop 每个pending Proc
            W->>DB: UPDATE Proc status=-2
            W->>DB: UPDATE Proc auditor_id/name
        end
        W->>DB: COMMIT
    else 校验失败
        W->>DB: ROLLBACK
        W-->>C: 返回错误信息
    end
    C-->>U: 返回结果
```

### Transaction Boundaries

```mermaid
graph TD
    subgraph Transfer事务["Transfer() 非事务包裹 部分子操作使用独立事务"]
        T1[查询Proc + Entry 非事务]
        T2[条件分支解析 非事务]
        T3[创建新Proc 非事务]
        T4[更新Entry.status 非事务]
        T5[标记当前Proc完成 非事务]
        T6[ExecPlugin 非事务]
        T7[triggerCC 非事务]
    end
    
    subgraph Revoke事务["Revoke() 完整事务"]
        R1[查询Entry+Emp]
        R2[校验: 发起人+状态]
        R3[查询pending Procs]
        R4[更新Entry.status=-2]
        R5[批量更新Procs.status=-2]
        R6[COMMIT]
    end
    
    subgraph AddSign事务["AddSign() 完整事务"]
        S1[查询Entry]
        S2[校验: 状态]
        S3[查询targetProc]
        S4[查询signEmp]
        S5[创建ProcAddSign]
        S6[创建新Proc]
        S7[COMMIT]
    end
    
    subgraph TransferProc事务["TransferProc() 完整事务"]
        TP1[查询targetProc]
        TP2[校验: 存在+状态]
        TP3[查询Entry]
        TP4[校验: 流程状态]
        TP5[查询targetEmp]
        TP6[创建新Proc]
        TP7[更新原Proc.status=3]
        TP8[COMMIT]
    end
    
    subgraph SetFirstProcessAuditor事务["SetFirstProcessAuditor 完整事务"]
        F1[查询Flowlink]
        F2[计算auditor_ids]
        F3[查询Emp列表]
        F4[批量创建Procs]
        F5[更新Entry.process_id]
        F6[COMMIT]
    end
    
    T1 --> T2 --> T3 --> T4 --> T5 --> T6 --> T7
    R1 --> R2 --> R3 --> R4 --> R5 --> R6
    S1 --> S2 --> S3 --> S4 --> S5 --> S6 --> S7
    TP1 --> TP2 --> TP3 --> TP4 --> TP5 --> TP6 --> TP7 --> TP8
    F1 --> F2 --> F3 --> F4 --> F5 --> F6
    
    style R6 fill:#9f9,stroke:#333
    style S7 fill:#9f9,stroke:#333
    style TP8 fill:#9f9,stroke:#333
    style F6 fill:#9f9,stroke:#333
    style T7 fill:#ff9,stroke:#f90
```

### Troubleshooting Guide

#### 审批人计算问题
1. 检查 Flowlink 表中目标 ProcessID 的记录是否存在
2. 确认 Type 字段（Sys/Emp/Dept）是否正确
3. 如果是 Auditor=-1003/-1004，检查 ApproverRule 字段值与 EntryData 是否匹配
4. 检查 Emp 表中对应 ID 的记录是否存在，以及 DeptID 是否正确
5. 查看 `services/workflow/workflow.go:268` 的 `GetProcessAuditorIds` 方法日志

#### 流程无法流转
1. 检查 Entry.Status 是否为 0（进行中）
2. 检查 Proc 是否存在且 status=0（待处理），emp_id 是否对应当前用户
3. 检查 Flowlink.NextProcessID 是否正确设置
4. 如果有条件分支，检查 ProcessVar.ExpressionField 是否与 EntryData.field_name 匹配
5. 检查 Condition Flowlink 的 Expression 是否为合法 JSON

#### 会签/或签异常
1. 会签：检查同一步骤下 Proc 数量是否与 Flowlink 定义的审批人数量一致
2. 或签：检查被跳过的 Proc 是否正确标记为 status=4
3. 查看 `checkConsensusComplete()` 和 `skipRemainingConcurrentProcs()` 方法

#### 子流程问题
1. 检查 Process.Position 是否为 2
2. 检查 Process.ChildFlowID 是否正确指向子 Flow
3. 检查 Entry.Pid 是否正确关联父子流程
4. 检查 ChildAfter 和 ChildBackProcess 配置

#### 超时检查不生效
1. 确认 Process.LimitTime > 0
2. 确认 Proc.Concurrence 有时间值
3. 确认 `workflow:timeout-check` 命令已正确注册并定期执行
4. 检查 commands/timeout_check.go 的实现

#### Hook 未触发
1. 确认在 AppServiceProvider.Boot() 中调用了 RegisterHook
2. 确认 Hook 名称与调用时使用的名称完全一致
3. 确认方法签名是否为 func(uint) error
4. 查看 services/workflow/hook.go 和 invokeHooks 方法的日志输出

#### 路由冲突
1. 启动项目时报路由重复错误，检查目标项目的 routes/api.go
2. 工作流注册的路由见 routes/api.go
3. 可能需要修改目标项目的路由前缀或工作流的 Controller 前缀
