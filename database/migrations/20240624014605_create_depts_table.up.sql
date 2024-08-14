CREATE TABLE depts (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
   `dept_name` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `pid` bigint(20) NOT NULL DEFAULT '0',
  `html` varchar(255)  NULL DEFAULT '-',
  `level` int(11) NOT NULL DEFAULT '0' COMMENT '部门层级',
  `director_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '部门主管 0表示不存在',
  `manager_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '部门经理 0表示不存在',
  `rank` int(11) NOT NULL DEFAULT '1',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY `dept_name` (`dept_name`),
  KEY idx_depts_created_at (created_at),
  KEY idx_depts_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
