package repository

import (
	"context"
	"fmt"

	"Diplom/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ComplianceRepository — read-only справочник стандартов и требований +
// маппинг requirement ↔ control. Запись — только через миграции.
type ComplianceRepository interface {
	ListStandards(ctx context.Context) ([]domain.ComplianceStandard, error)
	GetStandardByCode(ctx context.Context, code string) (*domain.ComplianceStandard, error)
	GetStandardByID(ctx context.Context, id int16) (*domain.ComplianceStandard, error)
	ListRequirements(ctx context.Context, standardID int16) ([]domain.ComplianceRequirement, error)
	ListAllRequirementControlLinks(ctx context.Context) ([]domain.RequirementControlLink, error)
	ListRequirementControlLinks(ctx context.Context, standardID int16) ([]domain.RequirementControlLink, error)
}

type complianceRepository struct {
	db *pgxpool.Pool
}

func NewComplianceRepository(db *pgxpool.Pool) ComplianceRepository {
	return &complianceRepository{db: db}
}

func (r *complianceRepository) ListStandards(ctx context.Context) ([]domain.ComplianceStandard, error) {
	const q = `
SELECT id, code, name, full_name, jurisdiction, description, sort_order
FROM compliance_standards
ORDER BY sort_order, id`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list standards: %w", err)
	}
	defer rows.Close()
	out := make([]domain.ComplianceStandard, 0, 4)
	for rows.Next() {
		var s domain.ComplianceStandard
		if err := rows.Scan(&s.ID, &s.Code, &s.Name, &s.FullName, &s.Jurisdiction, &s.Description, &s.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *complianceRepository) GetStandardByCode(ctx context.Context, code string) (*domain.ComplianceStandard, error) {
	const q = `
SELECT id, code, name, full_name, jurisdiction, description, sort_order
FROM compliance_standards
WHERE code = $1`
	var s domain.ComplianceStandard
	if err := r.db.QueryRow(ctx, q, code).Scan(
		&s.ID, &s.Code, &s.Name, &s.FullName, &s.Jurisdiction, &s.Description, &s.SortOrder,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *complianceRepository) GetStandardByID(ctx context.Context, id int16) (*domain.ComplianceStandard, error) {
	const q = `
SELECT id, code, name, full_name, jurisdiction, description, sort_order
FROM compliance_standards
WHERE id = $1`
	var s domain.ComplianceStandard
	if err := r.db.QueryRow(ctx, q, id).Scan(
		&s.ID, &s.Code, &s.Name, &s.FullName, &s.Jurisdiction, &s.Description, &s.SortOrder,
	); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *complianceRepository) ListRequirements(ctx context.Context, standardID int16) ([]domain.ComplianceRequirement, error) {
	const q = `
SELECT id, standard_id, code, category, title, description, priority, sort_order
FROM compliance_requirements
WHERE standard_id = $1
ORDER BY sort_order, code`
	rows, err := r.db.Query(ctx, q, standardID)
	if err != nil {
		return nil, fmt.Errorf("list requirements: %w", err)
	}
	defer rows.Close()
	out := make([]domain.ComplianceRequirement, 0, 32)
	for rows.Next() {
		var r domain.ComplianceRequirement
		if err := rows.Scan(&r.ID, &r.StandardID, &r.Code, &r.Category, &r.Title, &r.Description, &r.Priority, &r.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (r *complianceRepository) ListAllRequirementControlLinks(ctx context.Context) ([]domain.RequirementControlLink, error) {
	const q = `SELECT requirement_id, control_id, coverage_weight FROM requirement_controls`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list rc links: %w", err)
	}
	defer rows.Close()
	out := make([]domain.RequirementControlLink, 0, 128)
	for rows.Next() {
		var l domain.RequirementControlLink
		if err := rows.Scan(&l.RequirementID, &l.ControlID, &l.CoverageWeight); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *complianceRepository) ListRequirementControlLinks(ctx context.Context, standardID int16) ([]domain.RequirementControlLink, error) {
	const q = `
SELECT rc.requirement_id, rc.control_id, rc.coverage_weight
FROM requirement_controls rc
JOIN compliance_requirements r ON r.id = rc.requirement_id
WHERE r.standard_id = $1`
	rows, err := r.db.Query(ctx, q, standardID)
	if err != nil {
		return nil, fmt.Errorf("list rc links: %w", err)
	}
	defer rows.Close()
	out := make([]domain.RequirementControlLink, 0, 64)
	for rows.Next() {
		var l domain.RequirementControlLink
		if err := rows.Scan(&l.RequirementID, &l.ControlID, &l.CoverageWeight); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
