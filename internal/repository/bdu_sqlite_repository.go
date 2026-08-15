package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"Diplom/internal/domain"

	_ "modernc.org/sqlite"
)

type BDUSearchFilter struct {
	Query       string
	Software    string
	Vendor      string
	MinCVSS     *float64
	MinSeverity *int16
	Limit       int
}

type BDURepository interface {
	IsAvailable() bool
	Search(ctx context.Context, f BDUSearchFilter) ([]domain.BDUVulnerability, error)
	GetByID(ctx context.Context, id string) (*domain.BDUVulnerability, error)
}

type bduSQLiteRepository struct {
	db *sql.DB
}

func NewBDUSQLiteRepository(path string) (BDURepository, error) {
	if path == "" {
		return nil, nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("open bdu sqlite: %w", err)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro&cache=shared")
	if err != nil {
		return nil, fmt.Errorf("open bdu sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping bdu sqlite: %w", err)
	}
	return &bduSQLiteRepository{db: db}, nil
}

func (r *bduSQLiteRepository) IsAvailable() bool {
	return r != nil && r.db != nil
}

func (r *bduSQLiteRepository) Search(ctx context.Context, f BDUSearchFilter) ([]domain.BDUVulnerability, error) {
	if !r.IsAvailable() {
		return nil, fmt.Errorf("bdu sqlite is not configured")
	}
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	args := make([]any, 0)
	filters := []string{"1=1"}
	joinSoftware := false

	if f.Query != "" {
		q := "%" + f.Query + "%"
		filters = append(filters, "(v.id LIKE ? OR v.name LIKE ? OR v.description LIKE ? OR v.cves_joined LIKE ? OR v.software_names LIKE ? OR v.vendors LIKE ?)")
		args = append(args, q, q, q, q, q, q)
	}
	if f.Software != "" {
		joinSoftware = true
		q := "%" + f.Software + "%"
		filters = append(filters, "(s.name LIKE ? OR v.software_names LIKE ?)")
		args = append(args, q, q)
	}
	if f.Vendor != "" {
		joinSoftware = true
		q := "%" + f.Vendor + "%"
		filters = append(filters, "(s.vendor LIKE ? OR v.vendors LIKE ?)")
		args = append(args, q, q)
	}
	if f.MinCVSS != nil {
		filters = append(filters, "v.cvss_score >= ?")
		args = append(args, *f.MinCVSS)
	}
	if f.MinSeverity != nil {
		filters = append(filters, "v.severity_level >= ?")
		args = append(args, *f.MinSeverity)
	}

	from := "vulnerabilities v"
	if joinSoftware {
		from = "vulnerabilities v LEFT JOIN software s ON s.bdu_id = v.id"
	}

	q := fmt.Sprintf(`
SELECT DISTINCT
    v.id, v.name, v.description, v.severity, v.severity_level,
    v.cvss_score, v.cvss_vector, v.cves_joined, v.vendors, v.software_names,
    v.solution, v.has_exploit, v.has_fix
FROM %s
WHERE %s
ORDER BY COALESCE(v.cvss_score, 0) DESC, v.severity_level DESC, v.id
LIMIT ?`, from, strings.Join(filters, " AND "))
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("search bdu vulnerabilities: %w", err)
	}
	defer rows.Close()

	return scanBDUVulnerabilities(rows)
}

func (r *bduSQLiteRepository) GetByID(ctx context.Context, id string) (*domain.BDUVulnerability, error) {
	if !r.IsAvailable() {
		return nil, fmt.Errorf("bdu sqlite is not configured")
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT
    id, name, description, severity, severity_level,
    cvss_score, cvss_vector, cves_joined, vendors, software_names,
    solution, has_exploit, has_fix
FROM vulnerabilities
WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get bdu vulnerability: %w", err)
	}
	defer rows.Close()

	items, err := scanBDUVulnerabilities(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func scanBDUVulnerabilities(rows *sql.Rows) ([]domain.BDUVulnerability, error) {
	out := make([]domain.BDUVulnerability, 0)
	for rows.Next() {
		var v domain.BDUVulnerability
		var description, severity, cvssVector, cves, vendors, softwareNames, solution sql.NullString
		var cvss sql.NullFloat64
		var hasExploit, hasFix int
		if err := rows.Scan(
			&v.ID, &v.Name, &description, &severity, &v.SeverityLevel,
			&cvss, &cvssVector, &cves, &vendors, &softwareNames,
			&solution, &hasExploit, &hasFix,
		); err != nil {
			return nil, fmt.Errorf("scan bdu vulnerability: %w", err)
		}
		v.Description = nullStringPtr(description)
		v.Severity = nullStringPtr(severity)
		if cvss.Valid {
			v.CVSSScore = &cvss.Float64
		}
		v.CVSSVector = nullStringPtr(cvssVector)
		v.CVEs = nullStringPtr(cves)
		v.Vendors = nullStringPtr(vendors)
		v.SoftwareNames = nullStringPtr(softwareNames)
		v.Solution = nullStringPtr(solution)
		v.HasExploit = hasExploit != 0
		v.HasFix = hasFix != 0
		out = append(out, v)
	}
	return out, rows.Err()
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}
