package repository

import (
	"context"
	"fmt"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AssetFilter struct {
	Limit  int32
	Offset int32
}

type AssetRepository interface {
	Create(ctx context.Context, a *domain.Asset) error
	GetByID(ctx context.Context, id int64) (*domain.Asset, error)
	List(ctx context.Context, f AssetFilter) ([]domain.Asset, error)
	Update(ctx context.Context, a *domain.Asset) error
	Delete(ctx context.Context, id int64) error
}

type assetRepository struct {
	pool *pgxpool.Pool
}

func NewAssetRepository(pool *pgxpool.Pool) AssetRepository {
	return &assetRepository{pool: pool}
}

const assetSelectColumns = `
    id,
    name,
    asset_type_id,
    data_category_id,
    owner,
    description,
    environment,
    is_isolated,
    tags,
    created_at,
    updated_at
`

func scanAsset(row pgx.Row, a *domain.Asset) error {
	return row.Scan(
		&a.ID,
		&a.Name,
		&a.AssetTypeID,
		&a.DataCategoryID,
		&a.Owner,
		&a.Description,
		&a.Environment,
		&a.IsIsolated,
		&a.Tags,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
}

func (r *assetRepository) Create(ctx context.Context, a *domain.Asset) error {
	const q = `
INSERT INTO assets (
    name,
    asset_type_id,
    data_category_id,
    owner,
    description,
    environment,
    is_isolated,
    tags
) VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8
) RETURNING id, created_at, updated_at
`
	row := r.pool.QueryRow(ctx, q,
		a.Name,
		a.AssetTypeID,
		a.DataCategoryID,
		a.Owner,
		a.Description,
		a.Environment,
		a.IsIsolated,
		a.Tags,
	)

	if err := row.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return fmt.Errorf("scan created asset: %w", err)
	}

	return nil
}

func (r *assetRepository) GetByID(ctx context.Context, id int64) (*domain.Asset, error) {
	q := `SELECT ` + assetSelectColumns + `FROM assets WHERE id = $1`

	var a domain.Asset
	if err := scanAsset(r.pool.QueryRow(ctx, q, id), &a); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get asset by id: %w", err)
	}

	return &a, nil
}

func (r *assetRepository) List(ctx context.Context, f AssetFilter) ([]domain.Asset, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	q := `SELECT ` + assetSelectColumns + `FROM assets ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, q, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	var res []domain.Asset
	for rows.Next() {
		var a domain.Asset
		if err := scanAsset(rows, &a); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		res = append(res, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return res, nil
}

func (r *assetRepository) Update(ctx context.Context, a *domain.Asset) error {
	const q = `
UPDATE assets
SET
    name             = $1,
    asset_type_id    = $2,
    data_category_id = $3,
    owner            = $4,
    description      = $5,
    environment      = $6,
    is_isolated      = $7,
    tags             = $8,
    updated_at       = now()
WHERE id = $9
RETURNING updated_at
`
	row := r.pool.QueryRow(ctx, q,
		a.Name,
		a.AssetTypeID,
		a.DataCategoryID,
		a.Owner,
		a.Description,
		a.Environment,
		a.IsIsolated,
		a.Tags,
		a.ID,
	)

	if err := row.Scan(&a.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("asset not found")
		}
		return fmt.Errorf("update asset: %w", err)
	}
	return nil
}

func (r *assetRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM assets WHERE id = $1`
	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("asset not found")
	}
	return nil
}
