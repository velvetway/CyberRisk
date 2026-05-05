package threat

import (
	"context"
	"fmt"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

type Service interface {
	Create(ctx context.Context, in CreateThreatInput) (*domain.Threat, error)
	Get(ctx context.Context, id int64) (*domain.Threat, error)
	List(ctx context.Context, limit, offset int32) ([]domain.Threat, error)
	Update(ctx context.Context, id int64, in UpdateThreatInput) (*domain.Threat, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo repository.ThreatRepository
}

func NewService(repo repository.ThreatRepository) Service {
	return &service{repo: repo}
}

// CreateThreatInput is the API surface for creating/updating a threat under
// the PTSZI W-model.
//
// QThreat   — степень реализации угрозы, ∈ [0,1] (Q^threat в формуле W).
// QSeverity — степень опасности угрозы,  ∈ [0,1] (q^threat в формуле W).
type CreateThreatInput struct {
	Name                string  `json:"name"`
	ThreatCategoryID    *int16  `json:"threat_category_id"`
	SourceType          string  `json:"source_type"` // external|internal|third_party
	Description         *string `json:"description"`
	QThreat             float64 `json:"q_threat"`
	QSeverity           float64 `json:"q_severity"`
	BDUID               *string `json:"bdu_id"`
	AppliesToTargets    *string `json:"applies_to_targets"`
	AppliesToAssetTypes []int16 `json:"applies_to_asset_types"`
	ImpactC             bool    `json:"impact_c"`
	ImpactI             bool    `json:"impact_i"`
	ImpactA             bool    `json:"impact_a"`
}

type UpdateThreatInput = CreateThreatInput

func validateInput(in CreateThreatInput) (domain.ThreatSourceType, error) {
	if in.Name == "" {
		return "", fmt.Errorf("name is required")
	}
	if in.QThreat < 0 || in.QThreat > 1 {
		return "", fmt.Errorf("q_threat must be in [0, 1]")
	}
	if in.QSeverity < 0 || in.QSeverity > 1 {
		return "", fmt.Errorf("q_severity must be in [0, 1]")
	}
	st := domain.ThreatSourceType(in.SourceType)
	switch st {
	case domain.ThreatSourceExternal, domain.ThreatSourceInternal, domain.ThreatSourceThirdParty:
	default:
		return "", fmt.Errorf("invalid source_type (must be external|internal|third_party)")
	}
	return st, nil
}

func (s *service) Create(ctx context.Context, in CreateThreatInput) (*domain.Threat, error) {
	st, err := validateInput(in)
	if err != nil {
		return nil, err
	}

	t := &domain.Threat{
		Name:             in.Name,
		ThreatCategoryID: in.ThreatCategoryID,
		SourceType:       st,
		Description:      in.Description,
		QThreat:          in.QThreat,
		QSeverity:        in.QSeverity,
		BDUID:            in.BDUID,
	}

	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) Get(ctx context.Context, id int64) (*domain.Threat, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, limit, offset int32) ([]domain.Threat, error) {
	return s.repo.List(ctx, repository.ThreatFilter{Limit: limit, Offset: offset})
}

func (s *service) Update(ctx context.Context, id int64, in UpdateThreatInput) (*domain.Threat, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("threat not found")
	}

	st, err := validateInput(in)
	if err != nil {
		return nil, err
	}

	t.Name = in.Name
	t.ThreatCategoryID = in.ThreatCategoryID
	t.SourceType = st
	t.Description = in.Description
	t.QThreat = in.QThreat
	t.QSeverity = in.QSeverity
	t.BDUID = in.BDUID
	t.AppliesToTargets = in.AppliesToTargets
	t.AppliesToAssetTypes = in.AppliesToAssetTypes
	t.ImpactC = in.ImpactC
	t.ImpactI = in.ImpactI
	t.ImpactA = in.ImpactA

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
