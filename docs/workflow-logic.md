# 审批流工作引擎 — 逻辑架构文档

## 一、整体架构概览

本系统是一个基于 Goravel 框架的审批流扩展包，核心围绕"流程（Flow）→ 步骤（Process）→ 实例（Entry）→ 任务（Proc）"四层模型构建。所有审批操作通过 `services/workflow/workflow.go` 中的 `Workflow` 单例执行，采用状态机模式管理流程生命周期。

```
Flow (流程) → Process (步骤) → Entry (流程实例) → Proc (审批任务)
```

---

## 二、数据模型

### 2.1 Flow（流程）

| 字段 | 说明 |
|------|------|
| `FlowNo` | 流程编号 |
| `FlowName` | 流程名称 |
| `TemplateID` | 关联的表单模板 ID |
| `IsPublish` | 是否已发布 |
| `Jsplumb` | 流程图 JSON 数据（jsPlumb 可视化） |

一个 Flow 包含多个 Process 和 Flowlink。

### 2.2 Process（步骤）

| 字段 | 说明 |
|------|------|
| `ProcessName` | 步骤名称 |
| `Position` | 步骤位置：0=第一步，1=正常步骤，2=转入子流程 |
| `ChildFlowID` | 当 position=2 时，子流程 ID |
| `ChildAfter` | 子流程结束后：1=同时结束父流程，2=返回父流程 |
| `ChildBackProcess` | 子流程返回时进入的父流程步骤 ID |
| `LimitTime` | 步骤限时（秒） |
| `CcEmpIDs` | 抄送人 ID 列表（逗号分隔） |
| `AutoPerson` | 自动审批人设置 |

### 2.3 Flowlink（步骤流转关系）

| 字段 | 说明 |
|------|------|
| `Type` | 类型：`Condition`=条件流转，`Emp`=指定员工，`Dept`=指定部门，`Sys`=系统自动 |
| `Auditor` | 审批人设置（-1000=发起人，-1001=主管，-1002=经理，-1003=表单字段，-1004=动态表达式，或具体人员/部门 ID 列表） |
| `ApproverRule` | 审批人分配规则：-1003 时存储表单字段名，-1004 时存储动态表达式映射键（如 "director"/"manager"） |
| `ConcurrencyType` | 并签模式：0=依次（默认），1=会签（所有人通过才进入下一步），2=或签（一人通过即进入下一步，其余跳过） |
| `Expression` | 条件表达式（JSON 格式），值为 "1" 表示无条件通过 |
| `Sort` | 判断顺序 |
| `NextProcessID` | 下一步骤 ID |

### 2.4 Entry（流程实例）

| 字段 | 说明 |
|------|------|
| `Title` | 实例标题 |
| `FlowID` | 所属流程 |
| `EmpID` | 发起人 ID |
| `Status` | 状态：0=进行中，9=已完成，-1=已驳回，-2=已撤回 |
| `Pid` | 父实例 ID（子流程标记） |
| `Circle` | 轮次（重发时递增） |
| `ProcessID` | 当前所在步骤 |
| `EnterProcessID` | 进入步骤 ID |
| `EnterProcID` | 进入任务 ID |
| `Child` | 子流程当前步骤 |

### 2.5 Proc（审批任务）

| 字段 | 说明 |
|------|------|
| `EntryID` | 所属实例 |
| `ProcessID` | 当前步骤 |
| `EmpID` | 被指派审批人 |
| `Status` | 状态：0=待处理，1=已通过，-1=已驳回，-2=已撤回，3=已转交，4=已跳过，9=会签通过 |
| `Content` | 批复内容 |
| `Concurrence` | 并发时间（用于并行审批查找） |
| `IsRead` | 是否已查看 |
| `IsReal` | 审核人和操作人是否同一人 |
| `UnpassTargetProcessID` | 驳回到指定节点的目标步骤 ID |

### 2.6 辅助模型

- **ProcComment** — 评论：记录审批过程中的讨论
- **CcRecord** — 抄送记录：审批通过后自动抄送给相关人员
- **ProcAddSign** — 加签记录：前加签/后加签
- **EntryData** — 流程表单数据
- **ProcessVar** — 步骤变量（用于条件判断）

