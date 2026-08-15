/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: persistence.go
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-05-11
 */

package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"

	_ "github.com/go-sql-driver/mysql"
)

type Persistence struct {
	db          *sql.DB
	redis       *redis.Client
	redisPrefix string
}

var persistenceStatements = []string{
	`CREATE TABLE IF NOT EXISTS app_configs (
		config_key VARCHAR(64) PRIMARY KEY,
		config_json JSON NOT NULL,
		updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	`CREATE TABLE IF NOT EXISTS cashier_orders (
		id VARCHAR(64) PRIMARY KEY,
		order_no VARCHAR(64) NOT NULL UNIQUE,
		org_id VARCHAR(64) NOT NULL,
			store_id VARCHAR(64) NOT NULL,
			store_name VARCHAR(128) NOT NULL,
			status VARCHAR(32) NOT NULL,
			customer_name VARCHAR(128) NOT NULL DEFAULT '',
			customer_phone VARCHAR(32) NOT NULL DEFAULT '',
			payment_method VARCHAR(32) NOT NULL DEFAULT '',
			total_amount DECIMAL(12,2) NOT NULL,
		paid_amount DECIMAL(12,2) NOT NULL,
		remark TEXT NOT NULL,
		items_json JSON NOT NULL,
		created_by VARCHAR(128) NOT NULL,
		created_at DATETIME(6) NOT NULL,
		void_reason TEXT NOT NULL,
		voided_by VARCHAR(128) NOT NULL DEFAULT '',
		voided_at DATETIME(6) NULL,
		updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	`CREATE TABLE IF NOT EXISTS recycle_orders (
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	`CREATE TABLE IF NOT EXISTS recycle_attachments (
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	`CREATE TABLE IF NOT EXISTS customer_appointments (
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
}

const saveCashierOrderStatement = `
		INSERT INTO cashier_orders (
			id, order_no, org_id, store_id, store_name, status,
			customer_name, customer_phone, payment_method,
			total_amount, paid_amount, remark, items_json, created_by, created_at,
			void_reason, voided_by, voided_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			order_no = VALUES(order_no),
			org_id = VALUES(org_id),
			store_id = VALUES(store_id),
			store_name = VALUES(store_name),
			status = VALUES(status),
			customer_name = VALUES(customer_name),
			customer_phone = VALUES(customer_phone),
			payment_method = VALUES(payment_method),
			total_amount = VALUES(total_amount),
			paid_amount = VALUES(paid_amount),
			remark = VALUES(remark),
			items_json = VALUES(items_json),
			created_by = VALUES(created_by),
			created_at = VALUES(created_at),
			void_reason = VALUES(void_reason),
			voided_by = VALUES(voided_by),
			voided_at = VALUES(voided_at)`

func newPersistence(cfg Config) (*Persistence, error) {
	if strings.ToLower(strings.TrimSpace(cfg.Mode)) != "persistent" {
		return nil, nil
	}
	if strings.TrimSpace(cfg.MySQLDSN) == "" {
		return nil, errors.New("persistent mode requires MYSQL_DSN")
	}
	if strings.TrimSpace(cfg.RedisAddr) == "" {
		return nil, errors.New("persistent mode requires REDIS_ADDR")
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(20)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Username: cfg.RedisUser,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		_ = db.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	p := &Persistence{
		db:          db,
		redis:       rdb,
		redisPrefix: cfg.RedisPrefix,
	}
	if err := p.ensureSchema(ctx); err != nil {
		_ = p.Close()
		return nil, err
	}
	return p, nil
}

func (p *Persistence) Close() error {
	if p == nil {
		return nil
	}

	var closeErr error
	if p.redis != nil {
		closeErr = errors.Join(closeErr, p.redis.Close())
	}
	if p.db != nil {
		closeErr = errors.Join(closeErr, p.db.Close())
	}
	return closeErr
}

func (p *Persistence) ensureSchema(ctx context.Context) error {
	for _, statement := range persistenceStatements {
		if _, err := p.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("ensure schema: %w", err)
		}
	}
	if err := p.ensureColumn(ctx, "cashier_orders", "customer_name", "VARCHAR(128) NOT NULL DEFAULT ''"); err != nil {
		return fmt.Errorf("ensure cashier customer_name column: %w", err)
	}
	if err := p.ensureColumn(ctx, "cashier_orders", "customer_phone", "VARCHAR(32) NOT NULL DEFAULT ''"); err != nil {
		return fmt.Errorf("ensure cashier customer_phone column: %w", err)
	}
	if err := p.ensureColumn(ctx, "cashier_orders", "payment_method", "VARCHAR(32) NOT NULL DEFAULT ''"); err != nil {
		return fmt.Errorf("ensure cashier payment_method column: %w", err)
	}
	if err := p.ensureColumn(ctx, "cashier_orders", "void_reason", "TEXT NOT NULL"); err != nil {
		return fmt.Errorf("ensure cashier void_reason column: %w", err)
	}
	if err := p.ensureColumn(ctx, "cashier_orders", "voided_by", "VARCHAR(128) NOT NULL DEFAULT ''"); err != nil {
		return fmt.Errorf("ensure cashier voided_by column: %w", err)
	}
	if err := p.ensureColumn(ctx, "cashier_orders", "voided_at", "DATETIME(6) NULL"); err != nil {
		return fmt.Errorf("ensure cashier voided_at column: %w", err)
	}
	return nil
}

func (p *Persistence) ensureColumn(ctx context.Context, tableName string, columnName string, columnDefinition string) error {
	var count int
	if err := p.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = ?
		  AND COLUMN_NAME = ?`,
		tableName,
		columnName,
	).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := p.db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDefinition))
	return err
}

