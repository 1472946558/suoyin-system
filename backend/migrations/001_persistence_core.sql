CREATE TABLE IF NOT EXISTS app_configs (
    config_key VARCHAR(64) PRIMARY KEY,
    config_json JSON NOT NULL,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS payment_transactions (
    id VARCHAR(64) PRIMARY KEY,
    payment_no VARCHAR(64) NOT NULL UNIQUE,
    order_no VARCHAR(64) NOT NULL,
    biz_type VARCHAR(32) NOT NULL,
    org_id VARCHAR(64) NOT NULL,
    store_id VARCHAR(64) NOT NULL,
    store_name VARCHAR(128) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    method VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    callback_status VARCHAR(32) NOT NULL,
    provider_ref VARCHAR(128) NOT NULL,
    provider_payload JSON NULL,
    operator_name VARCHAR(128) NOT NULL,
    customer_label VARCHAR(128) NOT NULL,
    remark TEXT NOT NULL,
    anomaly TINYINT(1) NOT NULL DEFAULT 0,
    paid_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    INDEX idx_payment_transactions_order_no (order_no),
    INDEX idx_payment_transactions_store_paid_at (store_id, paid_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cashier_orders (
    id VARCHAR(64) PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    org_id VARCHAR(64) NOT NULL,
    store_id VARCHAR(64) NOT NULL,
    store_name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    payment_method VARCHAR(32) NOT NULL,
    total_amount DECIMAL(12,2) NOT NULL,
    paid_amount DECIMAL(12,2) NOT NULL,
    remark TEXT NOT NULL,
    items_json JSON NOT NULL,
    created_by VARCHAR(128) NOT NULL,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS recycle_orders (
    id VARCHAR(64) PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    org_id VARCHAR(64) NOT NULL,
    store_id VARCHAR(64) NOT NULL,
    store_name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL,
    customer_name VARCHAR(128) NOT NULL,
    customer_phone VARCHAR(32) NOT NULL,
    estimated_amount DECIMAL(12,2) NOT NULL,
    confirmed_amount DECIMAL(12,2) NOT NULL,
    items_json JSON NOT NULL,
    attachment_urls_json JSON NOT NULL,
    remark TEXT NOT NULL,
    created_by VARCHAR(128) NOT NULL,
    created_at DATETIME(6) NOT NULL,
    confirmed_at DATETIME(6) NULL,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS recycle_attachments (
    id VARCHAR(64) PRIMARY KEY,
    order_id VARCHAR(64) NOT NULL,
    org_id VARCHAR(64) NOT NULL,
    store_id VARCHAR(64) NOT NULL,
    category VARCHAR(32) NOT NULL,
    storage_provider VARCHAR(32) NOT NULL,
    object_key VARCHAR(255) NOT NULL,
    public_url TEXT NOT NULL,
    thumbnail_url TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(128) NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    source VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    uploaded_by VARCHAR(128) NOT NULL,
    uploaded_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    INDEX idx_recycle_attachments_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
