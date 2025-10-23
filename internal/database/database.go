package database

import (
	"database/sql"
	"fmt"
	"time"

	"alimpay-go/internal/model"
	"alimpay-go/pkg/logger"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// DB 数据库实例
type DB struct {
	*sql.DB
}

// Config 数据库配置
type Config struct {
	Type            string
	Path            string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

var globalDB *DB

// Init 初始化数据库
func Init(cfg *Config) (*DB, error) {
	// 打开数据库连接，添加参数以防止死锁
	// _busy_timeout: 设置忙等待超时（毫秒）
	// _journal_mode=WAL: 使用WAL模式提高并发性能
	// _synchronous=NORMAL: 平衡性能与数据安全
	// _cache_size=-64000: 设置缓存大小（64MB）
	dsn := cfg.Path + "?_busy_timeout=10000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-64000"
	db, err := sql.Open(cfg.Type, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数（SQLite建议单连接写入）
	// MaxOpenConns设置为1可以避免写入冲突
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 1
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 1
	}
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	globalDB = &DB{db}

	// 优化SQLite设置
	if err := globalDB.optimizeSQLite(); err != nil {
		logger.Warn("Failed to optimize SQLite settings", zap.Error(err))
	}

	// 初始化表结构
	if err := globalDB.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	logger.Info("Database initialized successfully",
		zap.String("path", cfg.Path),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns))
	return globalDB, nil
}

// optimizeSQLite 优化SQLite设置
func (db *DB) optimizeSQLite() error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",            // 使用WAL模式
		"PRAGMA synchronous=NORMAL",          // 平衡性能与安全
		"PRAGMA cache_size=-64000",           // 64MB缓存
		"PRAGMA temp_store=MEMORY",           // 临时表存储在内存
		"PRAGMA mmap_size=268435456",         // 256MB内存映射
		"PRAGMA page_size=4096",              // 页面大小
		"PRAGMA auto_vacuum=INCREMENTAL",     // 增量自动清理
		"PRAGMA busy_timeout=10000",          // 10秒忙等待超时
		"PRAGMA foreign_keys=ON",             // 启用外键约束
		"PRAGMA journal_size_limit=67108864", // 64MB日志大小限制
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			logger.Warn("Failed to execute pragma", zap.String("pragma", pragma), zap.Error(err))
		}
	}

	logger.Info("SQLite optimizations applied")
	return nil
}

// GetDB 获取全局数据库实例
func GetDB() *DB {
	return globalDB
}

// initTables 初始化数据库表
func (db *DB) initTables() error {
	// 创建订单表
	createOrderTableSQL := `
	CREATE TABLE IF NOT EXISTS codepay_orders (
		id VARCHAR(32) PRIMARY KEY,
		out_trade_no VARCHAR(64) NOT NULL,
		type VARCHAR(10) NOT NULL,
		pid VARCHAR(20) NOT NULL,
		name VARCHAR(255) NOT NULL,
		price DECIMAL(10, 2) NOT NULL,
		payment_amount DECIMAL(10, 2) DEFAULT 0,
		status TINYINT(1) DEFAULT 0,
		add_time DATETIME NOT NULL,
		pay_time DATETIME,
		notify_url VARCHAR(255),
		return_url VARCHAR(255),
		sitename VARCHAR(255)
	);`

	if _, err := db.Exec(createOrderTableSQL); err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_out_trade_no ON codepay_orders(out_trade_no);",
		"CREATE INDEX IF NOT EXISTS idx_status ON codepay_orders(status);",
		"CREATE INDEX IF NOT EXISTS idx_payment_amount ON codepay_orders(payment_amount);",
		"CREATE INDEX IF NOT EXISTS idx_add_time ON codepay_orders(add_time);",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	logger.Info("Database tables initialized successfully")
	return nil
}

