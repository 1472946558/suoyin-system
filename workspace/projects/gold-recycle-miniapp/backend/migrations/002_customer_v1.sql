-- 002_customer_v1.sql - 顾客端 V1 数据库迁移
-- 节点：N03
-- 日期：2026-08-15
-- 说明：新增 customer_appointments 独立表；customer_profiles 使用 app_configs JSON blob 存储
-- 回滚：DROP TABLE IF EXISTS customer_appointments;

CREATE TABLE IF NOT EXISTS customer_appointments (
    id VARCHAR(64) PRIMARY KEY,
    org_id VARCHAR(64) NOT NULL,
    customer_id VARCHAR(64) NOT NULL,
    customer_name VARCHAR(128) NOT NULL DEFAULT '',
    customer_phone VARCHAR(32) NOT NULL DEFAULT '',
    store_id VARCHAR(64) NOT NULL,
    store_name VARCHAR(128) NOT NULL DEFAULT '',
    store_address VARCHAR(512) NOT NULL DEFAULT '',
    store_phone VARCHAR(32) NOT NULL DEFAULT '',
    appointment_date DATE NOT NULL,
    appointment_time VARCHAR(8) NOT NULL,
    service_type VARCHAR(32) NOT NULL DEFAULT 'OLD_FOR_NEW',
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    remark TEXT NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    cancelled_at DATETIME(6) NULL,
    cancel_reason TEXT NOT NULL DEFAULT '',
    confirmed_at DATETIME(6) NULL,
    confirmed_by VARCHAR(128) NOT NULL DEFAULT '',
    completed_at DATETIME(6) NULL,
    INDEX idx_customer_appointments_customer (customer_id),
    INDEX idx_customer_appointments_store (store_id),
    INDEX idx_customer_appointments_date (appointment_date, appointment_time),
    INDEX idx_customer_appointments_status (status),
    INDEX idx_customer_appointments_phone_store_time (customer_phone, store_id, appointment_date, appointment_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
