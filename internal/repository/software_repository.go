package repository

import (
	"context"
	"fmt"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SoftwareFilter struct {
	Limit       int32
	Offset      int32
	CategoryID  *int16
	IsRussian   *bool
	FSTECOnly   bool
	SearchQuery string
}

type SoftwareRepository interface {
	Create(ctx context.Context, s *domain.Software) error
	GetByID(ctx context.Context, id int64) (*domain.Software, error)
	List(ctx context.Context, f SoftwareFilter) ([]domain.Software, error)
	Update(ctx context.Context, s *domain.Software) error
	Delete(ctx context.Context, id int64) error
	ListCategories(ctx context.Context) ([]domain.SoftwareCategory, error)
	ListAssetSoftware(ctx context.Context, assetID int64) ([]domain.AssetSoftwareWithSoftware, error)
}

type softwareRepository struct {
	pool *pgxpool.Pool
}

func NewSoftwareRepository(pool *pgxpool.Pool) SoftwareRepository {
	return &softwareRepository{pool: pool}
}

func (r *softwareRepository) Create(ctx context.Context, s *domain.Software) error {
	const q = `
INSERT INTO software_catalog (
    name, vendor, version, category_id,
    is_russian, registry_number, registry_date, registry_url,
    fstec_certified, fstec_certificate_num, fstec_certificate_date, fstec_protection_class, fstec_valid_until,
    fsb_certified, fsb_certificate_num, fsb_protection_class,
    description, website
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, $10, $11, $12, $13,
    $14, $15, $16,
    $17, $18
) RETURNING id, created_at, updated_at`

	return r.pool.QueryRow(ctx, q,
		s.Name, s.Vendor, s.Version, s.CategoryID,
		s.IsRussian, s.RegistryNumber, s.RegistryDate, s.RegistryURL,
		s.FSTECCertified, s.FSTECCertificateNum, s.FSTECCertificateDate, s.FSTECProtectionClass, s.FSTECValidUntil,
		s.FSBCertified, s.FSBCertificateNum, s.FSBProtectionClass,
		s.Description, s.Website,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *softwareRepository) GetByID(ctx context.Context, id int64) (*domain.Software, error) {
	const q = `
SELECT
    id, name, vendor, version, category_id,
    is_russian, registry_number, registry_date, registry_url,
    fstec_certified, fstec_certificate_num, fstec_certificate_date, fstec_protection_class, fstec_valid_until,
    fsb_certified, fsb_certificate_num, fsb_protection_class,
    description, website, created_at, updated_at
FROM software_catalog
WHERE id = $1`

	var s domain.Software
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.Name, &s.Vendor, &s.Version, &s.CategoryID,
		&s.IsRussian, &s.RegistryNumber, &s.RegistryDate, &s.RegistryURL,
		&s.FSTECCertified, &s.FSTECCertificateNum, &s.FSTECCertificateDate, &s.FSTECProtectionClass, &s.FSTECValidUntil,
		&s.FSBCertified, &s.FSBCertificateNum, &s.FSBProtectionClass,
		&s.Description, &s.Website, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get software by id: %w", err)
	}
	return &s, nil
}

func (r *softwareRepository) List(ctx context.Context, f SoftwareFilter) ([]domain.Software, error) {
	if f.Limit <= 0 {
		f.Limit = 100
	}

	q := `
SELECT
    id, name, vendor, version, category_id,
    is_russian, registry_number, registry_date, registry_url,
    fstec_certified, fstec_certificate_num, fstec_certificate_date, fstec_protection_class, fstec_valid_until,
    fsb_certified, fsb_certificate_num, fsb_protection_class,
    description, website, created_at, updated_at
FROM software_catalog
WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if f.CategoryID != nil {
		q += fmt.Sprintf(" AND category_id = $%d", argIdx)
		args = append(args, *f.CategoryID)
		argIdx++
	}

	if f.IsRussian != nil {
		q += fmt.Sprintf(" AND is_russian = $%d", argIdx)
		args = append(args, *f.IsRussian)
		argIdx++
	}

	if f.FSTECOnly {
		q += " AND fstec_certified = TRUE"
	}

	if f.SearchQuery != "" {
		q += fmt.Sprintf(" AND (name ILIKE $%d OR vendor ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+f.SearchQuery+"%")
		argIdx++
	}

	q += fmt.Sprintf(" ORDER BY is_russian DESC, name LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list software: %w", err)
	}
	defer rows.Close()

	var result []domain.Software
	for rows.Next() {
		var s domain.Software
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Vendor, &s.Version, &s.CategoryID,
			&s.IsRussian, &s.RegistryNumber, &s.RegistryDate, &s.RegistryURL,
			&s.FSTECCertified, &s.FSTECCertificateNum, &s.FSTECCertificateDate, &s.FSTECProtectionClass, &s.FSTECValidUntil,
			&s.FSBCertified, &s.FSBCertificateNum, &s.FSBProtectionClass,
			&s.Description, &s.Website, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan software: %w", err)
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

func (r *softwareRepository) Update(ctx context.Context, s *domain.Software) error {
	const q = `
UPDATE software_catalog SET
    name = $1, vendor = $2, version = $3, category_id = $4,
    is_russian = $5, registry_number = $6, registry_date = $7, registry_url = $8,
    fstec_certified = $9, fstec_certificate_num = $10, fstec_certificate_date = $11, fstec_protection_class = $12, fstec_valid_until = $13,
    fsb_certified = $14, fsb_certificate_num = $15, fsb_protection_class = $16,
    description = $17, website = $18, updated_at = now()
WHERE id = $19
RETURNING updated_at`

	err := r.pool.QueryRow(ctx, q,
		s.Name, s.Vendor, s.Version, s.CategoryID,
		s.IsRussian, s.RegistryNumber, s.RegistryDate, s.RegistryURL,
		s.FSTECCertified, s.FSTECCertificateNum, s.FSTECCertificateDate, s.FSTECProtectionClass, s.FSTECValidUntil,
		s.FSBCertified, s.FSBCertificateNum, s.FSBProtectionClass,
		s.Description, s.Website, s.ID,
	).Scan(&s.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("software not found")
		}
		return fmt.Errorf("update software: %w", err)
	}
	return nil
}

func (r *softwareRepository) Delete(ctx context.Context, id int64) error {
	ct, err := r.pool.Exec(ctx, "DELETE FROM software_catalog WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete software: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("software not found")
	}
	return nil
}

func (r *softwareRepository) ListCategories(ctx context.Context) ([]domain.SoftwareCategory, error) {
	const q = `SELECT id, code, name, COALESCE(description, '') FROM software_categories ORDER BY name`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list software categories: %w", err)
	}
	defer rows.Close()

	var result []domain.SoftwareCategory
	for rows.Next() {
		var c domain.SoftwareCategory
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		result = append(result, c)
	}

	return result, rows.Err()
}

func (r *softwareRepository) ListAssetSoftware(ctx context.Context, assetID int64) ([]domain.AssetSoftwareWithSoftware, error) {
	const q = `
SELECT
    asw.id, asw.asset_id, asw.software_id, asw.version, asw.install_date,
    asw.license_type, asw.license_expires, asw.notes, asw.created_at, asw.updated_at,
    sc.id, sc.name, sc.vendor, sc.version, sc.category_id,
    sc.is_russian, sc.registry_number, sc.registry_date, sc.registry_url,
    sc.fstec_certified, sc.fstec_certificate_num, sc.fstec_certificate_date, sc.fstec_protection_class, sc.fstec_valid_until,
    sc.fsb_certified, sc.fsb_certificate_num, sc.fsb_protection_class,
    sc.description, sc.website, sc.created_at, sc.updated_at
FROM asset_software AS asw
JOIN software_catalog AS sc ON sc.id = asw.software_id
WHERE asw.asset_id = $1
ORDER BY sc.name`

	rows, err := r.pool.Query(ctx, q, assetID)
	if err != nil {
		return nil, fmt.Errorf("list asset software: %w", err)
	}
	defer rows.Close()

	var result []domain.AssetSoftwareWithSoftware
	for rows.Next() {
		var link domain.AssetSoftware
		var sw domain.Software
		if err := rows.Scan(
			&link.ID, &link.AssetID, &link.SoftwareID, &link.Version, &link.InstallDate,
			&link.LicenseType, &link.LicenseExpires, &link.Notes, &link.CreatedAt, &link.UpdatedAt,
			&sw.ID, &sw.Name, &sw.Vendor, &sw.Version, &sw.CategoryID,
			&sw.IsRussian, &sw.RegistryNumber, &sw.RegistryDate, &sw.RegistryURL,
			&sw.FSTECCertified, &sw.FSTECCertificateNum, &sw.FSTECCertificateDate, &sw.FSTECProtectionClass, &sw.FSTECValidUntil,
			&sw.FSBCertified, &sw.FSBCertificateNum, &sw.FSBProtectionClass,
			&sw.Description, &sw.Website, &sw.CreatedAt, &sw.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan asset software: %w", err)
		}
		result = append(result, domain.AssetSoftwareWithSoftware{
			Link:     link,
			Software: sw,
		})
	}

	return result, rows.Err()
}
