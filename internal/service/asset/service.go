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

type CreateAssetInput struct {
	Name                string                 `json:"name"`
	Type                string                 `json:"type"`
	AssetTypeID         *int16                 `json:"asset_type_id"`
	Owner               *string                `json:"owner"`
	Description         *string                `json:"description"`
	Location            *string                `json:"location"`
	BusinessCriticality int16                  `json:"business_criticality"`
	Environment         string                 `json:"environment"`
	Tags                map[string]interface{} `json:"tags"`
	// Регуляторные поля
	DataCategory       *string `json:"data_category"`
	ProtectionLevel    *string `json:"protection_level"`
	KIICategory        *string `json:"kii_category"`
	HasPersonalData    bool    `json:"has_personal_data"`
	PersonalDataVolume *string `json:"personal_data_volume"`
	HasInternetAccess  bool    `json:"has_internet_access"`
	IsIsolated         bool    `json:"is_isolated"`
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

	// Автоматически рассчитываем бизнес-критичность на основе регуляторных полей
	businessCriticality := CalculateCriticality(
		in.DataCategory,
		in.ProtectionLevel,
		in.KIICategory,
		in.HasPersonalData,
		in.PersonalDataVolume,
		in.HasInternetAccess,
		in.IsIsolated,
		in.Environment,
	)

	// Автоматически рассчитываем CIA на основе критичности
	c, i, a := CalculateCIA(in.Type, in.Environment, businessCriticality)

	// Подготавливаем тип актива (nullable)
	var assetType *string
	if in.Type != "" {
		assetType = &in.Type
	}

	// Конвертируем регуляторные поля
	var dataCategory *domain.DataCategory
	if in.DataCategory != nil && *in.DataCategory != "" {
		dc := domain.DataCategory(*in.DataCategory)
		dataCategory = &dc
	}

	var protectionLevel *domain.ProtectionLevel
	if in.ProtectionLevel != nil && *in.ProtectionLevel != "" {
		pl := domain.ProtectionLevel(*in.ProtectionLevel)
		protectionLevel = &pl
	}

	var kiiCategory *domain.KIICategory
	if in.KIICategory != nil && *in.KIICategory != "" {
		kii := domain.KIICategory(*in.KIICategory)
		kiiCategory = &kii
	}

	asset := &domain.Asset{
		Name:                in.Name,
		Type:                assetType,
		AssetTypeID:         in.AssetTypeID,
		Owner:               in.Owner,
		Description:         in.Description,
		Location:            in.Location,
		BusinessCriticality: businessCriticality,
		Confidentiality:     c,
		Integrity:           i,
		Availability:        a,
		Environment:         env,
		Tags:                tagsBytes,
		// Регуляторные поля
		DataCategory:       dataCategory,
		ProtectionLevel:    protectionLevel,
		KIICategory:        kiiCategory,
		HasPersonalData:    in.HasPersonalData,
		PersonalDataVolume: in.PersonalDataVolume,
		HasInternetAccess:  in.HasInternetAccess,
		IsIsolated:         in.IsIsolated,
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

	// Обновляем поля
	a.Name = in.Name
	if in.Type != "" {
		a.Type = &in.Type
	} else {
		a.Type = nil
	}
	a.AssetTypeID = in.AssetTypeID
	a.Owner = in.Owner
	a.Description = in.Description
	a.Location = in.Location

	env := domain.AssetEnvironment(in.Environment)
	if env == "" {
		env = domain.AssetEnvProd
	}
	a.Environment = env

	// Автоматически рассчитываем бизнес-критичность на основе регуляторных полей
	a.BusinessCriticality = CalculateCriticality(
		in.DataCategory,
		in.ProtectionLevel,
		in.KIICategory,
		in.HasPersonalData,
		in.PersonalDataVolume,
		in.HasInternetAccess,
		in.IsIsolated,
		in.Environment,
	)

	// Пересчитываем CIA автоматически на основе новой критичности
	c, i, avail := CalculateCIA(in.Type, in.Environment, a.BusinessCriticality)
	a.Confidentiality = c
	a.Integrity = i
	a.Availability = avail

	if in.Tags != nil {
		b, err := json.Marshal(in.Tags)
		if err != nil {
			return nil, fmt.Errorf("marshal tags: %w", err)
		}
		a.Tags = b
	}

	// Обновляем регуляторные поля
	if in.DataCategory != nil && *in.DataCategory != "" {
		dc := domain.DataCategory(*in.DataCategory)
		a.DataCategory = &dc
	} else {
		a.DataCategory = nil
	}

	if in.ProtectionLevel != nil && *in.ProtectionLevel != "" {
		pl := domain.ProtectionLevel(*in.ProtectionLevel)
		a.ProtectionLevel = &pl
	} else {
		a.ProtectionLevel = nil
	}

	if in.KIICategory != nil && *in.KIICategory != "" {
		kii := domain.KIICategory(*in.KIICategory)
		a.KIICategory = &kii
	} else {
		a.KIICategory = nil
	}

	a.HasPersonalData = in.HasPersonalData
	a.PersonalDataVolume = in.PersonalDataVolume
	a.HasInternetAccess = in.HasInternetAccess
	a.IsIsolated = in.IsIsolated

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
