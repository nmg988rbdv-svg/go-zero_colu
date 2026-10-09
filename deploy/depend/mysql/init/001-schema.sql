CREATE DATABASE IF NOT EXISTS trade CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE trade;

CREATE TABLE IF NOT EXISTS users (
  user_id BIGINT NOT NULL,
  legacy_id BIGINT NOT NULL DEFAULT 0,
  username VARCHAR(64) NOT NULL,
  password VARCHAR(255) NOT NULL,
  phone_number BIGINT NOT NULL DEFAULT 0,
  status INT NOT NULL DEFAULT 1,
  created_at BIGINT NOT NULL DEFAULT 0,
  updated_at BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id),
  UNIQUE KEY uk_users_username (username),
  KEY idx_users_legacy_id (legacy_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS order_final (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  order_id VARCHAR(64) NOT NULL,
  user_id BIGINT NOT NULL,
  pk_id BIGINT NOT NULL,
  symbol_id INT NOT NULL,
  symbol_name VARCHAR(32) NOT NULL,
  price VARCHAR(80) NOT NULL DEFAULT '',
  base_amount VARCHAR(80) NOT NULL DEFAULT '',
  quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  side INT NOT NULL DEFAULT 0,
  status INT NOT NULL DEFAULT 0,
  order_type INT NOT NULL DEFAULT 0,
  filled_base_amount VARCHAR(80) NOT NULL DEFAULT '',
  un_filled_base_amount VARCHAR(80) NOT NULL DEFAULT '',
  filled_quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  un_filled_quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  filled_avg_price VARCHAR(80) NOT NULL DEFAULT '',
  finish_reason VARCHAR(32) NOT NULL DEFAULT '',
  created_at BIGINT NOT NULL DEFAULT 0,
  updated_at BIGINT NOT NULL DEFAULT 0,
  finished_at BIGINT NOT NULL DEFAULT 0,
  archived_at BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_final_order_id (order_id),
  KEY idx_order_final_user_pk (user_id, pk_id DESC),
  KEY idx_order_final_symbol_name (symbol_name),
  KEY idx_order_final_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS match_trade (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  match_sub_id VARCHAR(64) NOT NULL,
  match_id VARCHAR(64) NOT NULL,
  symbol_id INT NOT NULL,
  symbol_name VARCHAR(32) NOT NULL,
  base_coin_id INT NOT NULL,
  quote_coin_id INT NOT NULL,
  taker_is_buy TINYINT(1) NOT NULL,
  price VARCHAR(80) NOT NULL DEFAULT '',
  base_amount VARCHAR(80) NOT NULL DEFAULT '',
  quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  begin_price VARCHAR(80) NOT NULL DEFAULT '',
  end_price VARCHAR(80) NOT NULL DEFAULT '',
  high_price VARCHAR(80) NOT NULL DEFAULT '',
  low_price VARCHAR(80) NOT NULL DEFAULT '',
  match_time BIGINT NOT NULL,
  taker_pk_id BIGINT NOT NULL,
  taker_user_id BIGINT NOT NULL,
  taker_order_id VARCHAR(64) NOT NULL,
  taker_filled_base_amount VARCHAR(80) NOT NULL DEFAULT '',
  taker_un_filled_base_amount VARCHAR(80) NOT NULL DEFAULT '',
  taker_filled_quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  taker_un_filled_quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  taker_order_status INT NOT NULL,
  maker_pk_id BIGINT NOT NULL,
  maker_user_id BIGINT NOT NULL,
  maker_order_id VARCHAR(64) NOT NULL,
  maker_filled_base_amount VARCHAR(80) NOT NULL DEFAULT '',
  maker_un_filled_base_amount VARCHAR(80) NOT NULL DEFAULT '',
  maker_filled_quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  maker_un_filled_quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  maker_order_status INT NOT NULL,
  created_at BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_match_trade_sub_id (match_sub_id),
  KEY idx_match_trade_match_id (match_id),
  KEY idx_match_trade_symbol_name (symbol_name),
  KEY idx_match_trade_match_time (match_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS kline_history (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  symbol VARCHAR(32) NOT NULL,
  symbol_id INT NOT NULL,
  kline_type INT NOT NULL,
  start_time BIGINT NOT NULL,
  end_time BIGINT NOT NULL,
  open_price VARCHAR(80) NOT NULL DEFAULT '',
  high_price VARCHAR(80) NOT NULL DEFAULT '',
  low_price VARCHAR(80) NOT NULL DEFAULT '',
  close_price VARCHAR(80) NOT NULL DEFAULT '',
  volume VARCHAR(80) NOT NULL DEFAULT '',
  amount VARCHAR(80) NOT NULL DEFAULT '',
  price_range VARCHAR(80) NOT NULL DEFAULT '',
  updated_at BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_kline_symbol_type_start (symbol, kline_type, start_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS tick (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  pk_id BIGINT NOT NULL,
  match_id VARCHAR(64) NOT NULL,
  match_sub_id VARCHAR(64) NOT NULL,
  order_id VARCHAR(64) NOT NULL,
  user_id BIGINT NOT NULL,
  symbol VARCHAR(32) NOT NULL,
  price VARCHAR(80) NOT NULL DEFAULT '',
  base_amount VARCHAR(80) NOT NULL DEFAULT '',
  quote_amount VARCHAR(80) NOT NULL DEFAULT '',
  side INT NOT NULL,
  role INT NOT NULL,
  created_at BIGINT NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_tick_match_role (match_sub_id, role),
  KEY idx_tick_pk_id (pk_id),
  KEY idx_tick_match_id (match_id),
  KEY idx_tick_order_id (order_id),
  KEY idx_tick_user_id (user_id),
  KEY idx_tick_symbol_created (symbol, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO users (user_id, legacy_id, username, password, phone_number, status, created_at, updated_at)
VALUES (1, 1, 'zhangsan', '$2a$10$RqN.PFhXEfqhbwu40TA1ce47TF39hIYwDlKsfUjBjqDPMZ0DyW55W', 13800138000, 1, 1748050000, 1748050000)
ON DUPLICATE KEY UPDATE username = 'zhangsan';