// CreateOrder 创建订单
func (db *DB) CreateOrder(order *model.Order) error {
	query := `
		INSERT INTO codepay_orders (
			id, out_trade_no, type, pid, name, price, payment_amount,
			status, add_time, notify_url, return_url, sitename
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query,
		order.ID, order.OutTradeNo, order.Type, order.PID, order.Name,
		order.Price, order.PaymentAmount, order.Status, order.AddTime,
		order.NotifyURL, order.ReturnURL, order.Sitename,
	)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	logger.Info("Order created", zap.String("order_id", order.ID), zap.String("out_trade_no", order.OutTradeNo))
	return nil
}

// GetOrderByOutTradeNo 根据商户订单号获取订单
func (db *DB) GetOrderByOutTradeNo(outTradeNo, pid string) (*model.Order, error) {
	query := `
		SELECT id, out_trade_no, type, pid, name, price, payment_amount,
		       status, add_time, pay_time, notify_url, return_url, sitename
		FROM codepay_orders
		WHERE out_trade_no = ? AND pid = ?
	`

	var order model.Order
	var payTime sql.NullTime

	err := db.QueryRow(query, outTradeNo, pid).Scan(
		&order.ID, &order.OutTradeNo, &order.Type, &order.PID, &order.Name,
		&order.Price, &order.PaymentAmount, &order.Status, &order.AddTime,
		&payTime, &order.NotifyURL, &order.ReturnURL, &order.Sitename,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if payTime.Valid {
		order.PayTime = &payTime.Time
	}

	return &order, nil
}

// GetOrderByID 根据订单ID获取订单
func (db *DB) GetOrderByID(id string) (*model.Order, error) {
	query := `
		SELECT id, out_trade_no, type, pid, name, price, payment_amount,
		       status, add_time, pay_time, notify_url, return_url, sitename
		FROM codepay_orders
		WHERE id = ?
	`

	var order model.Order
	var payTime sql.NullTime

	err := db.QueryRow(query, id).Scan(
		&order.ID, &order.OutTradeNo, &order.Type, &order.PID, &order.Name,
		&order.Price, &order.PaymentAmount, &order.Status, &order.AddTime,
		&payTime, &order.NotifyURL, &order.ReturnURL, &order.Sitename,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if payTime.Valid {
		order.PayTime = &payTime.Time
	}

	return &order, nil
}

// GetPendingOrderByAmount 根据金额获取待支付订单（经营码模式）
func (db *DB) GetPendingOrderByAmount(amount float64) (*model.Order, error) {
	query := `
		SELECT id, out_trade_no, type, pid, name, price, payment_amount,
		       status, add_time, pay_time, notify_url, return_url, sitename
		FROM codepay_orders
		WHERE payment_amount = ? AND status = ?
		ORDER BY add_time ASC
		LIMIT 1
	`

	var order model.Order
	var payTime sql.NullTime

	err := db.QueryRow(query, amount, model.OrderStatusPending).Scan(
		&order.ID, &order.OutTradeNo, &order.Type, &order.PID, &order.Name,
		&order.Price, &order.PaymentAmount, &order.Status, &order.AddTime,
		&payTime, &order.NotifyURL, &order.ReturnURL, &order.Sitename,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pending order: %w", err)
	}

	if payTime.Valid {
		order.PayTime = &payTime.Time
	}

	return &order, nil
}

// CheckAmountExists 检查金额是否已存在（用于金额分配）
func (db *DB) CheckAmountExists(amount float64, sinceTime time.Time) (bool, error) {
	query := `
		SELECT COUNT(*) FROM codepay_orders
		WHERE payment_amount = ? AND status = ? AND add_time >= ?
	`

	var count int
	err := db.QueryRow(query, amount, model.OrderStatusPending, sinceTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check amount exists: %w", err)
	}

	return count > 0, nil
}

// UpdateOrderStatus 更新订单状态
func (db *DB) UpdateOrderStatus(id string, status int, payTime time.Time) error {
	query := `
		UPDATE codepay_orders
		SET status = ?, pay_time = ?
		WHERE id = ?
	`

	result, err := db.Exec(query, status, payTime, id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %s", id)
	}

	logger.Info("Order status updated", zap.String("order_id", id), zap.Int("status", status))
	return nil
}

// GetOrders 获取订单列表
func (db *DB) GetOrders(pid string, limit int) ([]*model.Order, error) {
	query := `
		SELECT id, out_trade_no, type, pid, name, price, payment_amount,
		       status, add_time, pay_time, notify_url, return_url, sitename
		FROM codepay_orders
		WHERE pid = ?
		ORDER BY add_time DESC
		LIMIT ?
	`

	rows, err := db.Query(query, pid, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		var payTime sql.NullTime

		err := rows.Scan(
			&order.ID, &order.OutTradeNo, &order.Type, &order.PID, &order.Name,
			&order.Price, &order.PaymentAmount, &order.Status, &order.AddTime,
			&payTime, &order.NotifyURL, &order.ReturnURL, &order.Sitename,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if payTime.Valid {
			order.PayTime = &payTime.Time
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

/*
GetOrdersByStatus 根据状态获取订单列表
@description 查询指定状态的所有订单
@param status 订单状态
@return []*model.Order 订单列表
@return error 查询错误
*/
func (db *DB) GetOrdersByStatus(status int) ([]*model.Order, error) {
	query := `
		SELECT id, out_trade_no, type, pid, name, price, payment_amount,
		       status, add_time, pay_time, notify_url, return_url, sitename
		FROM codepay_orders
		WHERE status = ?
		ORDER BY add_time DESC
	`

	rows, err := db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		var payTime sql.NullTime

		err := rows.Scan(
			&order.ID, &order.OutTradeNo, &order.Type, &order.PID, &order.Name,
			&order.Price, &order.PaymentAmount, &order.Status, &order.AddTime,
			&payTime, &order.NotifyURL, &order.ReturnURL, &order.Sitename,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if payTime.Valid {
			order.PayTime = &payTime.Time
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}

/*
GetTodayOrdersByStatus 获取今日指定状态的订单
@description 查询今天创建的指定状态订单
@param status 订单状态
@return []*model.Order 订单列表
@return error 查询错误
*/
func (db *DB) GetTodayOrdersByStatus(status int) ([]*model.Order, error) {
	query := `
		SELECT id, out_trade_no, type, pid, name, price, payment_amount,
		       status, add_time, pay_time, notify_url, return_url, sitename
		FROM codepay_orders
		WHERE status = ? AND DATE(add_time) = DATE('now', 'localtime')
		ORDER BY add_time DESC
	`

	rows, err := db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's orders by status: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		var payTime sql.NullTime

		err := rows.Scan(
			&order.ID, &order.OutTradeNo, &order.Type, &order.PID, &order.Name,
			&order.Price, &order.PaymentAmount, &order.Status, &order.AddTime,
			&payTime, &order.NotifyURL, &order.ReturnURL, &order.Sitename,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if payTime.Valid {
			order.PayTime = &payTime.Time
		}

		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}

// DeleteExpiredOrders 删除过期订单
func (db *DB) DeleteExpiredOrders(expiredTime time.Time) (int64, error) {
	query := `
		DELETE FROM codepay_orders
		WHERE status = ? AND add_time < ?
	`

	result, err := db.Exec(query, model.OrderStatusPending, expiredTime)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired orders: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logger.Info("Expired orders deleted", zap.Int64("count", rowsAffected))
	}

	return rowsAffected, nil
}

// CountOrders 统计订单数量
func (db *DB) CountOrders(status *int) (int, error) {
	var query string
	var args []interface{}

	if status != nil {
		query = "SELECT COUNT(*) FROM codepay_orders WHERE status = ?"
		args = append(args, *status)
	} else {
		query = "SELECT COUNT(*) FROM codepay_orders"
	}

	var count int
	err := db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count orders: %w", err)
	}

	return count, nil
}

// GetRecentOrders 获取最近的订单
func (db *DB) GetRecentOrders(limit int) ([]*model.Order, error) {
	query := `
		SELECT id, out_trade_no, type, pid, name, price, payment_amount,
		       status, add_time, pay_time, notify_url, return_url, sitename
		FROM codepay_orders
		ORDER BY add_time DESC
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent orders: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		var payTime sql.NullTime

		err := rows.Scan(
			&order.ID, &order.OutTradeNo, &order.Type, &order.PID, &order.Name,
			&order.Price, &order.PaymentAmount, &order.Status, &order.AddTime,
			&payTime, &order.NotifyURL, &order.ReturnURL, &order.Sitename,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if payTime.Valid {
			order.PayTime = &payTime.Time
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

// GetPendingOrdersSince 获取指定时间之后的待支付订单
func (db *DB) GetPendingOrdersSince(since time.Time) ([]*model.Order, error) {
	query := `
		SELECT id, out_trade_no, type, pid, name, price, payment_amount,
		       status, add_time, pay_time, notify_url, return_url, sitename
		FROM codepay_orders
		WHERE status = ? AND add_time >= ?
		ORDER BY add_time DESC
	`

	rows, err := db.Query(query, model.OrderStatusPending, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending orders: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var order model.Order
		var payTime sql.NullTime

		err := rows.Scan(
			&order.ID, &order.OutTradeNo, &order.Type, &order.PID, &order.Name,
			&order.Price, &order.PaymentAmount, &order.Status, &order.AddTime,
			&payTime, &order.NotifyURL, &order.ReturnURL, &order.Sitename,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		if payTime.Valid {
			order.PayTime = &payTime.Time
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}
	return nil
}
