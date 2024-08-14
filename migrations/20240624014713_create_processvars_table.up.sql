CREATE TABLE processvars (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
`process_id` bigint(20) NOT NULL,
`flow_id` bigint(20) NOT NULL COMMENT '流程id',
`expression_field` varchar(45) COLLATE utf8mb4_bin NOT NULL COMMENT '条件表达式字段名称',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_procesvars_created_at (created_at),
  KEY idx_procesvars_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
