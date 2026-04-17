package repository

import (
	"context"
	"fmt"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ThreatFilter struct {
	Limit  int32
	Offset int32
}

type ThreatRepository interface {
	Create(ctx context.Context, t *domain.Threat) error
	GetByID(ctx context.Context, id int64) (*domain.Threat, error)
	List(ctx context.Context, f ThreatFilter) ([]domain.Threat, error)
	Update(ctx context.Context, t *domain.Threat) error
	Delete(ctx context.Context, id int64) error
}

type threatRepository struct {
	pool *pgxpool.Pool
}

func NewThreatRepository(pool *pgxpool.Pool) ThreatRepository {
	return &threatRepository{pool: pool}
}

func (r *threatRepository) Create(ctx context.Context, t *domain.Threat) error {
	const q = `
INSERT INTO threats (
    name,
    threat_category_id,
    source_type,
    description,
    base_likelihood
) VALUES (
    $1,$2,$3,$4,$5
) RETURNING
    id,
    created_at,
    updated_at
`
	row := r.pool.QueryRow(ctx, q,
		t.Name,
		t.ThreatCategoryID,
		t.SourceType,
		t.Description,
		t.BaseLikelihood,
	)

	if err := row.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return fmt.Errorf("scan created threat: %w", err)
	}

	return nil
}

func (r *threatRepository) GetByID(ctx context.Context, id int64) (*domain.Threat, error) {
	const q = `
SELECT
    id,
    name,
    threat_category_id,
    source_type,
    description,
    base_likelihood,
    q_threat,
    q_severity,
    created_at,
    updated_at
FROM threats
WHERE id = $1
`
	var t domain.Threat
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&t.ID,
		&t.Name,
		&t.ThreatCategoryID,
		&t.SourceType,
		&t.Description,
		&t.BaseLikelihood,
		&t.QThreat,
		&t.QSeverity,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get threat by id: %w", err)
	}

	return &t, nil
}

func (r *threatRepository) List(ctx context.Context, f ThreatFilter) ([]domain.Threat, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	const q = `
SELECT
    id,
    name,
    threat_category_id,
    source_type,
    description,
    base_likelihood,
    q_threat,
    q_severity,
    created_at,
    updated_at
FROM threats
ORDER BY id
LIMIT $1 OFFSET $2
`
	rows, err := r.pool.Query(ctx, q, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("list threats: %w", err)
	}
	defer rows.Close()

	var res []domain.Threat
	for rows.Next() {
		var t domain.Threat
		if err := rows.Scan(
			&t.ID,
			&t.Name,
			&t.ThreatCategoryID,
			&t.SourceType,
			&t.Description,
			&t.BaseLikelihood,
			&t.QThreat,
			&t.QSeverity,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan threat: %w", err)
		}
		res = append(res, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return res, nil
}

func (r *threatRepository) Update(ctx context.Context, t *domain.Threat) error {
	const q = `
UPDATE threats
SET
    name = $1,
    threat_category_id = $2,
    source_type = $3,
    description = $4,
    base_likelihood = $5,
    updated_at = now()
WHERE id = $6
RETURNING updated_at
`
	row := r.pool.QueryRow(ctx, q,
		t.Name,
		t.ThreatCategoryID,
		t.SourceType,
		t.Description,
		t.BaseLikelihood,
		t.ID,
	)

	if err := row.Scan(&t.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("threat not found")
		}
		return fmt.Errorf("update threat: %w", err)
	}

	return nil
}

func (r *threatRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM threats WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete threat: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("threat not found")
	}
	return nil
}
