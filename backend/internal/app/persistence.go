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
