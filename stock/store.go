package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	pb "github.com/jaydeepjr28/commons/api"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// PostgresStore implements StockStore backed by PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(uri string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS stock_items (
		id VARCHAR(64) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		quantity INT NOT NULL CHECK (quantity >= 0),
		price_id VARCHAR(128) NOT NULL DEFAULT '',
		price BIGINT NOT NULL DEFAULT 9900,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW()
	);
	ALTER TABLE stock_items ADD COLUMN IF NOT EXISTS price BIGINT NOT NULL DEFAULT 9900;`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, fmt.Errorf("failed to initialize stock_items table: %w", err)
	}

	// Seed default items if empty
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM stock_items").Scan(&count); err == nil && count == 0 {
		seedQuery := `
		INSERT INTO stock_items (id, name, quantity, price_id, price) VALUES
			('1', 'Cheese Burger', 20, 'price_1PA7xYRoHcnwnsLn4CrfWAvG', 19900),
			('2', 'Potato Chips', 10, 'price_1PBaC2RoHcnwnsLnYLWpx1CK', 9900),
			('3', 'Parmesan Truffle Fries', 15, '', 14900)
		ON CONFLICT (id) DO NOTHING;`
		if _, err := db.ExecContext(ctx, seedQuery); err != nil {
			zap.L().Warn("Failed to seed initial stock items", zap.Error(err))
		} else {
			zap.L().Info("Seeded default stock items into PostgreSQL")
		}
	}

	// Ensure existing seeded items have proper pricing
	_, _ = db.ExecContext(ctx, `
		UPDATE stock_items SET price = 19900 WHERE id = '1' AND (price = 0 OR price = 9900);
		UPDATE stock_items SET price = 9900 WHERE id = '2' AND price = 0;
		UPDATE stock_items SET price = 14900 WHERE id = '3' AND (price = 0 OR price = 9900);
	`)

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) GetItem(ctx context.Context, id string) (*pb.Item, error) {
	query := `SELECT id, name, quantity, price_id, price FROM stock_items WHERE id = $1`
	var item pb.Item
	err := s.db.QueryRowContext(ctx, query, id).Scan(&item.ID, &item.Name, &item.Quantity, &item.PriceID, &item.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item not found")
		}
		return nil, err
	}
	return &item, nil
}

func (s *PostgresStore) GetItems(ctx context.Context, ids []string) ([]*pb.Item, error) {
	var rows *sql.Rows
	var err error

	if len(ids) == 0 {
		query := `SELECT id, name, quantity, price_id, price FROM stock_items ORDER BY id ASC`
		rows, err = s.db.QueryContext(ctx, query)
	} else {
		query := `SELECT id, name, quantity, price_id, price FROM stock_items WHERE id = ANY($1) ORDER BY id ASC`
		rows, err = s.db.QueryContext(ctx, query, pq.Array(ids))
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*pb.Item
	for rows.Next() {
		var item pb.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity, &item.PriceID, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *PostgresStore) CreateItem(ctx context.Context, item *pb.Item) (*pb.Item, error) {
	query := `
	INSERT INTO stock_items (id, name, quantity, price_id, price, updated_at)
	VALUES ($1, $2, $3, $4, CASE WHEN $5 > 0 THEN $5 ELSE 9900 END, NOW())
	ON CONFLICT (id) DO UPDATE SET
		name = EXCLUDED.name,
		quantity = EXCLUDED.quantity,
		price_id = CASE WHEN EXCLUDED.price_id != '' THEN EXCLUDED.price_id ELSE stock_items.price_id END,
		price = CASE WHEN EXCLUDED.price > 0 THEN EXCLUDED.price ELSE stock_items.price END,
		updated_at = NOW()
	RETURNING id, name, quantity, price_id, price`

	var created pb.Item
	err := s.db.QueryRowContext(ctx, query, item.ID, item.Name, item.Quantity, item.PriceID, item.Price).
		Scan(&created.ID, &created.Name, &created.Quantity, &created.PriceID, &created.Price)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (s *PostgresStore) UpdateItem(ctx context.Context, item *pb.Item) (*pb.Item, error) {
	query := `
	UPDATE stock_items
	SET name = CASE WHEN $2 != '' THEN $2 ELSE name END,
	    quantity = $3,
	    price_id = CASE WHEN $4 != '' THEN $4 ELSE price_id END,
	    price = CASE WHEN $5 > 0 THEN $5 ELSE price END,
	    updated_at = NOW()
	WHERE id = $1
	RETURNING id, name, quantity, price_id, price`

	var updated pb.Item
	err := s.db.QueryRowContext(ctx, query, item.ID, item.Name, item.Quantity, item.PriceID, item.Price).
		Scan(&updated.ID, &updated.Name, &updated.Quantity, &updated.PriceID, &updated.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item with id %s not found", item.ID)
		}
		return nil, err
	}
	return &updated, nil
}

func (s *PostgresStore) DecrementItems(ctx context.Context, items []*pb.Item) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		query := `UPDATE stock_items SET quantity = quantity - $1, updated_at = NOW() WHERE id = $2 AND quantity >= $1`
		res, err := tx.ExecContext(ctx, query, item.Quantity, item.ID)
		if err != nil {
			return err
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return fmt.Errorf("insufficient stock or item not found: %s", item.ID)
		}
	}

	return tx.Commit()
}

func (s *PostgresStore) IncrementItems(ctx context.Context, items []*pb.Item) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, item := range items {
		query := `UPDATE stock_items SET quantity = quantity + $1, updated_at = NOW() WHERE id = $2`
		_, err := tx.ExecContext(ctx, query, item.Quantity, item.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}


// MemoryStore implements StockStore in memory as an active or fallback store.
type MemoryStore struct {
	mu    sync.Mutex
	stock map[string]*pb.Item
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		stock: map[string]*pb.Item{
			"1": {
				ID:       "1",
				Name:     "Cheese Burger",
				PriceID:  "price_1PA7xYRoHcnwnsLn4CrfWAvG",
				Quantity: 20,
			},
			"2": {
				ID:       "2",
				Name:     "Potato Chips",
				PriceID:  "price_1PBaC2RoHcnwnsLnYLWpx1CK",
				Quantity: 10,
			},
		},
	}
}

func (s *MemoryStore) GetItem(ctx context.Context, id string) (*pb.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.stock[id]
	if !ok {
		return nil, fmt.Errorf("item not found")
	}
	return item, nil
}

func (s *MemoryStore) GetItems(ctx context.Context, ids []string) ([]*pb.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res []*pb.Item
	if len(ids) == 0 {
		for _, i := range s.stock {
			res = append(res, i)
		}
		return res, nil
	}

	for _, id := range ids {
		if i, ok := s.stock[id]; ok {
			res = append(res, i)
		}
	}

	return res, nil
}

func (s *MemoryStore) CreateItem(ctx context.Context, item *pb.Item) (*pb.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stock[item.ID] = item
	return item, nil
}

func (s *MemoryStore) UpdateItem(ctx context.Context, item *pb.Item) (*pb.Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.stock[item.ID]
	if !ok {
		return nil, fmt.Errorf("item with id %s not found", item.ID)
	}
	if item.Name != "" {
		existing.Name = item.Name
	}
	existing.Quantity = item.Quantity
	if item.PriceID != "" {
		existing.PriceID = item.PriceID
	}
	if item.Price > 0 {
		existing.Price = item.Price
	}
	return existing, nil
}

func (s *MemoryStore) DecrementItems(ctx context.Context, items []*pb.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		stockItem, ok := s.stock[item.ID]
		if !ok {
			return fmt.Errorf("item not found: %s", item.ID)
		}
		if stockItem.Quantity < item.Quantity {
			return fmt.Errorf("insufficient stock for item %s", item.ID)
		}
	}

	for _, item := range items {
		stockItem := s.stock[item.ID]
		stockItem.Quantity -= item.Quantity
	}

	return nil
}

func (s *MemoryStore) IncrementItems(ctx context.Context, items []*pb.Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range items {
		stockItem, ok := s.stock[item.ID]
		if ok {
			stockItem.Quantity += item.Quantity
		}
	}

	return nil
}


// NewStore initializes a PostgresStore if uri is non-empty, falling back gracefully to MemoryStore.
func NewStore(uri string) StockStore {
	if uri != "" {
		pgStore, err := NewPostgresStore(uri)
		if err == nil {
			zap.L().Info("Connected to PostgreSQL stock store", zap.String("uri", uri))
			return pgStore
		}
		zap.L().Warn("Failed to connect to PostgreSQL stock store, falling back to in-memory store", zap.Error(err))
	}
	return NewMemoryStore()
}
