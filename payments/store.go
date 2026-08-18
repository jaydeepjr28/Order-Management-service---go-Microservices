package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

type PaymentStore interface {
	HasEvent(ctx context.Context, eventID string) (bool, error)
	SaveEvent(ctx context.Context, eventID string) error
}

type pgStore struct {
	db *sql.DB
}

func NewPaymentStore(uri string) (PaymentStore, error) {
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
	CREATE TABLE IF NOT EXISTS processed_events (
		event_id VARCHAR(255) PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return nil, err
	}

	return &pgStore{db: db}, nil
}

func (s *pgStore) HasEvent(ctx context.Context, eventID string) (bool, error) {
	query := `SELECT event_id FROM processed_events WHERE event_id = $1`
	var id string
	err := s.db.QueryRowContext(ctx, query, eventID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *pgStore) SaveEvent(ctx context.Context, eventID string) error {
	query := `INSERT INTO processed_events (event_id) VALUES ($1)`
	_, err := s.db.ExecContext(ctx, query, eventID)
	return err
}
