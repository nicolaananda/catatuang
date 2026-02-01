package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nicolaananda/catatuang/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByMSISDN(ctx context.Context, msisdn string) (*domain.User, error) {
	query := `
		SELECT id, msisdn, plan, free_tx_count, premium_until, is_blocked, created_at, updated_at
		FROM users
		WHERE msisdn = $1
	`

	user := &domain.User{}
	err := r.db.QueryRowContext(ctx, query, msisdn).Scan(
		&user.ID,
		&user.MSISDN,
		&user.Plan,
		&user.FreeTxCount,
		&user.PremiumUntil,
		&user.IsBlocked,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (msisdn, plan, free_tx_count)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, user.MSISDN, user.Plan, user.FreeTxCount).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET plan = $1, free_tx_count = $2, premium_until = $3, is_blocked = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query,
		user.Plan,
		user.FreeTxCount,
		user.PremiumUntil,
		user.IsBlocked,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepository) IncrementFreeTxCount(ctx context.Context, userID int64) error {
	query := `UPDATE users SET free_tx_count = free_tx_count + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to increment tx count: %w", err)
	}
	return nil
}

func (r *UserRepository) GetAll(ctx context.Context) ([]*domain.User, error) {
	query := `
		SELECT id, msisdn, plan, free_tx_count, premium_until, is_blocked, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID,
			&user.MSISDN,
			&user.Plan,
			&user.FreeTxCount,
			&user.PremiumUntil,
			&user.IsBlocked,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}
