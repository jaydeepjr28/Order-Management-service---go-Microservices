package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	pb "github.com/jaydeepjr28/commons/api"
	_ "github.com/lib/pq"
)

// store encapsulates the PostgreSQL client for the Orders service.
type store struct {
	db *sql.DB
}

func NewStore(uri string) (*store, error) {
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		customer_id VARCHAR(255) NOT NULL,
		status VARCHAR(255) NOT NULL,
		payment_link VARCHAR(255),
		items JSONB
	);`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return nil, err
	}

	return &store{db: db}, nil
}

// Create inserts a new order document into Postgres and returns its generated ID.
func (s *store) Create(ctx context.Context, o Order) (string, error) {
	itemsJSON, err := json.Marshal(o.Items)
	if err != nil {
		return "", err
	}

	query := `INSERT INTO orders (customer_id, status, payment_link, items) VALUES ($1, $2, $3, $4) RETURNING id`
	var id string
	err = s.db.QueryRowContext(ctx, query, o.CustomerID, o.Status, o.PaymentLink, itemsJSON).Scan(&id)
	return id, err
}

// Get fetches a specific order by its ID, ensuring it belongs to the given customerID.
func (s *store) Get(ctx context.Context, id, customerID string) (*Order, error) {
	var o Order
	var itemsJSON []byte

	query := `SELECT id, customer_id, status, payment_link, items FROM orders WHERE id = $1 AND customer_id = $2`
	err := s.db.QueryRowContext(ctx, query, id, customerID).Scan(&o.ID, &o.CustomerID, &o.Status, &o.PaymentLink, &itemsJSON)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
		return nil, err
	}

	return &o, nil
}

// Update modifies specific fields (PaymentLink and Status) of an existing order.
func (s *store) Update(ctx context.Context, id string, newOrder *pb.Order) error {
	query := `UPDATE orders SET payment_link = $1, status = $2 WHERE id = $3`
	_, err := s.db.ExecContext(ctx, query, newOrder.PaymentLink, newOrder.Status, id)
	return err
}

// GetByCustomer fetches all orders belonging to a specific customer.
func (s *store) GetByCustomer(ctx context.Context, customerID string) ([]*Order, error) {
	query := `SELECT id, customer_id, status, payment_link, items FROM orders WHERE customer_id = $1 ORDER BY id DESC`
	rows, err := s.db.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var o Order
		var itemsJSON []byte
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.Status, &o.PaymentLink, &itemsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(itemsJSON, &o.Items); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