---

## 三、核心流程

### 3.1 发起流程（Entry.Store）

**入口：** `entry_controller.go → Store()`

**详细步骤：**

1. **获取流程信息**
   - 根据 `flow_id` 查询 Flow，加载关联的 Template 和 TemplateForms

2. **定位起始节点**
   - 在 Flowlink 表中查找 `position=0` 的第一步（按 sort 排序取第一条）
   - 加载 NextProcess 信息

3. **表单校验**
   - 使用 `DynamicValidator` 根据 Template 动态生成校验规则
   - 对提交数据进行验证，失败则返回校验错误

4. **创建流程实例**
   - 在 `Entry` 表中创建新记录，初始化字段：
     - `Title` = 提交的标题
     - `FlowID` = 流程 ID
     - `EmpID` = 当前用户 ID
     - `Status` = 0（进行中）
     - `Circle` = 1
   - 加载完整关联数据（Flow、Emp.Dept、Procs、EnterProcess）

5. **初始化审批链**
   - 调用 `workflow.SetFirstProcessAuditor(entry, flowlink)`：
     a. 查询该步骤的所有 Flowlink（type != Condition）
     b. 如果未找到（ID=0），说明第一步无指定审核人，自动创建一条状态为 9（通过）的 Proc，然后进入下一步
     c. 通过 `GetProcessAuditorIds()` 计算下一步审批人列表
     d. 查询审批人对应的 Emp 记录
     e. 为每个审批人创建 Proc 任务（status=0 待处理）
     f. 更新 Entry 的 ProcessID

6. **保存表单数据**
   - 遍历所有提交字段（排除 title 和 flow_id）
   - 如果是数组类型，拼接为字符串后存入 EntryData
   - 否则直接存入 EntryData

**关键设计：**
- `SetFirstProcessAuditor` 内部使用事务保证原子性
- 第一步若未指定审核人，会自动跳过并进入下一步

---

### 3.2 审批流转（Workflow.Transfer）

**入口：** `proc_controller.go → Pass()` → `workflow.Pass()` → `workflow.Transfer()`

**详细步骤：**

1. **查找当前任务**
   - 根据 `process_id` 和当前用户 ID，在 Proc 表中查找 status=0 的任务
   - 加载关联数据（Entry.Emp.Dept、Entry.ParentEntry）
   - 如果找不到任务，返回错误

2. **检查并签模式**
   - 查询当前步骤的 Flowlink，获取 `ConcurrencyType`：
     - **0=依次（默认）**：正常顺序流转
     - **1=会签**：所有审批人都需审批，全部通过才进入下一步，任一驳回则整个步骤驳回
     - **2=或签**：第一个审批人通过即进入下一步，其余审批人自动跳过（status=4）

3. **会签逻辑（ConcurrencyType=1）**
   - 统计该步骤下所有 Proc 数量
   - 统计已通过和已驳回的 Proc 数量
   - 如果所有人已完成：
     - 有驳回 → 调用 `handleRejectEntry()` 驳回整个流程
     - 全部通过 → 继续后续流转
   - 如果还有人未审批 → 仅标记当前 Proc 为已通过，返回等待其他人

4. **或签逻辑（ConcurrencyType=2）**
   - 标记同一步骤下所有其他待处理 Proc 为已跳过（status=4）
   - 继续后续流转

5. **判断是否有条件分支**
   - 查询该步骤的 Flowlink 中 type=Condition 的数量

#### 情况一：有条件分支（fkcount > 1）

3a. **解析条件表达式**
   - 从 ProcessVar 表获取当前步骤的变量定义
   - 从 EntryData 表查询条件字段的实际值
   - 遍历所有 Condition 类型的 Flowlink：
     - 如果 Expression="1"，无条件匹配，直接选中
     - 否则，将 Expression（JSON 格式的 ProcessCondition 数组）解析为条件列表
     - 对每个条件进行校验：
       - 字段名必须匹配当前 ProcessVar 定义的 ExpressionField
       - 操作符必须在白名单内（=, !=, >, <, >=, <=, like, in, not in, between）
       - 值中的单引号会被转义
       - 字段名必须只包含字母、数字、下划线
     - 构造 SQL 查询 entrydatas 表，检查是否有记录满足条件
     - 使用参数化查询（? 占位符）防止 SQL 注入
     - 如果找到匹配的记录，选中该 Flowlink

