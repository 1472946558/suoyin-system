CREATE DATABASE IF NOT EXISTS flash_delivery
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE flash_delivery;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_no VARCHAR(32) NOT NULL,
  wechat_openid VARCHAR(64) DEFAULT NULL,
  union_id VARCHAR(64) DEFAULT NULL,
  phone VARCHAR(20) DEFAULT NULL,
  nickname VARCHAR(64) DEFAULT NULL,
  avatar_url VARCHAR(255) DEFAULT NULL,
  city VARCHAR(64) DEFAULT NULL,
  default_from_address VARCHAR(255) DEFAULT NULL,
  default_to_address VARCHAR(255) DEFAULT NULL,
  status ENUM('active', 'disabled') NOT NULL DEFAULT 'active',
  last_login_at DATETIME DEFAULT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_user_no (user_no),
  UNIQUE KEY uk_users_wechat_openid (wechat_openid),
  KEY idx_users_phone (phone),
  KEY idx_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS orders (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_no VARCHAR(32) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  from_address VARCHAR(255) NOT NULL,
  to_address VARCHAR(255) NOT NULL,
  item_type VARCHAR(64) NOT NULL,
  weight_label VARCHAR(32) NOT NULL,
  note VARCHAR(500) DEFAULT NULL,
  distance_km DECIMAL(8,2) DEFAULT NULL,
  estimate_price DECIMAL(10,2) DEFAULT NULL,
  final_price DECIMAL(10,2) DEFAULT NULL,
  status ENUM(
    'created',
    'accepted',
    'picked',
    'delivering',
    'completed',
    'cancelled',
    'refunding'
  ) NOT NULL DEFAULT 'created',
  rider_name VARCHAR(64) DEFAULT NULL,
  rider_phone VARCHAR(20) DEFAULT NULL,
  payment_status ENUM('pending', 'paid', 'refunded') NOT NULL DEFAULT 'pending',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  accepted_at DATETIME DEFAULT NULL,
  completed_at DATETIME DEFAULT NULL,
  cancelled_at DATETIME DEFAULT NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_orders_order_no (order_no),
  KEY idx_orders_user_id (user_id),
  KEY idx_orders_status (status),
  KEY idx_orders_created_at (created_at),
  CONSTRAINT fk_orders_user_id
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS order_tracks (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_id BIGINT UNSIGNED NOT NULL,
  track_status ENUM(
    'created',
    'accepted',
    'picked',
    'delivering',
    'completed',
    'cancelled',
    'refunding'
  ) NOT NULL,
  title VARCHAR(64) NOT NULL,
  description VARCHAR(255) DEFAULT NULL,
  operator_type ENUM('system', 'rider', 'admin', 'user') NOT NULL DEFAULT 'system',
  operator_name VARCHAR(64) DEFAULT NULL,
  lat DECIMAL(10,7) DEFAULT NULL,
  lng DECIMAL(10,7) DEFAULT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_order_tracks_order_id (order_id),
  KEY idx_order_tracks_status (track_status),
  KEY idx_order_tracks_created_at (created_at),
  CONSTRAINT fk_order_tracks_order_id
    FOREIGN KEY (order_id) REFERENCES orders (id)
    ON UPDATE CASCADE
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS support_tickets (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  ticket_no VARCHAR(32) NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  order_id BIGINT UNSIGNED DEFAULT NULL,
  category ENUM('delivery', 'payment', 'refund', 'complaint', 'other') NOT NULL DEFAULT 'other',
  subject VARCHAR(120) NOT NULL,
  content TEXT NOT NULL,
  contact_name VARCHAR(64) DEFAULT NULL,
  contact_phone VARCHAR(20) DEFAULT NULL,
  status ENUM('open', 'processing', 'resolved', 'closed') NOT NULL DEFAULT 'open',
  resolved_at DATETIME DEFAULT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_support_tickets_ticket_no (ticket_no),
  KEY idx_support_tickets_user_id (user_id),
  KEY idx_support_tickets_order_id (order_id),
  KEY idx_support_tickets_status (status),
  CONSTRAINT fk_support_tickets_user_id
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON UPDATE CASCADE
    ON DELETE RESTRICT,
  CONSTRAINT fk_support_tickets_order_id
    FOREIGN KEY (order_id) REFERENCES orders (id)
    ON UPDATE CASCADE
    ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO order_tracks (
  order_id,
  track_status,
  title,
  description,
  operator_type,
  operator_name
)
SELECT
  o.id,
  'created',
  '订单已创建',
  '系统已接收订单，等待派单',
  'system',
  'system'
FROM orders o
LEFT JOIN order_tracks t
  ON t.order_id = o.id
 AND t.track_status = 'created'
WHERE t.id IS NULL;
