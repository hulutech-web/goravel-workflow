package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

// M20240624000000CreateWorkflowBaseTables creates all foundational workflow tables
// and adds feature columns (concurrency, approver rules, arbitrary-node reject, CC,
// add-sign, comments, cc-records, performance indexes).
// Note: users are assumed to be provided by the host application.
type M20240624000000CreateWorkflowBaseTables struct{}

func (r *M20240624000000CreateWorkflowBaseTables) Signature() string {
	return "20240624000000_create_workflow_base_tables"
}

func (r *M20240624000000CreateWorkflowBaseTables) Up() error {
	// --- depts table ---
	if err := facades.Schema().Create("depts", func(table schema.Blueprint) {
		table.ID()
		table.String("dept_name").Default("").Comment("部门名称")
		table.BigInteger("pid").Default(0).Comment("父级部门ID")
		table.Integer("director_id").Default(0).Comment("部门主管ID")
		table.Integer("manager_id").Default(0).Comment("部门经理ID")
		table.Integer("rank").Default(1).Comment("排序")
		table.String("html").Default("").Nullable().Comment("树形结构HTML")
		table.Integer("level").Default(0).Nullable().Comment("层级")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- emps table ---
	if err := facades.Schema().Create("emps", func(table schema.Blueprint) {
		table.ID()
		table.String("name").Comment("姓名")
		table.String("email").Comment("邮箱")
		table.String("password").Comment("密码")
		table.String("workno").Comment("工号")
		table.Integer("dept_id").Default(0).Comment("部门ID")
		table.Integer("leave").Default(0).Comment("离职状态")
		table.BigInteger("user_id").Default(0).Nullable().Comment("关联用户ID")
		table.Timestamps()
		table.Index("email")
		table.Index("workno")
	}); err != nil {
		return err
	}

	// --- flows table ---
	if err := facades.Schema().Create("flows", func(table schema.Blueprint) {
		table.ID()
		table.String("flow_no").Comment("工作流编号")
		table.String("flow_name").Default("").Comment("工作流名称")
		table.BigInteger("template_id").Default(0)
		table.Text("flowchart").Nullable()
		table.Text("jsplumb").Comment("jsplumb流程图数据")
		table.BigInteger("type_id").Default(0).Comment("流程设计文件")
		table.TinyInteger("is_publish").Default(0).Comment("是否发布，发布后可用")
		table.TinyInteger("is_show").Default(1).Comment("是否显示")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- flowtypes table ---
	if err := facades.Schema().Create("flowtypes", func(table schema.Blueprint) {
		table.ID()
		table.String("type_name").Default("")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- processes table ---
	if err := facades.Schema().Create("processes", func(table schema.Blueprint) {
		table.ID()
		table.BigInteger("flow_id").Default(0).Comment("流程id")
		table.String("process_name").Default("").Comment("步骤名称")
		table.Integer("limit_time").Default(0).Comment("限定时间,单位秒")
		table.String("type").Default("operation").Comment("流程图显示操作框类型")
		table.String("icon").Default("").Comment("流程图显示图标")
		table.String("process_to").Default("")
		table.Text("style").Nullable()
		table.String("style_color").Default("#78a300")
		table.SmallInteger("style_height").Default(30)
		table.SmallInteger("style_width").Default(30)
		table.String("position_left").Default("100px")
		table.String("position_top").Default("200px")
		table.SmallInteger("position").Default(1).Comment("步骤位置：1正常步骤 2转入子流程 0第一步")
		table.BigInteger("child_flow_id").Default(0).Comment("子流程id")
		table.TinyInteger("child_after").Default(2).Comment("子流程结束后 1同时结束父流程 2返回父流程")
		table.BigInteger("child_back_process").Default(0).Comment("子流程结束后返回父流程进程")
		table.String("description").Default("").Comment("步骤描述")
		table.Text("cc_emp_ids").Nullable().Comment("抄送人ID列表,逗号分隔")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- processvars table ---
	if err := facades.Schema().Create("processvars", func(table schema.Blueprint) {
		table.ID()
		table.BigInteger("process_id")
		table.BigInteger("flow_id").Comment("流程id")
		table.String("expression_field").Comment("条件表达式字段名称")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- templates table ---
	if err := facades.Schema().Create("templates", func(table schema.Blueprint) {
		table.ID()
		table.String("template_name").Default("")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- templateforms table ---
	if err := facades.Schema().Create("templateforms", func(table schema.Blueprint) {
		table.ID()
		table.Integer("template_id").Default(0)
		table.String("field").Default("").Comment("表单字段英文名")
		table.String("field_name").Default("").Comment("表单字段中文名")
		table.String("field_type").Default("").Comment("表单字段类型")
		table.Text("field_value").Nullable().Comment("表单字段值，select radio checkbox用")
		table.Text("field_default_value").Nullable().Comment("表单字段默认值")
		table.Text("field_rules").Nullable()
		table.Integer("sort").Default(100).Comment("排序")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- entries table ---
	if err := facades.Schema().Create("entries", func(table schema.Blueprint) {
		table.ID()
		table.String("title").Default("").Comment("标题")
		table.BigInteger("flow_id").Default(0)
		table.BigInteger("emp_id").Default(0).Comment("发起人")
		table.Integer("process_id").Default(0).Comment("当前步骤id")
		table.SmallInteger("circle").Default(1).Comment("第几轮申请")
		table.Integer("status").Comment("当前状态 0处理中 9通过 -1驳回 -2撤销 -9草稿")
		table.BigInteger("pid").Default(0).Comment("父流程")
		table.BigInteger("enter_process_id").Default(0).Comment("进入子流程的父流程步骤id")
		table.BigInteger("enter_proc_id").Default(0).Comment("进入子流程的进程id")
		table.BigInteger("child").Default(0).Comment("子流程 process_id")
		table.Timestamps()
		table.Index("flow_id")
		table.Index("emp_id")
		table.Index("process_id")
	}); err != nil {
		return err
	}

	// --- entrydatas table ---
	if err := facades.Schema().Create("entrydatas", func(table schema.Blueprint) {
		table.ID()
		table.BigInteger("entry_id").Default(0)
		table.BigInteger("flow_id").Default(0)
		table.String("field_name").Default("")
		table.Text("field_value").Nullable()
		table.String("field_remark").Default("")
		table.Timestamps()
		table.Index("entry_id")
		table.Index("flow_id")
	}); err != nil {
		return err
	}

	// --- procs table ---
	if err := facades.Schema().Create("procs", func(table schema.Blueprint) {
		table.ID()
		table.BigInteger("entry_id")
		table.BigInteger("flow_id").Comment("流程id")
		table.BigInteger("process_id").Comment("当前步骤")
		table.String("process_name").Default("").Comment("当前步骤名称")
		table.BigInteger("emp_id").Comment("审核人")
		table.String("emp_name").Nullable().Comment("审核人名称")
		table.String("dept_name").Nullable().Comment("审核人部门名称")
		table.BigInteger("auditor_id").Default(0).Comment("具体操作人")
		table.String("auditor_name").Default("").Comment("操作人名称")
		table.String("auditor_dept").Default("").Comment("操作人部门")
		table.Integer("status").Comment("当前处理状态 0待处理 1已通过 9会签 -1驳回 -2撤回")
		table.Integer("unpass_target_process_id").Default(0).Comment("驳回到指定节点的目标步骤ID")
		table.String("content").Nullable().Comment("批复内容")
		table.Integer("is_read").Default(0).Comment("是否查看")
		table.TinyInteger("is_real").Default(1).Comment("审核人和操作人是否同一人")
		table.SmallInteger("circle").Default(1)
		table.Text("beizhu").Comment("备注")
		table.DateTime("concurrence").Nullable().Comment("并发时间")
		table.Timestamps()
		table.Index("entry_id")
		table.Index("flow_id")
		table.Index("emp_id")
		table.Index("process_id")
	}); err != nil {
		return err
	}

	// --- flowlinks table ---
	if err := facades.Schema().Create("flowlinks", func(table schema.Blueprint) {
		table.ID()
		table.BigInteger("flow_id").Default(0)
		table.BigInteger("process_id").Default(0).Comment("当前步骤id")
		table.BigInteger("next_process_id").Default(0).Comment("下一个步骤id")
		table.Integer("condition").Default(0).Comment("流转条件 0无条件 1有条件")
		table.Text("condition_expr").Nullable().Comment("条件表达式")
		table.String("condition_source").Default("").Comment("条件来源 前端/后端")
		table.TinyInteger("is_deleted").Default(0).Comment("逻辑删除")
		table.TinyInteger("concurrency_type").Default(0).Comment("并签模式: 0=依次, 1=会签, 2=或签")
		table.String("approver_rule").Default("").Comment("审批人分配规则: -1003=表单字段, -1004=动态表达式")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- attachments table ---
	if err := facades.Schema().Create("attachments", func(table schema.Blueprint) {
		table.ID()
		table.String("path").Comment("文件路径")
		table.String("name").Comment("文件名")
		table.String("ext").Nullable().Comment("扩展名")
		table.Char("type", 20).Nullable().Comment("上传方式local,oss")
		table.UnsignedBigInteger("user_id").Nullable()
		table.Integer("size").Default(0).Unsigned().Comment("文件大小")
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- products table ---
	if err := facades.Schema().Create("products", func(table schema.Blueprint) {
		table.ID()
		table.String("name")
		table.String("special")
		table.String("dimension")
		table.Integer("quantity").Unsigned()
		table.String("unit")
		table.Float("unit_price")
		table.Float("discount_price")
		table.Float("amount")
		table.Text("description").Nullable()
		table.String("image_url").Nullable()
		table.Timestamps()
	}); err != nil {
		return err
	}

	// --- proc_add_signs table ---
	if err := facades.Schema().Create("proc_add_signs", func(table schema.Blueprint) {
		table.ID()
		table.BigInteger("entry_id").Default(0)
		table.BigInteger("proc_id").Default(0).Comment("原审批任务ID")
		table.String("sign_type").Default("before").Comment("前加签/后加签")
		table.BigInteger("sign_emp_id").Default(0).Comment("被加签员工ID")
		table.String("sign_emp_name").Default("").Comment("被加签员工名称")
		table.Integer("status").Default(0).Comment("0待处理 1已完成")
		table.Timestamps()
		table.Index("entry_id").Name("idx_signs_entry_id")
		table.Index("proc_id").Name("idx_signs_proc_id")
		table.Index("status").Name("idx_signs_status")
	}); err != nil {
		return err
	}

	// --- proc_comments table ---
	if err := facades.Schema().Create("proc_comments", func(table schema.Blueprint) {
		table.ID()
		table.BigInteger("entry_id")
		table.BigInteger("proc_id")
		table.BigInteger("emp_id").Comment("发言人ID")
		table.String("emp_name").Default("").Comment("发言人名称")
		table.Text("content").Comment("评论内容")
		table.Integer("status").Default(1).Comment("1正常 2删除")
		table.Timestamps()
		table.Index("entry_id").Name("idx_comments_entry_id")
		table.Index("proc_id").Name("idx_comments_proc_id")
	}); err != nil {
		return err
	}

	// --- cc_records table ---
	if err := facades.Schema().Create("cc_records", func(table schema.Blueprint) {
		table.ID()
		table.BigInteger("entry_id")
		table.BigInteger("flow_id")
		table.BigInteger("process_id")
		table.BigInteger("proc_id")
		table.BigInteger("emp_id").Comment("抄送人ID")
		table.String("emp_name").Default("").Comment("抄送人名称")
		table.Integer("status").Default(0).Comment("0未读 1已读")
		table.Timestamps()
		table.Index("entry_id").Name("idx_cc_entry_id")
		table.Index("emp_id").Name("idx_cc_emp_id")
	}); err != nil {
		return err
	}

	// --- Performance indexes ---
	indexes := []struct {
		table   string
		index   string
		columns []string
	}{
		{"procs", "idx_procs_status_entry", []string{"status", "entry_id"}},
		{"procs", "idx_procs_emp_status", []string{"emp_id", "status"}},
		{"entrydatas", "idx_entrydatas_entry_field", []string{"entry_id", "field_name"}},
		{"entries", "idx_entry_status", []string{"status"}},
		{"entries", "idx_entry_flow_status", []string{"flow_id", "status"}},
	}
	for _, idx := range indexes {
		if !facades.Schema().HasIndex(idx.table, idx.index) {
			if err := facades.Schema().Table(idx.table, func(table schema.Blueprint) {
				table.Index(idx.columns...).Name(idx.index)
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *M20240624000000CreateWorkflowBaseTables) Down() error {
	tables := []string{
		"cc_records", "proc_comments", "proc_add_signs",
		"products", "attachments", "procs", "entrydatas", "entries",
		"templateforms", "templates", "processvars", "processes",
		"flowlinks", "flowtypes", "flows", "emps", "depts",
	}
	for _, t := range tables {
		facades.Schema().DropIfExists(t)
	}

	// Drop performance indexes
	indexesToDrop := []struct {
		table string
		index string
	}{
		{"procs", "idx_procs_status_entry"},
		{"procs", "idx_procs_emp_status"},
		{"entrydatas", "idx_entrydatas_entry_field"},
		{"entries", "idx_entry_status"},
		{"entries", "idx_entry_flow_status"},
	}
	for _, idx := range indexesToDrop {
		if facades.Schema().HasIndex(idx.table, idx.index) {
			facades.Schema().Table(idx.table, func(table schema.Blueprint) {
				table.DropIndexByName(idx.index)
			})
		}
	}

	return nil
}
