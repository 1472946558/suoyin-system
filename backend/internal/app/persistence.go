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
	`CREATE TABLE IF NOT EXISTS payment_transactions (
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
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	`CREATE TABLE IF NOT EXISTS cashier_orders (
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
}

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
	return nil
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

func (p *Persistence) loadPaymentTransactions(ctx context.Context) ([]PaymentTransaction, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, payment_no, order_no, biz_type, org_id, store_id, store_name, amount,
		       method, status, callback_status, provider_ref, provider_payload, operator_name,
		       customer_label, remark, anomaly, paid_at, created_at, updated_at
		FROM payment_transactions
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PaymentTransaction, 0)
	for rows.Next() {
		var item PaymentTransaction
		var payload sql.NullString
		var paidAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.PaymentNo,
			&item.OrderNo,
			&item.BizType,
			&item.OrgID,
			&item.StoreID,
			&item.StoreName,
			&item.Amount,
			&item.Method,
			&item.Status,
			&item.CallbackStatus,
			&item.ProviderRef,
			&payload,
			&item.OperatorName,
			&item.CustomerLabel,
			&item.Remark,
			&item.Anomaly,
			&paidAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if payload.Valid {
			item.ProviderPayload = payload.String
		}
		if paidAt.Valid {
			item.PaidAt = &paidAt.Time
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (p *Persistence) savePaymentTransaction(ctx context.Context, item PaymentTransaction) error {
	var payload interface{}
	if strings.TrimSpace(item.ProviderPayload) != "" {
		payloadValue := strings.TrimSpace(item.ProviderPayload)
		if json.Valid([]byte(payloadValue)) {
			payload = payloadValue
		} else {
			encoded, err := json.Marshal(map[string]string{"raw": payloadValue})
			if err != nil {
				return err
			}
			payload = string(encoded)
		}
	}
	_, err := p.db.ExecContext(ctx, `
		INSERT INTO payment_transactions (
			id, payment_no, order_no, biz_type, org_id, store_id, store_name, amount,
			method, status, callback_status, provider_ref, provider_payload, operator_name,
			customer_label, remark, anomaly, paid_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			payment_no = VALUES(payment_no),
			order_no = VALUES(order_no),
			biz_type = VALUES(biz_type),
			org_id = VALUES(org_id),
			store_id = VALUES(store_id),
			store_name = VALUES(store_name),
			amount = VALUES(amount),
			method = VALUES(method),
			status = VALUES(status),
			callback_status = VALUES(callback_status),
			provider_ref = VALUES(provider_ref),
			provider_payload = VALUES(provider_payload),
			operator_name = VALUES(operator_name),
			customer_label = VALUES(customer_label),
			remark = VALUES(remark),
			anomaly = VALUES(anomaly),
			paid_at = VALUES(paid_at),
			created_at = VALUES(created_at),
			updated_at = VALUES(updated_at)`,
		item.ID,
		item.PaymentNo,
		item.OrderNo,
		item.BizType,
		item.OrgID,
		item.StoreID,
		item.StoreName,
		item.Amount,
		item.Method,
		item.Status,
		item.CallbackStatus,
		item.ProviderRef,
		payload,
		item.OperatorName,
		item.CustomerLabel,
		item.Remark,
		item.Anomaly,
		item.PaidAt,
		item.CreatedAt,
		item.UpdatedAt,
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

func (p *Persistence) loadCashierOrders(ctx context.Context) ([]CashierOrder, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, order_no, org_id, store_id, store_name, status, payment_method,
		       total_amount, paid_amount, remark, items_json, created_by, created_at
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
		if err := rows.Scan(
			&order.ID,
			&order.OrderNo,
			&order.OrgID,
			&order.StoreID,
			&order.StoreName,
			&order.Status,
			&order.PaymentMethod,
			&order.TotalAmount,
			&order.PaidAmount,
			&order.Remark,
			&itemsJSON,
			&order.CreatedBy,
			&order.CreatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, err
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

	_, err = p.db.ExecContext(ctx, `
		INSERT INTO cashier_orders (
			id, order_no, org_id, store_id, store_name, status, payment_method,
			total_amount, paid_amount, remark, items_json, created_by, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			order_no = VALUES(order_no),
			org_id = VALUES(org_id),
			store_id = VALUES(store_id),
			store_name = VALUES(store_name),
			status = VALUES(status),
			payment_method = VALUES(payment_method),
			total_amount = VALUES(total_amount),
			paid_amount = VALUES(paid_amount),
			remark = VALUES(remark),
			items_json = VALUES(items_json),
			created_by = VALUES(created_by),
			created_at = VALUES(created_at)`,
		order.ID,
		order.OrderNo,
		order.OrgID,
		order.StoreID,
		order.StoreName,
		order.Status,
		order.PaymentMethod,
		order.TotalAmount,
		order.PaidAmount,
		order.Remark,
		itemsJSON,
		order.CreatedBy,
		order.CreatedAt,
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