func (p *Persistence) sessionKey(token string) string {
	prefix := strings.TrimSpace(p.redisPrefix)
	if prefix == "" {
		prefix = "gold:"
	}
	return prefix + "session:" + token
}

func (p *Persistence) loadConfig(ctx context.Context, key string, target interface{}) (bool, error) {
	var raw []byte
	err := p.db.QueryRowContext(ctx, `SELECT config_json FROM app_configs WHERE config_key = ?`, key).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return false, err
	}
	return true, nil
}

func (p *Persistence) saveConfig(ctx context.Context, key string, value interface{}) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = p.db.ExecContext(ctx, `
		INSERT INTO app_configs (config_key, config_json)
		VALUES (?, ?)
		ON DUPLICATE KEY UPDATE config_json = VALUES(config_json)`,
		key,
		raw,
	)
	return err
}

func (p *Persistence) loadRecycleAttachments(ctx context.Context) (map[string][]AttachmentAsset, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, order_id, org_id, store_id, category, storage_provider, object_key, public_url,
		       thumbnail_url, file_name, content_type, size_bytes, source, status, uploaded_by, uploaded_at
		FROM recycle_attachments
		ORDER BY uploaded_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(map[string][]AttachmentAsset)
	for rows.Next() {
		var item AttachmentAsset
		if err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.OrgID,
			&item.StoreID,
			&item.Category,
			&item.StorageProvider,
			&item.ObjectKey,
			&item.PublicURL,
			&item.ThumbnailURL,
			&item.FileName,
			&item.ContentType,
			&item.SizeBytes,
			&item.Source,
			&item.Status,
			&item.UploadedBy,
			&item.UploadedAt,
		); err != nil {
			return nil, err
		}
		items[item.OrderID] = append(items[item.OrderID], item)
	}
	return items, rows.Err()
}

