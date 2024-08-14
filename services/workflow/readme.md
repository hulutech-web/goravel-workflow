### goravel流程审批框架
#### 流程流转
```go
func (w *Workflow) Transfer(process_id int, user models.User) error{}
```
#### 设置头节点,设置第一个审批人
```go
func (w *Workflow) SetFirstProcessAuditor(entry models.Entry, flowlink models.Flowlink) error{}
```
#### 获取审批人
```go
func (w *Workflow) GetProcessAuditorIds(entry models.Entry, next_process_id int) []int{}
```
#### 发送通知
```go
func (w *Workflow) Notify(proc models.Proc) error{}
```