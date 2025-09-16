package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MAPiryazev/OzonTest/internal/config"
	"github.com/MAPiryazev/OzonTest/internal/customerrors"
	"github.com/MAPiryazev/OzonTest/internal/models"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(cfg *config.DBConfig) (*PostgresStorage, error) {
	dbCredentials := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sql.Open("postgres", dbCredentials)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBCreation, err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBMaxConnLifeTime) * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ошибка при проверке соединиения с БД: %v", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (p *PostgresStorage) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}

func (p *PostgresStorage) CreateUser(user *models.User) error {
	ctx := context.Background()
	query := `INSERT INTO users (id, username) VALUES ($1, $2)`
	_, err := p.db.ExecContext(ctx, query, user.ID, user.Username)
	if err != nil {
		return fmt.Errorf("create user: %v", err)
	}
	return nil
}

func (p *PostgresStorage) GetUserByID(id string) (*models.User, error) {
	query := `SELECT id, username FROM users WHERE id = $1`
	row := p.db.QueryRow(query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: user with id %s", customerrors.ErrNotFound, id)
		}
		return nil, fmt.Errorf("%w: %v", customerrors.ErrDBQuery, err)
	}

	return &user, nil
}
