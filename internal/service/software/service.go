package software

import (
	"context"
	"fmt"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

type Service interface {
	Create(ctx context.Context, in CreateSoftwareInput) (*domain.Software, error)
	Get(ctx context.Context, id int64) (*domain.Software, error)
	List(ctx context.Context, f ListFilter) ([]domain.Software, error)
	Update(ctx context.Context, id int64, in UpdateSoftwareInput) (*domain.Software, error)
	Delete(ctx context.Context, id int64) error
	ListCategories(ctx context.Context) ([]domain.SoftwareCategory, error)

	// Поиск российских аналогов
	FindRussianAlternatives(ctx context.Context, categoryID int16) ([]domain.Software, error)
	// Поиск сертифицированного ПО
	FindCertifiedSoftware(ctx context.Context, categoryID int16, fstecOnly bool) ([]domain.Software, error)
	// Предложение замены конкретному ПО на российские аналоги
	SuggestAlternatives(ctx context.Context, softwareID int64) ([]domain.Software, error)
	// Рекомендации по замене ПО на конкретном активе
	SuggestAlternativesForAsset(ctx context.Context, assetID int64) ([]AssetSoftwareAlternative, error)
}

type service struct {
	repo repository.SoftwareRepository
}

func NewService(repo repository.SoftwareRepository) Service {
	return &service{repo: repo}
}

type CreateSoftwareInput struct {
	Name                 string  `json:"name"`
	Vendor               string  `json:"vendor"`
	Version              *string `json:"version"`
	CategoryID           *int16  `json:"category_id"`
	IsRussian            bool    `json:"is_russian"`
	RegistryNumber       *string `json:"registry_number"`
	RegistryDate         *string `json:"registry_date"`
	RegistryURL          *string `json:"registry_url"`
	FSTECCertified       bool    `json:"fstec_certified"`
	FSTECCertificateNum  *string `json:"fstec_certificate_num"`
	FSTECCertificateDate *string `json:"fstec_certificate_date"`
	FSTECProtectionClass *string `json:"fstec_protection_class"`
	FSTECValidUntil      *string `json:"fstec_valid_until"`
	FSBCertified         bool    `json:"fsb_certified"`
	FSBCertificateNum    *string `json:"fsb_certificate_num"`
	FSBProtectionClass   *string `json:"fsb_protection_class"`
	Description          *string `json:"description"`
	Website              *string `json:"website"`
}

type UpdateSoftwareInput = CreateSoftwareInput

type ListFilter struct {
	Limit       int32   `json:"limit"`
	Offset      int32   `json:"offset"`
	CategoryID  *int16  `json:"category_id"`
	IsRussian   *bool   `json:"is_russian"`
	FSTECOnly   bool    `json:"fstec_only"`
	SearchQuery string  `json:"search"`
}

// AssetSoftwareAlternative описывает ПО на активе и список российских аналогов
type AssetSoftwareAlternative struct {
	AssetSoftwareID int64            `json:"asset_software_id"`
	AssetID         int64            `json:"asset_id"`
	Version         *string          `json:"version,omitempty"`
	Notes           *string          `json:"notes,omitempty"`
	Software        domain.Software  `json:"software"`
	Alternatives    []domain.Software `json:"alternatives"`
}

func (s *service) Create(ctx context.Context, in CreateSoftwareInput) (*domain.Software, error) {
	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if in.Vendor == "" {
		return nil, fmt.Errorf("vendor is required")
	}

	sw := &domain.Software{
		Name:                 in.Name,
		Vendor:               in.Vendor,
		Version:              in.Version,
		CategoryID:           in.CategoryID,
		IsRussian:            in.IsRussian,
		RegistryNumber:       in.RegistryNumber,
		RegistryDate:         in.RegistryDate,
		RegistryURL:          in.RegistryURL,
		FSTECCertified:       in.FSTECCertified,
		FSTECCertificateNum:  in.FSTECCertificateNum,
		FSTECCertificateDate: in.FSTECCertificateDate,
		FSTECProtectionClass: in.FSTECProtectionClass,
		FSTECValidUntil:      in.FSTECValidUntil,
		FSBCertified:         in.FSBCertified,
		FSBCertificateNum:    in.FSBCertificateNum,
		FSBProtectionClass:   in.FSBProtectionClass,
		Description:          in.Description,
		Website:              in.Website,
	}

	if err := s.repo.Create(ctx, sw); err != nil {
		return nil, err
	}

	return sw, nil
}

func (s *service) Get(ctx context.Context, id int64) (*domain.Software, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, f ListFilter) ([]domain.Software, error) {
	return s.repo.List(ctx, repository.SoftwareFilter{
		Limit:       f.Limit,
		Offset:      f.Offset,
		CategoryID:  f.CategoryID,
		IsRussian:   f.IsRussian,
		FSTECOnly:   f.FSTECOnly,
		SearchQuery: f.SearchQuery,
	})
}

func (s *service) Update(ctx context.Context, id int64, in UpdateSoftwareInput) (*domain.Software, error) {
	sw, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sw == nil {
		return nil, fmt.Errorf("software not found")
	}

	sw.Name = in.Name
	sw.Vendor = in.Vendor
	sw.Version = in.Version
	sw.CategoryID = in.CategoryID
	sw.IsRussian = in.IsRussian
	sw.RegistryNumber = in.RegistryNumber
	sw.RegistryDate = in.RegistryDate
	sw.RegistryURL = in.RegistryURL
	sw.FSTECCertified = in.FSTECCertified
	sw.FSTECCertificateNum = in.FSTECCertificateNum
	sw.FSTECCertificateDate = in.FSTECCertificateDate
	sw.FSTECProtectionClass = in.FSTECProtectionClass
	sw.FSTECValidUntil = in.FSTECValidUntil
	sw.FSBCertified = in.FSBCertified
	sw.FSBCertificateNum = in.FSBCertificateNum
	sw.FSBProtectionClass = in.FSBProtectionClass
	sw.Description = in.Description
	sw.Website = in.Website

	if err := s.repo.Update(ctx, sw); err != nil {
		return nil, err
	}

	return sw, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) ListCategories(ctx context.Context) ([]domain.SoftwareCategory, error) {
	return s.repo.ListCategories(ctx)
}

// FindRussianAlternatives находит российское ПО в заданной категории
func (s *service) FindRussianAlternatives(ctx context.Context, categoryID int16) ([]domain.Software, error) {
	isRussian := true
	return s.repo.List(ctx, repository.SoftwareFilter{
		CategoryID: &categoryID,
		IsRussian:  &isRussian,
		Limit:      50,
	})
}

// FindCertifiedSoftware находит сертифицированное ПО
func (s *service) FindCertifiedSoftware(ctx context.Context, categoryID int16, fstecOnly bool) ([]domain.Software, error) {
	return s.repo.List(ctx, repository.SoftwareFilter{
		CategoryID: &categoryID,
		FSTECOnly:  fstecOnly,
		Limit:      50,
	})
}

// SuggestAlternatives подбирает российские аналоги для заданного ПО
func (s *service) SuggestAlternatives(ctx context.Context, softwareID int64) ([]domain.Software, error) {
	if softwareID <= 0 {
		return nil, fmt.Errorf("invalid software id")
	}

	sw, err := s.repo.GetByID(ctx, softwareID)
	if err != nil {
		return nil, err
	}
	if sw == nil {
		return nil, fmt.Errorf("software not found")
	}

	isRussian := true
	filter := repository.SoftwareFilter{
		Limit:     20,
		IsRussian: &isRussian,
	}

	if sw.CategoryID != nil {
		categoryID := *sw.CategoryID
		filter.CategoryID = &categoryID
	}

	list, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	alternatives := filterOutSoftware(list, sw.ID)

	// Если в категории нет альтернатив, пробуем общий список российских решений
	if len(alternatives) == 0 {
		filter.CategoryID = nil
		list, err = s.repo.List(ctx, filter)
		if err != nil {
			return nil, err
		}
		alternatives = filterOutSoftware(list, sw.ID)
	}

	return alternatives, nil
}

// SuggestAlternativesForAsset подбирает аналоги для всех нероссийских программ на активе
func (s *service) SuggestAlternativesForAsset(ctx context.Context, assetID int64) ([]AssetSoftwareAlternative, error) {
	if assetID <= 0 {
		return nil, fmt.Errorf("invalid asset id")
	}

	installed, err := s.repo.ListAssetSoftware(ctx, assetID)
	if err != nil {
		return nil, err
	}

	if len(installed) == 0 {
		return nil, nil
	}

	var result []AssetSoftwareAlternative
	for _, item := range installed {
		if item.Software.IsRussian {
			continue
		}

		alternatives, err := s.SuggestAlternatives(ctx, item.Software.ID)
		if err != nil {
			return nil, err
		}
		if len(alternatives) == 0 {
			continue
		}

		result = append(result, AssetSoftwareAlternative{
			AssetSoftwareID: item.Link.ID,
			AssetID:         item.Link.AssetID,
			Version:         item.Link.Version,
			Notes:           item.Link.Notes,
			Software:        item.Software,
			Alternatives:    alternatives,
		})
	}

	return result, nil
}

func filterOutSoftware(list []domain.Software, excludeID int64) []domain.Software {
	result := make([]domain.Software, 0, len(list))
	for _, item := range list {
		if item.ID == excludeID {
			continue
		}
		result = append(result, item)
	}
	return result
}
