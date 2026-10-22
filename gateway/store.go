package main

import (
	"context"
	"crypto/rand"
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
		name VARCHAR(255) NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'customer',
		is_verified BOOLEAN DEFAULT TRUE,
		email_verified_at TIMESTAMPTZ DEFAULT NOW(),
		provider VARCHAR(50) DEFAULT 'local',
		provider_id VARCHAR(255),
		avatar_url VARCHAR(500)
	);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'customer';
	ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT TRUE;
	ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ DEFAULT NOW();
	ALTER TABLE users ADD COLUMN IF NOT EXISTS provider VARCHAR(50) DEFAULT 'local';
	ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_id VARCHAR(255);
	ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
	`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return nil, err
	}

	return &store{db: db}, nil
}

type UserModel struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Password   string `json:"-"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	Provider   string `json:"provider,omitempty"`
	ProviderID string `json:"provider_id,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
}

func (s *store) CreateUser(ctx context.Context, u *UserModel) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)

	query := `INSERT INTO users (email, password, name, role) VALUES ($1, $2, $3, $4) RETURNING id`
	var id string
	err = s.db.QueryRowContext(ctx, query, u.Email, u.Password, u.Name, u.Role).Scan(&id)
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (s *store) GetUserByEmail(ctx context.Context, email string) (*UserModel, error) {
	var user UserModel
	var provider, providerID, avatarURL sql.NullString
	query := `SELECT id, email, password, name, role, provider, provider_id, avatar_url FROM users WHERE email = $1`
	err := s.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.Role, &provider, &providerID, &avatarURL)
	if err != nil {
		return nil, err
	}
	if provider.Valid {
		user.Provider = provider.String
	}
	if providerID.Valid {
		user.ProviderID = providerID.String
	}
	if avatarURL.Valid {
		user.AvatarURL = avatarURL.String
	}
	return &user, nil
}

func (s *store) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := s.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}

func (s *store) CreateVerifiedUser(ctx context.Context, u *UserModel) error {
	query := `INSERT INTO users (email, password, name, role, is_verified, email_verified_at) VALUES ($1, $2, $3, $4, TRUE, NOW()) RETURNING id`
	var id string
	err := s.db.QueryRowContext(ctx, query, u.Email, u.Password, u.Name, u.Role).Scan(&id)
	if err != nil {
		return err
	}
	u.ID = id
	return nil
}

func (s *store) UpsertOAuthUser(ctx context.Context, u *UserModel) (isNew bool, err error) {
	existing, _ := s.GetUserByEmail(ctx, u.Email)
	if existing != nil {
		updateQuery := `
			UPDATE users 
			SET provider = CASE WHEN provider IS NULL OR provider = 'local' THEN $1 ELSE provider END,
			    provider_id = COALESCE(provider_id, $2),
			    avatar_url = COALESCE(avatar_url, $3),
			    name = CASE WHEN name = '' THEN $4 ELSE name END,
			    is_verified = TRUE
			WHERE email = $5
			RETURNING id, name, role
		`
		err = s.db.QueryRowContext(ctx, updateQuery, u.Provider, u.ProviderID, u.AvatarURL, u.Name, u.Email).Scan(&u.ID, &u.Name, &u.Role)
		return false, err
	}

	randBytes := make([]byte, 24)
	_, _ = rand.Read(randBytes)
	hashed, _ := bcrypt.GenerateFromPassword(randBytes, bcrypt.DefaultCost)
	u.Password = string(hashed)

	insertQuery := `
		INSERT INTO users (email, password, name, role, is_verified, email_verified_at, provider, provider_id, avatar_url) 
		VALUES ($1, $2, $3, $4, TRUE, NOW(), $5, $6, $7) 
		RETURNING id
	`
	err = s.db.QueryRowContext(ctx, insertQuery, u.Email, u.Password, u.Name, u.Role, u.Provider, u.ProviderID, u.AvatarURL).Scan(&u.ID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *store) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
