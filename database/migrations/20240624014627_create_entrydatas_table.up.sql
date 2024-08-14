CREATE TABLE entrydatas (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
`entry_id` bigint(20) NOT NULL DEFAULT '0',
`flow_id` bigint(20) NOT NULL DEFAULT '0',
`field_name` varchar(128) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
`field_value` text COLLATE utf8mb4_bin,
`field_remark` varchar(255) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY `entry_id` (`entry_id`),
  KEY `workflow_id` (`flow_id`),
  KEY idx_entrydatas_created_at (created_at),
  KEY idx_entrydatas_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
