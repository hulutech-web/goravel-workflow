CREATE TABLE attachments (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `path` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件路径',
    `name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '文件名',
    `ext` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '扩展名',
    `type` char(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '上传方式local,oss',
    `user_id` bigint(20) unsigned DEFAULT NULL,
    `size` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '文件大小',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_attachments_created_at (created_at),
  KEY idx_attachments_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
