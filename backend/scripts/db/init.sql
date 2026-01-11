SET NAMES utf8mb4;
CREATE DATABASE IF NOT EXISTS slot_factory;
USE slot_factory;

-- 玩家錢包表
CREATE TABLE IF NOT EXISTS wallets (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    player_id VARCHAR(255) NOT NULL UNIQUE COMMENT '玩家唯一標識',
    balance DECIMAL(18, 4) NOT NULL DEFAULT 0.0000 COMMENT '錢包餘額',
    currency VARCHAR(10) NOT NULL DEFAULT 'TWD' COMMENT '幣種',
    version BIGINT NOT NULL DEFAULT 0 COMMENT '樂觀鎖版本號 (用於併發控制)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_player_id (player_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='玩家錢包表';

-- 錢包流水表 (Audit Log)
-- 雖然 K8s 上為了省資源可能不寫入，但 Schema 還是要先定義好
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    player_id VARCHAR(255) NOT NULL,
    amount DECIMAL(18, 4) NOT NULL COMMENT '變動金額',
    transaction_type VARCHAR(20) NOT NULL COMMENT '交易類型: SPIN, DEPOSIT, WITHDRAW',
    balance_after DECIMAL(18, 4) NOT NULL COMMENT '變動後餘額',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_player_id_created (player_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='錢包流水表';
