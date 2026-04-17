package repository

import (
	"context"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ThreatSourceRepository interface {
	List(ctx context.Context) ([]domain.ThreatSource, error)
	ForThreat(ctx context.Context, threatID int64) ([]domain.ThreatSource, error)
}

type threatSourceRepository struct {
	pool *pgxpool.Pool
}

func NewThreatSourceRepository(pool *pgxpool.Pool) ThreatSourceRepository {
	return &threatSourceRepository{pool: pool}
}

func (r *threatSourceRepository) List(ctx context.Context) ([]domain.ThreatSource, error) {
	const q = `
		SELECT id, code, name, COALESCE(description, ''), created_at
		FROM threat_sources
		ORDER BY id`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ThreatSource, 0)
	for rows.Next() {
		var s domain.ThreatSource
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.Description, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *threatSourceRepository) ForThreat(ctx context.Context, threatID int64) ([]domain.ThreatSource, error) {
	const q = `
		SELECT s.id, s.code, s.name, COALESCE(s.description, ''), s.created_at
		FROM threat_sources s
		JOIN source_threats st ON st.threat_source_id = s.id
		WHERE st.threat_id = $1
		ORDER BY s.id`
	rows, err := r.pool.Query(ctx, q, threatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ThreatSource, 0)
	for rows.Next() {
		var s domain.ThreatSource
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.Description, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
