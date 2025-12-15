// internal/service/risk/service.go
package risk

import (
	"context"
	"fmt"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

// Service описывает публичные методы сервиса риска.
type Service interface {
	// Быстрый расчёт для пары "актив + угроза" + рекомендации.
	PreviewRisk(ctx context.Context, assetID, threatID int64) (*RiskResult, []RiskRecommendation, error)
	// Набор точек для карты рисков по всем активам и угрозам.
	Overview(ctx context.Context) ([]OverviewPoint, error)
	// Автоматический расчёт всех рисков для конкретного актива.
	AssetRiskProfile(ctx context.Context, assetID int64) ([]AssetRisk, error)
}

// OverviewPoint — точка на глобальной карте рисков.
type OverviewPoint struct {
	AssetID    int64  `json:"asset_id"`
	AssetName  string `json:"asset_name"`
	ThreatID   int64  `json:"threat_id"`
	ThreatName string `json:"threat_name"`

	Impact     int16  `json:"impact"`
	Likelihood int16  `json:"likelihood"`
	Score      int16  `json:"score"`
	Level      string `json:"level"`
}

// AssetRisk — риск для актива от конкретной угрозы с рекомендациями.
type AssetRisk struct {
	ThreatID          int64                `json:"threat_id"`
	ThreatName        string               `json:"threat_name"`
	ThreatDescription string               `json:"threat_description"`
	ThreatType        string               `json:"threat_type"`
	Impact            int16                `json:"impact"`
	Likelihood        int16                `json:"likelihood"`
	Score             int16                `json:"score"`
	Level             string               `json:"level"`
	Recommendations   []RiskRecommendation `json:"recommendations"`
}

// service — внутренняя реализация Service.
type service struct {
	assetsRepo     repository.AssetRepository
	threatsRepo    repository.ThreatRepository
	vulnsRepo      repository.VulnerabilityRepository
	assetVulnsRepo repository.AssetVulnerabilityRepository

	calculator *Calculator // тип и логика определены в calculator.go
}

// NewService конструирует экземпляр сервиса риска.
func NewService(
	assets repository.AssetRepository,
	threats repository.ThreatRepository,
	vulns repository.VulnerabilityRepository,
	assetVulns repository.AssetVulnerabilityRepository,
) Service {
	return &service{
		assetsRepo:     assets,
		threatsRepo:    threats,
		vulnsRepo:      vulns,
		assetVulnsRepo: assetVulns,
		calculator:     NewCalculator(),
	}
}

// PreviewRisk — расчёт для конкретного актива и угрозы + список рекомендаций.
func (s *service) PreviewRisk(ctx context.Context, assetID, threatID int64) (*RiskResult, []RiskRecommendation, error) {
	if assetID <= 0 || threatID <= 0 {
		return nil, nil, fmt.Errorf("assetID and threatID must be positive")
	}

	asset, err := s.assetsRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, nil, fmt.Errorf("asset not found")
	}

	threat, err := s.threatsRepo.GetByID(ctx, threatID)
	if err != nil {
		return nil, nil, fmt.Errorf("get threat: %w", err)
	}
	if threat == nil {
		return nil, nil, fmt.Errorf("threat not found")
	}

	vulns, err := s.vulnsForAsset(ctx, assetID)
	if err != nil {
		return nil, nil, err
	}

	res := s.calculator.Calculate(asset, threat, vulns)
	recs := GenerateRecommendations(asset, threat, vulns, res)

	return &res, recs, nil
}

// Overview — строит точки для карты рисков по всем активам и угрозам.
func (s *service) Overview(ctx context.Context) ([]OverviewPoint, error) {
	assets, err := s.assetsRepo.List(ctx, repository.AssetFilter{})
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}

	threats, err := s.threatsRepo.List(ctx, repository.ThreatFilter{})
	if err != nil {
		return nil, fmt.Errorf("list threats: %w", err)
	}

	// заранее кешируем уязвимости по активам, чтобы не дергать БД в цикле NxM
	vulnsByAsset := make(map[int64][]domain.Vulnerability, len(assets))
	for _, a := range assets {
		vv, err := s.vulnsForAsset(ctx, a.ID)
		if err != nil {
			return nil, err
		}
		vulnsByAsset[a.ID] = vv
	}

	var result []OverviewPoint

	for _, a := range assets {
		for _, t := range threats {
			vulns := vulnsByAsset[a.ID]
			r := s.calculator.Calculate(&a, &t, vulns)

			result = append(result, OverviewPoint{
				AssetID:    a.ID,
				AssetName:  a.Name,
				ThreatID:   t.ID,
				ThreatName: t.Name,
				Impact:     r.Impact,
				Likelihood: r.Likelihood,
				Score:      r.Score,
				Level:      r.Level,
			})
		}
	}

	return result, nil
}

// AssetRiskProfile — автоматически рассчитывает риски для актива от всех угроз.
func (s *service) AssetRiskProfile(ctx context.Context, assetID int64) ([]AssetRisk, error) {
	if assetID <= 0 {
		return nil, fmt.Errorf("assetID must be positive")
	}

	// Проверяем существование актива
	asset, err := s.assetsRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}

	// Получаем все угрозы
	threats, err := s.threatsRepo.List(ctx, repository.ThreatFilter{})
	if err != nil {
		return nil, fmt.Errorf("list threats: %w", err)
	}

	// Получаем уязвимости для актива один раз
	vulns, err := s.vulnsForAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}

	var results []AssetRisk

	// Рассчитываем риск для каждой угрозы
	for _, threat := range threats {
		riskResult := s.calculator.Calculate(asset, &threat, vulns)
		recommendations := GenerateRecommendations(asset, &threat, vulns, riskResult)

		var desc string
		if threat.Description != nil {
			desc = *threat.Description
		}

		results = append(results, AssetRisk{
			ThreatID:          threat.ID,
			ThreatName:        threat.Name,
			ThreatDescription: desc,
			ThreatType:        string(threat.SourceType),
			Impact:            riskResult.Impact,
			Likelihood:        riskResult.Likelihood,
			Score:             riskResult.Score,
			Level:             riskResult.Level,
			Recommendations:   recommendations,
		})
	}

	return results, nil
}

// vulnsForAsset — вспомогательный метод: уязвимости, привязанные к активу.
func (s *service) vulnsForAsset(ctx context.Context, assetID int64) ([]domain.Vulnerability, error) {
	links, err := s.assetVulnsRepo.ListByAsset(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("list asset vulnerabilities: %w", err)
	}

	if len(links) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(links))
	for _, link := range links {
		ids = append(ids, link.VulnerabilityID)
	}

	allVulns, err := s.vulnsRepo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list vulnerabilities: %w", err)
	}

	vulnByID := make(map[int64]domain.Vulnerability, len(allVulns))
	for _, v := range allVulns {
		vulnByID[v.ID] = v
	}

	var vulns []domain.Vulnerability
	for _, link := range links {
		if v, ok := vulnByID[link.VulnerabilityID]; ok {
			vulns = append(vulns, v)
		}
	}
	return vulns, nil
}
