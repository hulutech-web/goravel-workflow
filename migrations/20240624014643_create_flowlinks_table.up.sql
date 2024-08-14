CREATE TABLE flowlinks (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `flow_id` bigint(20) NOT NULL COMMENT '流程id',
    `type` varchar(45) COLLATE utf8mb4_bin NOT NULL COMMENT 'Condition:表示步骤流转\nRole:当前步骤操作人\nDept:部门\nEmp:员工\nSys:系统自动选人',
    `process_id` bigint(20) NOT NULL COMMENT '当前步骤id',
    `next_process_id` int NOT NULL DEFAULT '-1' COMMENT '下一个步骤 Condition -1未指定 0结束 -9上级查找\ntype=Role时为0，不启用',
    `auditor` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '0' COMMENT '审批人 系统自动 指定人员 指定部门 指定角色\ntype=Condition时不启用，-1002：部门经理 -1001:发起部门主管 0:不自动选人',
    `expression` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '条件判断表达式\n为1表示true，通过的话直接进入下一步骤$day <= 3 $day > 3的形式，根据表达式不同走不同的分支，每一个表达式都生成flowlink的新记录，其中type为Condition',
    `sort` int(11) NOT NULL COMMENT '条件判断顺序',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_flowlinks_created_at (created_at),
  KEY idx_flowlinks_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
