package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"Diplom/internal/domain"
)

type MinreestrFilter struct {
	Query       string
	CategoryID  *int16
	RussianOnly bool
	FSTECOnly   bool
	Limit       int
}

type MinreestrRepository interface {
	IsAvailable() bool
	Search(ctx context.Context, f MinreestrFilter) ([]domain.Software, error)
}

type minreestrSQLiteRepository struct {
	db *sql.DB
}

func NewMinreestrSQLiteRepository(path string) (MinreestrRepository, error) {
	if path == "" {
		return nil, nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("open minreestr sqlite: %w", err)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro&cache=shared")
	if err != nil {
		return nil, fmt.Errorf("open minreestr sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping minreestr sqlite: %w", err)
	}
	return &minreestrSQLiteRepository{db: db}, nil
}

func (r *minreestrSQLiteRepository) IsAvailable() bool {
	return r != nil && r.db != nil
}

func (r *minreestrSQLiteRepository) Search(ctx context.Context, f MinreestrFilter) ([]domain.Software, error) {
	if !r.IsAvailable() {
		return nil, fmt.Errorf("minreestr sqlite is not configured")
	}
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	args := make([]any, 0)
	filters := []string{"1=1"}
	if f.Query != "" {
		q := "%" + f.Query + "%"
		filters = append(filters, "(name LIKE ? OR vendor LIKE ? OR description LIKE ? OR registry_number LIKE ?)")
		args = append(args, q, q, q, q)
	}
	if f.CategoryID != nil {
		filters = append(filters, "category_id = ?")
		args = append(args, *f.CategoryID)
	}
	if f.RussianOnly {
		filters = append(filters, "is_russian IN ('1','t','true',1)")
	}
	if f.FSTECOnly {
		filters = append(filters, "fstec_certified IN ('1','t','true',1)")
	}

	q := fmt.Sprintf(`
SELECT
    id, name, vendor, nullif(version, ''), nullif(category_id, 0),
    is_russian, nullif(registry_number, ''), nullif(registry_url, ''),
    fstec_certified, nullif(fstec_certificate_num, ''), nullif(fstec_protection_class, ''),
    fsb_certified, nullif(fsb_certificate_num, ''), nullif(fsb_protection_class, ''),
    nullif(description, ''), nullif(website, '')
FROM software
WHERE %s
ORDER BY is_russian DESC, fstec_certified DESC, name
LIMIT ?`, strings.Join(filters, " AND "))
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("search minreestr: %w", err)
	}
	defer rows.Close()

	out := make([]domain.Software, 0)
	for rows.Next() {
		var s domain.Software
		var version, registryNumber, registryURL, certNum, protectionClass, fsbCertNum, fsbClass, description, website sql.NullString
		var categoryID sql.NullInt64
		var isRussian, fstecCertified, fsbCertified any
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Vendor, &version, &categoryID,
			&isRussian, &registryNumber, &registryURL,
			&fstecCertified, &certNum, &protectionClass,
			&fsbCertified, &fsbCertNum, &fsbClass,
			&description, &website,
		); err != nil {
			return nil, fmt.Errorf("scan minreestr software: %w", err)
		}
		s.Version = nullStringPtr(version)
		if categoryID.Valid {
			v := int16(categoryID.Int64)
			s.CategoryID = &v
		}
		s.IsRussian = sqliteBool(isRussian)
		s.RegistryNumber = nullStringPtr(registryNumber)
		s.RegistryURL = nullStringPtr(registryURL)
		s.FSTECCertified = sqliteBool(fstecCertified)
		s.FSTECCertificateNum = nullStringPtr(certNum)
		s.FSTECProtectionClass = nullStringPtr(protectionClass)
		s.FSBCertified = sqliteBool(fsbCertified)
		s.FSBCertificateNum = nullStringPtr(fsbCertNum)
		s.FSBProtectionClass = nullStringPtr(fsbClass)
		s.Description = nullStringPtr(description)
		s.Website = nullStringPtr(website)
		out = append(out, s)
	}
	return out, rows.Err()
}

func sqliteBool(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case int64:
		return x != 0
	case string:
		switch strings.ToLower(x) {
		case "1", "t", "true", "yes":
			return true
		default:
			return false
		}
	case []byte:
		return sqliteBool(string(x))
	default:
		return false
	}
}
