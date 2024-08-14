CREATE TABLE templates (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
 `template_name` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '',

  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_templates_created_at (created_at),
  KEY idx_templates_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
