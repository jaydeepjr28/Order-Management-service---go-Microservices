package main

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

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
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		name VARCHAR(255) NOT NULL
	);`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return nil, err
	}

	return &store{db: db}, nil
}

type UserModel struct {
	ID		string	`json:"id"`
	Email		string	`json:"email"`
	Password	string	`json:"-"`
	Name		string	`json:"name"`
}

func (s *store) CreateUser(ctx context.Context, u *UserModel) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

	query := `INSERT INTO users (email, password, name) VALUES ($1, $2, $3) RETURNING id`
	var id string
	err = s.db.QueryRowContext(ctx, query, u.Email, u.Password, u.Name).Scan(&id)
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (s *store) GetUserByEmail(ctx context.Context, email string) (*UserModel, error) {
	var user UserModel
	query := `SELECT id, email, password, name FROM users WHERE email = $1`
	err := s.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Name)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *store) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