3b. **计算下一步审批人**
   - 通过 `GetProcessAuditorIds()` 获取下一步审批人列表

3c. **创建新任务**
   - 为每个审批人创建 Proc 任务（status=0）
   - 通知下一个审批人（NotifyNextAuditor）
   - 更新 Entry 的 ProcessID

#### 情况二：无条件分支（fkcount <= 1）

3d. **检查是否为最后一步**
   - 如果 NextProcessID=-1，说明是最后一步：
     - 更新 Entry 状态为 9（已完成）
     - 如果有父流程且 ChildAfter=1，同时结束父流程并通知发起人
     - 如果 ChildAfter=2，进入父流程的设置步骤或下一步
   - 否则：
     - 计算下一步审批人
     - 创建新 Proc 任务
     - 通知下一个审批人
     - 更新 Entry 的 ProcessID

3e. **处理子流程**
   - 如果当前步骤的 ChildFlowID > 0：
     - 查找或创建子流程 Entry
     - 查询子流程的起始 Flowlink
     - 调用 `SetFirstProcessAuditor()` 初始化子流程审批链
     - 更新父流程 Entry 的 child 字段

4. **标记当前任务完成**
   - 将当前 Proc 的 status 更新为 1（已通过）
   - 填充 AuditorID、AuditorName、Content 等字段
   - 序列化 PluginConfig 到 Beizhu 字段

5. **执行插件**
   - 调用 `ExecPluginMethod("DistributePlugin", flowID, processID)`

6. **触发抄送**
   - 调用 `triggerCC()`，根据当前步骤的 CcEmpIDs 创建抄送记录

**关键设计：**
- 整个流转过程在同一个事务外执行（部分子操作使用独立事务）
- 条件分支使用白名单 + 参数化查询防护 SQL 注入
- 父子流程通过 Pid 字段关联

---

### 3.3 驳回（Workflow.UnPass / UnPassTo）

**入口：** `proc_controller.go → UnPass()` → `workflow.UnPass()` / `workflow.UnPassTo(proc_id, user, content, targetProcessID)`

#### 普通驳回（UnPass — 驳回到上一步）

1. 查询当前用户对应的 Emp 记录
2. 根据 proc_id 查找 Proc 任务，加载 Entry 信息
3. 调用 `UnPassTo(proc_id, user, content, 0)`，targetProcessID=0 触发普通驳回逻辑
4. 查找同一 Entry 下同一 Process 下 status=0 的待处理任务
5. 如果找不到待处理任务 → 直接驳回整个 Entry
6. 将待处理任务标记为 status=-1（已驳回）
7. 更新 Entry 状态为 -1（已驳回）
8. 如果有父流程，同步更新父流程状态
9. 通知发起人（NotifySendOne）

#### 驳回到指定节点（UnPassTo — 任意节点驳回）

1. 查询当前用户对应的 Emp 记录
2. 根据 proc_id 查找 Proc 任务
3. 根据 targetProcessID 查找目标 Proc：
   - 如果目标步骤有待处理 Proc → 直接使用
   - 如果不存在 → 查询所有 Proc（按 ID 倒序取最新一条）
4. 如果目标 Proc 不存在 → 返回错误
5. 将目标步骤之后（按 ID 排序）的所有 Proc 标记为已跳过（status=4）
6. 重置目标 Proc 为待处理状态（status=0）
7. 更新 Entry 状态为进行中，ProcessID 指向目标步骤
8. 通知目标审批人（NotifyNextAuditor）和发起人（NotifySendOne）

---

### 3.4 撤回（Workflow.Revoke）

**入口：** `proc_controller.go / entry_controller.go → Revoke()` → `workflow.Revoke()`

**详细步骤：**

1. **事务开始**
2. 查询 Entry 并加载 Emp 信息
3. **校验：**
   - Entry 是否存在
   - 当前用户是否是发起人（entry.EmpID == user.ID）
   - Entry 状态是否为 0（进行中）
   - 是否存在已处理的待办任务（auditor_id != 0）