func (p *Persistence) saveRecycleAttachments(ctx context.Context, orderID string, items []AttachmentAsset) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM recycle_attachments WHERE order_id = ?`, orderID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO recycle_attachments (
				id, order_id, org_id, store_id, category, storage_provider, object_key, public_url,
				thumbnail_url, file_name, content_type, size_bytes, source, status, uploaded_by, uploaded_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID,
			item.OrderID,
			item.OrgID,
			item.StoreID,
			item.Category,
			item.StorageProvider,
			item.ObjectKey,
			item.PublicURL,
			item.ThumbnailURL,
			item.FileName,
			item.ContentType,
			item.SizeBytes,
			item.Source,
			item.Status,
			item.UploadedBy,
			item.UploadedAt,
		); err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

func (p *Persistence) saveSession(ctx context.Context, session Session) error {
	payload, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	return p.redis.Set(ctx, p.sessionKey(session.Token), payload, ttl).Err()
}

func (p *Persistence) loadSession(ctx context.Context, token string) (Session, bool, error) {
	raw, err := p.redis.Get(ctx, p.sessionKey(token)).Bytes()
	if errors.Is(err, redis.Nil) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}

	var session Session
	if err := json.Unmarshal(raw, &session); err != nil {
		return Session{}, false, err
	}
	return session, true, nil
}

func (p *Persistence) deleteSession(ctx context.Context, token string) error {
	return p.redis.Del(ctx, p.sessionKey(token)).Err()
}

// --- 顾客会话 ---

func (p *Persistence) customerSessionKey(token string) string {
	prefix := strings.TrimSpace(p.redisPrefix)
	if prefix == "" {
		prefix = "gold:"
	}
	return prefix + "customer_session:" + token
}

func (p *Persistence) saveCustomerSession(ctx context.Context, session CustomerSession) error {
	payload, err := json.Marshal(session)
	if err != nil {
		return err
	}
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = 12 * time.Hour
	}
	return p.redis.Set(ctx, p.customerSessionKey(session.Token), payload, ttl).Err()
}

func (p *Persistence) loadCustomerSession(ctx context.Context, token string) (CustomerSession, bool, error) {
	raw, err := p.redis.Get(ctx, p.customerSessionKey(token)).Bytes()
	if errors.Is(err, redis.Nil) {
		return CustomerSession{}, false, nil
	}
	if err != nil {
		return CustomerSession{}, false, err
	}
	var session CustomerSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return CustomerSession{}, false, err
	}
	return session, true, nil
}

func (p *Persistence) deleteCustomerSession(ctx context.Context, token string) error {
	return p.redis.Del(ctx, p.customerSessionKey(token)).Err()
}

// --- 顾客预约 ---

func (p *Persistence) loadCustomerAppointments(ctx context.Context) ([]CustomerAppointment, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, org_id, customer_id, customer_name, customer_phone,
		       store_id, store_name, store_address, store_phone,
		       appointment_date, appointment_time, status, remark,
		       created_at, updated_at, cancelled_at, cancel_reason,
		       confirmed_at, confirmed_by, completed_at
		FROM customer_appointments
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CustomerAppointment, 0)
	for rows.Next() {
		var appt CustomerAppointment
		var cancelledAt, confirmedAt, completedAt sql.NullTime
		if err := rows.Scan(
			&appt.ID, &appt.OrgID, &appt.CustomerID, &appt.CustomerName, &appt.CustomerPhone,
			&appt.StoreID, &appt.StoreName, &appt.StoreAddress, &appt.StorePhone,
			&appt.AppointmentDate, &appt.AppointmentTime, &appt.Status, &appt.Remark,
			&appt.CreatedAt, &appt.UpdatedAt, &cancelledAt, &appt.CancelReason,
			&confirmedAt, &appt.ConfirmedBy, &completedAt,
		); err != nil {
			return nil, err
		}
		if cancelledAt.Valid {
			appt.CancelledAt = &cancelledAt.Time
		}
		if confirmedAt.Valid {
			appt.ConfirmedAt = &confirmedAt.Time
		}
		if completedAt.Valid {
			appt.CompletedAt = &completedAt.Time
		}
		items = append(items, appt)
	}
	return items, rows.Err()
}

func (p *Persistence) saveCustomerAppointment(ctx context.Context, appt CustomerAppointment) error {
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO customer_appointments (
			id, org_id, customer_id, customer_name, customer_phone,
			store_id, store_name, store_address, store_phone,
			appointment_date, appointment_time, status, remark,
			created_at, updated_at, cancelled_at, cancel_reason,
			confirmed_at, confirmed_by, completed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			customer_name = VALUES(customer_name),
			customer_phone = VALUES(customer_phone),
			store_name = VALUES(store_name),
			store_address = VALUES(store_address),
			store_phone = VALUES(store_phone),
			status = VALUES(status),
			remark = VALUES(remark),
			updated_at = VALUES(updated_at),
			cancelled_at = VALUES(cancelled_at),
			cancel_reason = VALUES(cancel_reason),
			confirmed_at = VALUES(confirmed_at),
			confirmed_by = VALUES(confirmed_by),
			completed_at = VALUES(completed_at)`,
		appt.ID, appt.OrgID, appt.CustomerID, appt.CustomerName, appt.CustomerPhone,
		appt.StoreID, appt.StoreName, appt.StoreAddress, appt.StorePhone,
		appt.AppointmentDate, appt.AppointmentTime, appt.Status, appt.Remark,
		appt.CreatedAt, appt.UpdatedAt, appt.CancelledAt, appt.CancelReason,
		appt.ConfirmedAt, appt.ConfirmedBy, appt.CompletedAt,
	)
	return err
}

