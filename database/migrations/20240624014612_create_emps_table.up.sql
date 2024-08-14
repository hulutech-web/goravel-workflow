CREATE TABLE emps (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
 `email` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
 `password` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL,
 `workno` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工号',
 `dept_id` bigint(20) NOT NULL DEFAULT '0' COMMENT '部门id',
 `leave` smallint(6) NOT NULL DEFAULT '0' COMMENT '离职状态',
  user_id bigint(20)  NULL DEFAULT '0' COMMENT '用户id',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_emps_created_at (created_at),
  KEY idx_emps_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