4. 将 Entry 状态更新为 -2（已撤回）
5. 将所有 pending 的 Proc 任务标记为 status=-2，填充操作人信息
6. **事务提交**

**关键设计：**
- 只有发起人能撤回
- 只有未被任何人处理过的流程才能撤回
- 使用事务保证 Entry 和 Proc 状态一致性

---

### 3.5 加签（Workflow.AddSign）

**入口：** `proc_controller.go → AddSign()` → `workflow.AddSign()`

**详细步骤：**

1. **事务开始**
2. 查询 Entry，校验存在性和状态（必须为 0）
3. 查找当前用户在该 Entry 下的待处理 Proc 任务
4. 查询被加签的 Emp 信息
5. 创建 ProcAddSign 记录（sign_type: before=前加签, after=后加签）
6. 创建新的 Proc 任务给被加签人（status=0）
7. **事务提交**

---

### 3.6 转交（Workflow.TransferProc）

**入口：** `proc_controller.go → TransferProc()` → `workflow.TransferProc()`

**详细步骤：**

1. **事务开始**
2. 查询目标 Proc 任务，校验存在性
3. **校验：**
   - Proc 是否属于当前 Entry
   - Proc 状态是否为 0（待处理）
   - Entry 状态是否为 0（进行中）
4. 查询被转交人的 Emp 信息
5. 创建新的 Proc 任务给被转交人
6. 将原 Proc 标记为 status=3（已转交），填充转交内容
7. **事务提交**

---

### 3.7 评论（Workflow.AddComment / GetComments）

**入口：** `proc_controller.go → AddComment() / GetComments()`

**AddComment 步骤：**
1. 创建 ProcComment 记录，包含 EntryID、ProcID、EmpID、EmpName、Content
2. 状态默认为 1（正常）

**GetComments 步骤：**
1. 查询指定 EntryID 下 status=1 的所有评论
2. 按 ID 升序排列

---

### 3.8 抄送（Workflow.triggerCC）

**入口：** `workflow.Transfer()` 完成后自动调用

**详细步骤：**

1. 查询当前步骤（Process）的 CcEmpIDs 字段
2. 如果没有抄送人，直接返回
3. 查询 Entry 信息
4. 查询所有抄送人对应的 Emp 记录
5. 为每个抄送人创建 CcRecord 记录（status=0 未读）

---

### 3.9 获取审批人（Workflow.GetProcessAuditorIds）

**详细步骤：**

1. 查询 Flowlink 中 type=Sys 的记录（系统自动审批人设置）
2. 根据 Auditor 字段值分类处理：
   - `-1000`：发起人自己
   - `-1001`：发起人部门主管（DirectorID）
   - `-1002`：发起人部门经理（ManagerID）
   - `-1003`（表单字段指定）：从 `ApproverRule` 字段获取表单字段名，查询 EntryData 表获取该字段的值作为审批人 ID
   - `-1004`（动态表达式）：使用 `ApproverRule` 作为映射键：
     - `"director"` → 部门主管
     - `"manager"` → 部门经理
     - 其他数字 → 直接作为 ID
3. 如果没有 Sys 类型，继续查询：
   - `type=Emp`：指定员工（逗号分隔的 ID 列表）
   - `type=Dept`：指定部门（查询各部门的 DirectorID）
4. 使用 `uniqueSlice()` 去重后返回

---

### 3.10 第一步审批人初始化（Workflow.SetFirstProcessAuditor）

**入口：** `Entry.Store()` 和 `Entry.Resend()` 中调用

**详细步骤：**

1. 开启事务
2. 查询该 Flowlink（type != Condition）
3. 如果未找到（ID=0），说明第一步无指定审核人：
   - 自动创建一条 status=9（通过）的 Proc
   - 进入下一步骤
4. 通过 `GetProcessAuditorIds()` 计算审批人列表
5. 查询审批人对应的 Emp 记录
6. 为每个审批人创建 Proc 任务（status=0）
7. 更新 Entry 的 ProcessID
8. 事务提交

---

## 四、辅助功能