func (p *Persistence) loadCashierOrders(ctx context.Context) ([]CashierOrder, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, order_no, org_id, store_id, store_name, status,
		       customer_name, customer_phone, payment_method,
		       total_amount, paid_amount, remark, items_json, created_by, created_at,
		       void_reason, voided_by, voided_at
		FROM cashier_orders
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]CashierOrder, 0)
	for rows.Next() {
		var order CashierOrder
		var itemsJSON []byte
		var voidedAt sql.NullTime
		if err := rows.Scan(
			&order.ID,
			&order.OrderNo,
			&order.OrgID,
			&order.StoreID,
			&order.StoreName,
			&order.Status,
			&order.CustomerName,
			&order.CustomerPhone,
			&order.PaymentMethod,
			&order.TotalAmount,
			&order.PaidAmount,
			&order.Remark,
			&itemsJSON,
			&order.CreatedBy,
			&order.CreatedAt,
			&order.VoidReason,
			&order.VoidedBy,
			&voidedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, err
		}
		if voidedAt.Valid {
			order.VoidedAt = &voidedAt.Time
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (p *Persistence) saveCashierOrder(ctx context.Context, order CashierOrder) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, saveCashierOrderStatement,
		order.ID,
		order.OrderNo,
		order.OrgID,
		order.StoreID,
		order.StoreName,
		order.Status,
		order.CustomerName,
		order.CustomerPhone,
		order.PaymentMethod,
		order.TotalAmount,
		order.PaidAmount,
		order.Remark,
		itemsJSON,
		order.CreatedBy,
		order.CreatedAt,
		order.VoidReason,
		order.VoidedBy,
		order.VoidedAt,
	)
	return err
}

func (p *Persistence) loadRecycleOrders(ctx context.Context) (map[string]RecycleOrder, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, order_no, org_id, store_id, store_name, status, customer_name, customer_phone,
		       estimated_amount, confirmed_amount, items_json, attachment_urls_json,
		       remark, created_by, created_at, confirmed_at
		FROM recycle_orders
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make(map[string]RecycleOrder)
	for rows.Next() {
		var order RecycleOrder
		var itemsJSON []byte
		var attachmentsJSON []byte
		var confirmedAt sql.NullTime
		if err := rows.Scan(
			&order.ID,
			&order.OrderNo,
			&order.OrgID,
			&order.StoreID,
			&order.StoreName,
			&order.Status,
			&order.CustomerName,
			&order.CustomerPhone,
			&order.EstimatedAmount,
			&order.ConfirmedAmount,
			&itemsJSON,
			&attachmentsJSON,
			&order.Remark,
			&order.CreatedBy,
			&order.CreatedAt,
			&confirmedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(attachmentsJSON, &order.AttachmentURLs); err != nil {
			return nil, err
		}
		if confirmedAt.Valid {
			order.ConfirmedAt = &confirmedAt.Time
		}
		orders[order.ID] = order
	}
	return orders, rows.Err()
}

func (p *Persistence) saveRecycleOrder(ctx context.Context, order RecycleOrder) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return err
	}
	attachmentsJSON, err := json.Marshal(order.AttachmentURLs)
	if err != nil {
		return err
	}

	_, err = p.db.ExecContext(ctx, `
		INSERT INTO recycle_orders (
			id, order_no, org_id, store_id, store_name, status, customer_name, customer_phone,
			estimated_amount, confirmed_amount, items_json, attachment_urls_json, remark,
			created_by, created_at, confirmed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			order_no = VALUES(order_no),
			org_id = VALUES(org_id),
			store_id = VALUES(store_id),
			store_name = VALUES(store_name),
			status = VALUES(status),
			customer_name = VALUES(customer_name),
			customer_phone = VALUES(customer_phone),
			estimated_amount = VALUES(estimated_amount),
			confirmed_amount = VALUES(confirmed_amount),
			items_json = VALUES(items_json),
			attachment_urls_json = VALUES(attachment_urls_json),
			remark = VALUES(remark),
			created_by = VALUES(created_by),
			created_at = VALUES(created_at),
			confirmed_at = VALUES(confirmed_at)`,
		order.ID,
		order.OrderNo,
		order.OrgID,
		order.StoreID,
		order.StoreName,
		order.Status,
		order.CustomerName,
		order.CustomerPhone,
		order.EstimatedAmount,
		order.ConfirmedAmount,
		itemsJSON,
		attachmentsJSON,
		order.Remark,
		order.CreatedBy,
		order.CreatedAt,
		order.ConfirmedAt,
	)
	return err
}
