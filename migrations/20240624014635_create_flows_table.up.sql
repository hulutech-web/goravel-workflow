CREATE TABLE flows (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `flow_no` varchar(45) COLLATE utf8mb4_bin NOT NULL COMMENT '工作流编号',
  `flow_name` varchar(45) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '工作流名称',
  `template_id` bigint(20) NOT NULL DEFAULT '0',
  `flowchart` text COLLATE utf8mb4_bin,
  `jsplumb` text COLLATE utf8mb4_bin COMMENT 'jsplumb流程图数据',
  `type_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '流程设计文件',
  `is_publish` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否发布，发布后可用',
  `is_show` tinyint(4) NOT NULL DEFAULT '1' COMMENT '是否显示',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_flows_created_at (created_at),
  KEY idx_flows_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