### 4.1 钩子系统（Hook System）

- `RegisterHook(name, method)`：注册命名钩子，支持同一名称注册多个函数
- `invokeHooks(hookName, id)`：按名称调用所有注册的钩子函数
- 预置钩子：`NotifySendOneHook`（通知发起人）、`NotifyNextAuditorHook`（通知下一审批人）
- 使用 reflect.Value 存储函数引用，调用时进行签名校验

### 4.2 单例模式

- `NewBaseWorkflow()` 使用 sync.Once 确保全局唯一实例
- hooks map 在初始化时自动创建
- nil 接收者会返回错误而非 panic

### 4.3 定时超时检查命令

- `workflow:timeout-check` 命令行工具（通过 `commands/timeout_check.go` 实现）
- 查询所有 status=0（待处理）的 Proc 任务
- 遍历每个 Proc：
  - 查询对应 Process 的 `LimitTime`（秒）
  - 如果 LimitTime <= 0，跳过该步骤
  - 计算 `Concurrence`（审批创建时间）与当前时间的差值（秒）
  - 如果超过 LimitTime → 标记 Proc 为已驳回（status=-1），Content="超时未处理，系统自动驳回"
  - 同步更新 Entry 状态为已驳回（status=-1）
  - 打印日志记录超时详情
- **部署方式**：需通过系统 cron 或 Goravel scheduler 定期触发（如每 5 分钟）

---

## 五、状态码汇总

### Entry 状态
| 值 | 含义 |
|----|------|
| 0 | 进行中 |
| 9 | 已完成 |
| -1 | 已驳回 |
| -2 | 已撤回 |

### Proc 状态
| 值 | 含义 |
|----|------|
| 0 | 待处理 |
| 1 | 已通过 |
| -1 | 已驳回 |
| -2 | 已撤回 |
| 3 | 已转交 |
| 4 | 已跳过（或签中未被选中的审批人） |
| 9 | 会签通过（第一步无指定审核人时自动通过） |

### Flowlink.ConcurrencyType
| 值 | 含义 |
|----|------|
| 0 | 依次审批（默认） |
| 1 | 会签：所有审批人都需通过 |
| 2 | 或签：一人通过即进入下一步，其余跳过 |

### Flowlink.Auditor 特殊值
| 值 | 含义 |
|----|------|
| -1000 | 发起人自己 |
| -1001 | 部门主管 |
| -1002 | 部门经理 |
| -1003 | 从表单字段读取审批人（ApproverRule 存储字段名） |
| -1004 | 动态表达式计算审批人（ApproverRule 存储映射键） |

### Flowlink.Type
| 值 | 含义 |
|----|------|
| Condition | 条件流转 |
| Emp | 指定员工 |
| Dept | 指定部门 |
| Sys | 系统自动 |

---

## 六、安全与工程化改进

### 6.1 SQL 注入防护
- **操作符白名单**：条件表达式中的操作符仅允许 `=, !=, >, <, >=, <=, like, in, not in, between`
- **值转义**：值中的单引号会被转义为 `\'`
- **字段名正则校验**：字段名必须只包含字母、数字、下划线（`^[a-zA-Z0-9_]+$`）
- **参数化查询**：SQL 使用 `?` 占位符，防止 SQL 注入

### 6.2 并发安全
- **RWMutex**：钩子系统使用 `sync.RWMutex`，允许多个 goroutine 同时读取 hooks map
- **nil 安全**：Workflow 方法在 nil 接收者时返回错误而非 panic

### 6.3 事务保护
- **Revoke**：完整事务，批量更新 Entry 和所有 pending Proc
- **AddSign**：完整事务，创建 ProcAddSign 和新 Proc
- **TransferProc**：完整事务，创建新 Proc 并标记原 Proc
- **SetFirstProcessAuditor**：使用外层事务，避免嵌套事务问题

### 6.4 性能优化
- **索引优化**：为 procs(status, entry_id)、procs(emp_id, status)、entrydatas(entry_id, field_name)、entry(status)、entry(flow_id, status) 添加复合索引
- **单次 UPDATE**：Revoke 中所有 pending Procs 的更新在循环中完成，避免重复 SQL
