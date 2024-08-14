CREATE TABLE procs (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `entry_id` bigint(20) NOT NULL,
  `flow_id` bigint(20) NOT NULL COMMENT '流程id',
  `process_id` bigint(20) NOT NULL COMMENT '当前步骤',
  `process_name` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '当前步骤名称',
  `emp_id` bigint(20) NOT NULL COMMENT '审核人',
  `emp_name` varchar(32) COLLATE utf8mb4_bin DEFAULT NULL COMMENT '审核人名称',
  `dept_name` varchar(32) COLLATE utf8mb4_bin DEFAULT NULL COMMENT '审核人部门名称',
  `auditor_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '具体操作人',
  `auditor_name` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '操作人名称',
  `auditor_dept` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '操作人部门',
  `status` int(11) NOT NULL COMMENT '当前处理状态 0待处理 9通过 -1驳回\n0：处理中\n-1：驳回\n9：会签',
  `content` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL COMMENT '批复内容',
  `is_read` int(11) NOT NULL DEFAULT '0' COMMENT '是否查看',
  `is_real` tinyint(4) NOT NULL DEFAULT '1' COMMENT '审核人和操作人是否同一人',
  `circle` smallint(6) NOT NULL DEFAULT '1',
  `beizhu` text COLLATE utf8mb4_bin COMMENT '备注',
  `concurrence` datetime(3)  NULL  COMMENT '',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  KEY `entry_id` (`entry_id`),
  KEY `workflow_id` (`flow_id`),
  KEY `emp_id` (`emp_id`),
  KEY `step_id` (`process_id`),
  PRIMARY KEY (id),
  KEY idx_procs_created_at (created_at),
  KEY idx_procs_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
