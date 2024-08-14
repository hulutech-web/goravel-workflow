CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
  avatarUrl varchar(255) NULL DEFAULT '' COMMENT '头像地址',
  password varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  email varchar(255) NULL DEFAULT '' COMMENT '邮箱',
  gender int NULL DEFAULT 1 COMMENT '性别',
  mobile varchar(255) NULL unique DEFAULT '' COMMENT '手机号',
  id_number varchar(255) NULL unique DEFAULT '' COMMENT '身份证号',
  is_member int NULL DEFAULT 0 COMMENT '是否会员',
  state int NULL DEFAULT 1 COMMENT '状态: 1正常, 2禁用',
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
