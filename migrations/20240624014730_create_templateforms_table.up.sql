CREATE TABLE templateforms (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,`template_id` int(11) NOT NULL DEFAULT '0',
  `field` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '表单字段英文名',
  `field_name` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '表单字段中文名',
  `field_type` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '' COMMENT '表单字段类型',
  `field_value` text COLLATE utf8mb4_bin COMMENT '表单字段值，select radio checkbox用',
  `field_default_value` text COLLATE utf8mb4_bin COMMENT '表单字段默认值',
  `field_rules` text COLLATE utf8mb4_bin,
  `sort` int(11) NOT NULL DEFAULT '100' COMMENT '排序',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_templateforms_created_at (created_at),
  KEY idx_templateforms_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
