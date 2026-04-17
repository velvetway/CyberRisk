package repository

import (
	"context"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DestructiveActionRepository interface {
	List(ctx context.Context) ([]domain.DestructiveAction, error)
	ForThreat(ctx context.Context, threatID int64) ([]domain.DestructiveAction, error)
}

type destructiveActionRepository struct {
	pool *pgxpool.Pool
}

func NewDestructiveActionRepository(pool *pgxpool.Pool) DestructiveActionRepository {
	return &destructiveActionRepository{pool: pool}
}

func (r *destructiveActionRepository) List(ctx context.Context) ([]domain.DestructiveAction, error) {
	const q = `
		SELECT id, code, name,
		       affects_confidentiality, affects_integrity, affects_availability,
		       COALESCE(description, ''), created_at
		FROM destructive_actions
		ORDER BY id`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.DestructiveAction, 0)
	for rows.Next() {
		var d domain.DestructiveAction
		if err := rows.Scan(&d.ID, &d.Code, &d.Name,
			&d.AffectsConfidentiality, &d.AffectsIntegrity, &d.AffectsAvailability,
			&d.Description, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *destructiveActionRepository) ForThreat(ctx context.Context, threatID int64) ([]domain.DestructiveAction, error) {
	const q = `
		SELECT da.id, da.code, da.name,
		       da.affects_confidentiality, da.affects_integrity, da.affects_availability,
		       COALESCE(da.description, ''), da.created_at
		FROM destructive_actions da
		JOIN threat_destructive_actions tda ON tda.destructive_action_id = da.id
		WHERE tda.threat_id = $1
		ORDER BY da.id`
	rows, err := r.pool.Query(ctx, q, threatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.DestructiveAction, 0)
	for rows.Next() {
		var d domain.DestructiveAction
		if err := rows.Scan(&d.ID, &d.Code, &d.Name,
			&d.AffectsConfidentiality, &d.AffectsIntegrity, &d.AffectsAvailability,
			&d.Description, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
