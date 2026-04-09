package repository

import (
	"context"
	"errors"
	"fmt"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE username = $1`

	var u domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Role,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	query := `
		SELECT id, username, password_hash, role, is_active, created_at, updated_at
		FROM users WHERE id = $1`

	var u domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.Role,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}
