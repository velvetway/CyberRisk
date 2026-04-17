// internal/service/risk/service.go
package risk

import (
	"context"
	"fmt"
	"math"

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
	// Новый PTSZI граф атаки для пары актив+угроза с формулой W_i.
	AssembleAttackPath(ctx context.Context, assetID, threatID int64) (*domain.AttackPath, error)
	// Справочники для PTSZI-графа
	ListThreatSources(ctx context.Context) ([]domain.ThreatSource, error)
	ListDestructiveActions(ctx context.Context) ([]domain.DestructiveAction, error)
}

// OverviewPoint — точка на глобальной карте рисков.
type OverviewPoint struct {
	AssetID    int64  `json:"asset_id"`
	AssetName  string `json:"asset_name"`
	ThreatID   int64  `json:"threat_id"`
	ThreatName string `json:"threat_name"`

	// Legacy back-compat display (derived from W below)
	Impact     int16  `json:"impact"`
	Likelihood int16  `json:"likelihood"`
	Score      int16  `json:"score"`
	Level      string `json:"level"`

	// PTSZI model — W in [0,1] plus its components
	W         float64 `json:"w"`
	QThreat   float64 `json:"q_threat"`
	QSeverity float64 `json:"q_severity"`
	QReaction float64 `json:"q_reaction"`
	Z         float64 `json:"z"`
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
	sourceRepo     repository.ThreatSourceRepository
	daRepo         repository.DestructiveActionRepository
	graphRepo      repository.RiskGraphRepository

	calculator *Calculator // тип и логика определены в calculator.go
}

// NewService конструирует экземпляр сервиса риска.
func NewService(
	assets repository.AssetRepository,
	threats repository.ThreatRepository,
	vulns repository.VulnerabilityRepository,
	assetVulns repository.AssetVulnerabilityRepository,
	sources repository.ThreatSourceRepository,
	das repository.DestructiveActionRepository,
	graph repository.RiskGraphRepository,
) Service {
	return &service{
		assetsRepo:     assets,
		threatsRepo:    threats,
		vulnsRepo:      vulns,
		assetVulnsRepo: assetVulns,
		sourceRepo:     sources,
		daRepo:         das,
		graphRepo:      graph,
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

	result := make([]OverviewPoint, 0, len(assets)*len(threats))

	for _, a := range assets {
		for _, t := range threats {
			path, err := s.AssembleAttackPath(ctx, a.ID, t.ID)
			if err != nil {
				// Если один путь не собрался — не валим весь overview, пропускаем пару.
				continue
			}
			result = append(result, OverviewPoint{
				AssetID:    a.ID,
				AssetName:  a.Name,
				ThreatID:   t.ID,
				ThreatName: t.Name,

				// Back-compat display: W scaled to legacy 1..25 / 1..5 ranges
				Impact:     int16(math.Round(path.QSeverity * 5)),
				Likelihood: int16(math.Round(path.QThreat * 5)),
				Score:      int16(math.Round(path.W * 25)),
				Level:      path.Level,

				W:         path.W,
				QThreat:   path.QThreat,
				QSeverity: path.QSeverity,
				QReaction: path.QReaction,
				Z:         path.Z,
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

// AssembleAttackPath — строит полную цепочку S → ST → VL → DA для пары
// (актив, угроза) и считает W_i по формуле ПТСЗИ.
func (s *service) AssembleAttackPath(ctx context.Context, assetID, threatID int64) (*domain.AttackPath, error) {
	if assetID <= 0 || threatID <= 0 {
		return nil, fmt.Errorf("assetID and threatID must be positive")
	}

	asset, err := s.assetsRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}

	threat, err := s.threatsRepo.GetByID(ctx, threatID)
	if err != nil {
		return nil, fmt.Errorf("get threat: %w", err)
	}
	if threat == nil {
		return nil, fmt.Errorf("threat not found")
	}

	sources, err := s.sourceRepo.ForThreat(ctx, threatID)
	if err != nil {
		return nil, fmt.Errorf("load sources: %w", err)
	}
	das, err := s.daRepo.ForThreat(ctx, threatID)
	if err != nil {
		return nil, fmt.Errorf("load destructive actions: %w", err)
	}
	vls, err := s.graphRepo.LoadVulnerableLinks(ctx, assetID, threatID)
	if err != nil {
		return nil, fmt.Errorf("load vulnerable links: %w", err)
	}

	qR := QReactionFromVLs(vls)
	z := ZFromAsset(*asset)
	w := CalculateW(threat.QThreat, threat.QSeverity, qR, z)

	bduID := ""
	if threat.BDUID != nil {
		bduID = *threat.BDUID
	}

	return &domain.AttackPath{
		Asset:              domain.AssetRef{ID: asset.ID, Name: asset.Name},
		Threat:             domain.ThreatRef{ID: threat.ID, Name: threat.Name, BDUID: bduID},
		Sources:            sources,
		VulnerableLinks:    vls,
		DestructiveActions: das,
		QThreat:            threat.QThreat,
		QSeverity:          threat.QSeverity,
		QReaction:          qR,
		Z:                  z,
		W:                  w,
		Level:              LevelFromW(w),
	}, nil
}

// ListThreatSources — справочник всех источников угроз (S1..Sn).
func (s *service) ListThreatSources(ctx context.Context) ([]domain.ThreatSource, error) {
	return s.sourceRepo.List(ctx)
}

// ListDestructiveActions — справочник всех деструктивных действий (DA1..DAn).
func (s *service) ListDestructiveActions(ctx context.Context) ([]domain.DestructiveAction, error) {
	return s.daRepo.List(ctx)
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
