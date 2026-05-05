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

const threatSelectColumns = `
    id,
    name,
    threat_category_id,
    source_type,
    description,
    q_threat,
    q_severity,
    bdu_id,
    applies_to_targets,
    applies_to_asset_types,
    impact_c,
    impact_i,
    impact_a,
    status,
    created_at,
    updated_at
`

func scanThreat(row pgx.Row, t *domain.Threat) error {
	return row.Scan(
		&t.ID,
		&t.Name,
		&t.ThreatCategoryID,
		&t.SourceType,
		&t.Description,
		&t.QThreat,
		&t.QSeverity,
		&t.BDUID,
		&t.AppliesToTargets,
		&t.AppliesToAssetTypes,
		&t.ImpactC,
		&t.ImpactI,
		&t.ImpactA,
		&t.Status,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
}

func (r *threatRepository) Create(ctx context.Context, t *domain.Threat) error {
	const q = `
INSERT INTO threats (
    name,
    threat_category_id,
    source_type,
    description,
    q_threat,
    q_severity,
    bdu_id
) VALUES (
    $1,$2,$3,$4,$5,$6,$7
) RETURNING id, created_at, updated_at
`
	row := r.pool.QueryRow(ctx, q,
		t.Name,
		t.ThreatCategoryID,
		t.SourceType,
		t.Description,
		t.QThreat,
		t.QSeverity,
		t.BDUID,
	)

	if err := row.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return fmt.Errorf("scan created threat: %w", err)
	}
	return nil
}

func (r *threatRepository) GetByID(ctx context.Context, id int64) (*domain.Threat, error) {
	q := `SELECT ` + threatSelectColumns + `FROM threats WHERE id = $1`

	var t domain.Threat
	if err := scanThreat(r.pool.QueryRow(ctx, q, id), &t); err != nil {
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

	q := `SELECT ` + threatSelectColumns + `FROM threats ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, q, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("list threats: %w", err)
	}
	defer rows.Close()

	var res []domain.Threat
	for rows.Next() {
		var t domain.Threat
		if err := scanThreat(rows, &t); err != nil {
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
    name                  = $1,
    threat_category_id    = $2,
    source_type           = $3,
    description           = $4,
    q_threat              = $5,
    q_severity            = $6,
    bdu_id                = $7,
    applies_to_targets    = $8,
    applies_to_asset_types = $9,
    impact_c              = $10,
    impact_i              = $11,
    impact_a              = $12,
    updated_at            = now()
WHERE id = $13
RETURNING updated_at
`
	row := r.pool.QueryRow(ctx, q,
		t.Name,
		t.ThreatCategoryID,
		t.SourceType,
		t.Description,
		t.QThreat,
		t.QSeverity,
		t.BDUID,
		t.AppliesToTargets,
		t.AppliesToAssetTypes,
		t.ImpactC,
		t.ImpactI,
		t.ImpactA,
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
