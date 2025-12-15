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

func (r *assetRepository) Create(ctx context.Context, a *domain.Asset) error {
	const q = `
INSERT INTO assets (
    name,
    type,
    asset_type_id,
    owner,
    description,
    location,
    business_criticality,
    confidentiality,
    integrity,
    availability,
    environment,
    tags,
    data_category,
    protection_level,
    kii_category,
    has_personal_data,
    personal_data_volume,
    has_internet_access,
    is_isolated
) VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19
) RETURNING
    id,
    created_at,
    updated_at
`
	row := r.pool.QueryRow(ctx, q,
		a.Name,
		a.Type,
		a.AssetTypeID,
		a.Owner,
		a.Description,
		a.Location,
		a.BusinessCriticality,
		a.Confidentiality,
		a.Integrity,
		a.Availability,
		a.Environment,
		a.Tags,
		a.DataCategory,
		a.ProtectionLevel,
		a.KIICategory,
		a.HasPersonalData,
		a.PersonalDataVolume,
		a.HasInternetAccess,
		a.IsIsolated,
	)

	if err := row.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return fmt.Errorf("scan created asset: %w", err)
	}

	return nil
}

func (r *assetRepository) GetByID(ctx context.Context, id int64) (*domain.Asset, error) {
	const q = `
SELECT
    id,
    name,
    type,
    asset_type_id,
    owner,
    description,
    location,
    business_criticality,
    confidentiality,
    integrity,
    availability,
    environment,
    tags,
    data_category,
    protection_level,
    kii_category,
    has_personal_data,
    personal_data_volume,
    has_internet_access,
    is_isolated,
    created_at,
    updated_at
FROM assets
WHERE id = $1
`
	var a domain.Asset
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&a.ID,
		&a.Name,
		&a.Type,
		&a.AssetTypeID,
		&a.Owner,
		&a.Description,
		&a.Location,
		&a.BusinessCriticality,
		&a.Confidentiality,
		&a.Integrity,
		&a.Availability,
		&a.Environment,
		&a.Tags,
		&a.DataCategory,
		&a.ProtectionLevel,
		&a.KIICategory,
		&a.HasPersonalData,
		&a.PersonalDataVolume,
		&a.HasInternetAccess,
		&a.IsIsolated,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
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

	const q = `
SELECT
    id,
    name,
    type,
    asset_type_id,
    owner,
    description,
    location,
    business_criticality,
    confidentiality,
    integrity,
    availability,
    environment,
    tags,
    data_category,
    protection_level,
    kii_category,
    has_personal_data,
    personal_data_volume,
    has_internet_access,
    is_isolated,
    created_at,
    updated_at
FROM assets
ORDER BY id
LIMIT $1 OFFSET $2
`
	rows, err := r.pool.Query(ctx, q, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	var res []domain.Asset
	for rows.Next() {
		var a domain.Asset
		if err := rows.Scan(
			&a.ID,
			&a.Name,
			&a.Type,
			&a.AssetTypeID,
			&a.Owner,
			&a.Description,
			&a.Location,
			&a.BusinessCriticality,
			&a.Confidentiality,
			&a.Integrity,
			&a.Availability,
			&a.Environment,
			&a.Tags,
			&a.DataCategory,
			&a.ProtectionLevel,
			&a.KIICategory,
			&a.HasPersonalData,
			&a.PersonalDataVolume,
			&a.HasInternetAccess,
			&a.IsIsolated,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
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
    name = $1,
    type = $2,
    asset_type_id = $3,
    owner = $4,
    description = $5,
    location = $6,
    business_criticality = $7,
    confidentiality = $8,
    integrity = $9,
    availability = $10,
    environment = $11,
    tags = $12,
    data_category = $13,
    protection_level = $14,
    kii_category = $15,
    has_personal_data = $16,
    personal_data_volume = $17,
    has_internet_access = $18,
    is_isolated = $19,
    updated_at = now()
WHERE id = $20
RETURNING updated_at
`
	row := r.pool.QueryRow(ctx, q,
		a.Name,
		a.Type,
		a.AssetTypeID,
		a.Owner,
		a.Description,
		a.Location,
		a.BusinessCriticality,
		a.Confidentiality,
		a.Integrity,
		a.Availability,
		a.Environment,
		a.Tags,
		a.DataCategory,
		a.ProtectionLevel,
		a.KIICategory,
		a.HasPersonalData,
		a.PersonalDataVolume,
		a.HasInternetAccess,
		a.IsIsolated,
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
