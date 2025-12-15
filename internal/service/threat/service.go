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

type CreateThreatInput struct {
	Name             string  `json:"name"`
	ThreatCategoryID *int16  `json:"threat_category_id"`
	SourceType       string  `json:"source_type"` // external|internal|third_party
	Description      *string `json:"description"`
	BaseLikelihood   int16   `json:"base_likelihood"` // 1–5
}

type UpdateThreatInput = CreateThreatInput

func (s *service) Create(ctx context.Context, in CreateThreatInput) (*domain.Threat, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if in.BaseLikelihood < 1 || in.BaseLikelihood > 5 {
		return nil, fmt.Errorf("base_likelihood must be between 1 and 5")
	}

	st := domain.ThreatSourceType(in.SourceType)
	switch st {
	case domain.ThreatSourceExternal, domain.ThreatSourceInternal, domain.ThreatSourceThirdParty:
	default:
		return nil, fmt.Errorf("invalid source_type (must be external|internal|third_party)")
	}

	t := &domain.Threat{
		Name:             in.Name,
		ThreatCategoryID: in.ThreatCategoryID,
		SourceType:       st,
		Description:      in.Description,
		BaseLikelihood:   in.BaseLikelihood,
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

	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if in.BaseLikelihood < 1 || in.BaseLikelihood > 5 {
		return nil, fmt.Errorf("base_likelihood must be between 1 and 5")
	}

	st := domain.ThreatSourceType(in.SourceType)
	switch st {
	case domain.ThreatSourceExternal, domain.ThreatSourceInternal, domain.ThreatSourceThirdParty:
	default:
		return nil, fmt.Errorf("invalid source_type (must be external|internal|third_party)")
	}

	t.Name = in.Name
	t.ThreatCategoryID = in.ThreatCategoryID
	t.SourceType = st
	t.Description = in.Description
	t.BaseLikelihood = in.BaseLikelihood

	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
