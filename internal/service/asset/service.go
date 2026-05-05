package asset

import (
	"context"
	"encoding/json"
	"fmt"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

type Service interface {
	Create(ctx context.Context, in CreateAssetInput) (*domain.Asset, error)
	Get(ctx context.Context, id int64) (*domain.Asset, error)
	List(ctx context.Context, limit, offset int32) ([]domain.Asset, error)
	Update(ctx context.Context, id int64, in UpdateAssetInput) (*domain.Asset, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo repository.AssetRepository
}

func NewService(repo repository.AssetRepository) Service {
	return &service{repo: repo}
}

// CreateAssetInput is the API surface for creating/updating an asset under
// the PTSZI W-model. Only fields that have meaning in the W formula or are
// needed for UI/labelling are accepted.
type CreateAssetInput struct {
	Name        string                 `json:"name"`
	AssetTypeID *int16                 `json:"asset_type_id"`
	Owner       *string                `json:"owner"`
	Description *string                `json:"description"`
	Environment string                 `json:"environment"`
	IsIsolated  bool                   `json:"is_isolated"`
	Tags        map[string]interface{} `json:"tags"`
}

type UpdateAssetInput = CreateAssetInput

func (s *service) Create(ctx context.Context, in CreateAssetInput) (*domain.Asset, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	env := domain.AssetEnvironment(in.Environment)
	if env == "" {
		env = domain.AssetEnvProd
	}

	var tagsBytes []byte
	if in.Tags != nil {
		b, err := json.Marshal(in.Tags)
		if err != nil {
			return nil, fmt.Errorf("marshal tags: %w", err)
		}
		tagsBytes = b
	}

	asset := &domain.Asset{
		Name:        in.Name,
		AssetTypeID: in.AssetTypeID,
		Owner:       in.Owner,
		Description: in.Description,
		Environment: env,
		IsIsolated:  in.IsIsolated,
		Tags:        tagsBytes,
	}

	if err := s.repo.Create(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *service) Get(ctx context.Context, id int64) (*domain.Asset, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, limit, offset int32) ([]domain.Asset, error) {
	return s.repo.List(ctx, repository.AssetFilter{Limit: limit, Offset: offset})
}

func (s *service) Update(ctx context.Context, id int64, in UpdateAssetInput) (*domain.Asset, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, fmt.Errorf("asset not found")
	}

	a.Name = in.Name
	a.AssetTypeID = in.AssetTypeID
	a.Owner = in.Owner
	a.Description = in.Description

	env := domain.AssetEnvironment(in.Environment)
	if env == "" {
		env = domain.AssetEnvProd
	}
	a.Environment = env
	a.IsIsolated = in.IsIsolated

	if in.Tags != nil {
		b, err := json.Marshal(in.Tags)
		if err != nil {
			return nil, fmt.Errorf("marshal tags: %w", err)
		}
		a.Tags = b
	}

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
