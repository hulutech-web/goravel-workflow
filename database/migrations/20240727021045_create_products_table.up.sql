CREATE TABLE products (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  special varchar(255) NOT NULL,
  dimension varchar(255) NOT NULL,
  quantity int(10) unsigned NOT NULL,
  unit varchar(255) NOT NULL,
  unit_price float NOT NULL,
  discount_price float NOT NULL,
  amount float NOT NULL,
  description text  NULL,
  image_url varchar(255)  NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_products_created_at (created_at),
  KEY idx_products_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
