CREATE TABLE flowtypes (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
`type_name` varchar(64) COLLATE utf8mb4_bin NOT NULL DEFAULT '',

  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_flowtypes_created_at (created_at),
  KEY idx_flowtypes_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
