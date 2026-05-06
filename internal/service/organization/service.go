// Package organization — сводные оценки на уровне организации (всех активов).
//
// Это надстройка над per-asset сервисами:
//   • risk.AssembleAssetAttackPaths — для каждого актива получаем W_max и угрозы
//   • compliance.AssetOverview      — для каждого актива получаем %-соответствие
//   • control.ListAttached          — кол-во внедрённых контролей
//
// Основные методы:
//   Overview()       — сводные метрики (cards для дашборда)
//   AssetMatrix()    — табличный список активов с показателями
//   CriticalRisks(N) — топ-N (asset × threat) с наибольшим W
package organization

import (
	"context"
	"sort"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
	"Diplom/internal/service/compliance"
	"Diplom/internal/service/risk"
)

type Service interface {
	Overview(ctx context.Context) (*domain.OrganizationOverview, error)
	AssetMatrix(ctx context.Context) ([]domain.AssetMatrixRow, error)
	CriticalRisks(ctx context.Context, limit int) ([]domain.CriticalRisk, error)
}

type service struct {
	assetRepo     repository.AssetRepository
	controlRepo   repository.ControlRepository
	riskSvc       risk.Service
	complianceSvc compliance.Service

	assetTypeName func(ctx context.Context, id int16) string
}

func NewService(
	assetRepo repository.AssetRepository,
	controlRepo repository.ControlRepository,
	riskSvc risk.Service,
	complianceSvc compliance.Service,
	assetTypeName func(ctx context.Context, id int16) string,
) Service {
	return &service{
		assetRepo:     assetRepo,
		controlRepo:   controlRepo,
		riskSvc:       riskSvc,
		complianceSvc: complianceSvc,
		assetTypeName: assetTypeName,
	}
}

// loadAssets — все активы (без фильтрации). Дальнейший расчёт идёт по этому набору.
func (s *service) loadAssets(ctx context.Context) ([]domain.Asset, error) {
	return s.assetRepo.List(ctx, repository.AssetFilter{Limit: 5000})
}

// perAssetSnapshot — кэш per-asset данных, чтобы Overview / AssetMatrix /
// CriticalRisks не пересчитывали одно и то же.
type perAssetSnapshot struct {
	asset       domain.Asset
	typeName    string
	paths       *domain.AssetAttackPathsResponse
	complByStd  []domain.AssetComplianceOverview
	controlCnt  int
}

func (s *service) snapshot(ctx context.Context) ([]perAssetSnapshot, error) {
	assets, err := s.loadAssets(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]perAssetSnapshot, 0, len(assets))
	for _, a := range assets {
		snap := perAssetSnapshot{asset: a}
		if a.AssetTypeID != nil && s.assetTypeName != nil {
			snap.typeName = s.assetTypeName(ctx, *a.AssetTypeID)
		}
		// non-fatal на per-asset ошибки: просто оставим нули, чтобы один
		// сломанный актив не валил всю сводку
		if paths, err := s.riskSvc.AssembleAssetAttackPaths(ctx, a.ID); err == nil {
			snap.paths = paths
		}
		if compl, err := s.complianceSvc.AssetOverview(ctx, a.ID); err == nil {
			snap.complByStd = compl
		}
		if ctrls, err := s.controlRepo.ListAttached(ctx, a.ID); err == nil {
			snap.controlCnt = len(ctrls)
		}
		out = append(out, snap)
	}
	return out, nil
}

func (s *service) Overview(ctx context.Context) (*domain.OrganizationOverview, error) {
	snaps, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}

	o := &domain.OrganizationOverview{
		AssetsByEnvironment: map[string]int{},
		RiskDistribution:    map[string]int{"critical": 0, "high": 0, "medium": 0, "low": 0},
	}

	typeBuckets := map[string]*domain.AssetTypeBucket{}
	stdAggregator := map[string]*stdAcc{} // ключ = стандарт.code

	var (
		wMaxOverall    float64
		wMaxAssetName  string
		wMaxThreatName string
		wSumPerAsset   float64
		assetsWithRisk int
	)

	for _, sp := range snaps {
		o.TotalAssets++
		if sp.asset.IsIsolated {
			o.IsolatedAssets++
		}
		env := string(sp.asset.Environment)
		if env == "" {
			env = "unknown"
		}
		o.AssetsByEnvironment[env]++

		typeKey := sp.typeName
		if typeKey == "" {
			typeKey = "(не задан)"
		}
		bucket, ok := typeBuckets[typeKey]
		if !ok {
			bucket = &domain.AssetTypeBucket{TypeName: typeKey}
			if sp.asset.AssetTypeID != nil {
				v := *sp.asset.AssetTypeID
				bucket.TypeID = &v
			}
			typeBuckets[typeKey] = bucket
		}
		bucket.Count++

		o.TotalControls += sp.controlCnt
		if sp.paths != nil {
			lvl := sp.paths.Aggregate.Level
			if lvl == "" {
				lvl = "low"
			}
			o.RiskDistribution[lvl]++

			if sp.paths.Aggregate.WMax > wMaxOverall {
				wMaxOverall = sp.paths.Aggregate.WMax
				wMaxAssetName = sp.asset.Name
				// найдём имя самой опасной угрозы
				for _, p := range sp.paths.Paths {
					if p.W == sp.paths.Aggregate.WMax {
						wMaxThreatName = p.Threat.Name
						break
					}
				}
			}
			wSumPerAsset += sp.paths.Aggregate.WMax
			o.UncoveredVLs += sp.paths.Aggregate.UncoveredCount
			assetsWithRisk++
		} else {
			o.RiskDistribution["low"]++
		}

		for _, comp := range sp.complByStd {
			a, ok := stdAggregator[comp.Standard.Code]
			if !ok {
				a = &stdAcc{std: comp.Standard, min: 1.0, max: 0.0}
				stdAggregator[comp.Standard.Code] = a
			}
			a.sum += comp.OverallScore
			a.count++
			if comp.OverallScore < a.min {
				a.min = comp.OverallScore
			}
			if comp.OverallScore > a.max {
				a.max = comp.OverallScore
			}
		}
	}

	// плоский список ассет-типов, отсортированный по убыванию count
	for _, b := range typeBuckets {
		o.AssetsByType = append(o.AssetsByType, *b)
	}
	sort.Slice(o.AssetsByType, func(i, j int) bool {
		return o.AssetsByType[i].Count > o.AssetsByType[j].Count
	})

	for _, a := range stdAggregator {
		summary := domain.OrganizationComplianceSummary{
			Standard:    a.std,
			AvgScore:    a.sum / float64(maxInt(a.count, 1)),
			MinScore:    a.min,
			MaxScore:    a.max,
			AssetsCount: a.count,
		}
		o.ComplianceByStd = append(o.ComplianceByStd, summary)
	}
	sort.Slice(o.ComplianceByStd, func(i, j int) bool {
		return o.ComplianceByStd[i].Standard.SortOrder < o.ComplianceByStd[j].Standard.SortOrder
	})

	o.WMax = wMaxOverall
	o.WMaxAsset = wMaxAssetName
	o.WMaxThreat = wMaxThreatName
	if assetsWithRisk > 0 {
		o.AvgWPerAsset = wSumPerAsset / float64(assetsWithRisk)
	}

	return o, nil
}

func (s *service) AssetMatrix(ctx context.Context) ([]domain.AssetMatrixRow, error) {
	snaps, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.AssetMatrixRow, 0, len(snaps))
	for _, sp := range snaps {
		row := domain.AssetMatrixRow{
			AssetID:         sp.asset.ID,
			Name:            sp.asset.Name,
			TypeName:        sp.typeName,
			Environment:     string(sp.asset.Environment),
			IsIsolated:      sp.asset.IsIsolated,
			ControlCount:    sp.controlCnt,
			ComplianceByStd: sp.complByStd,
		}
		if sp.paths != nil {
			row.WMax = sp.paths.Aggregate.WMax
			row.Level = sp.paths.Aggregate.Level
			row.ThreatCount = sp.paths.Aggregate.ThreatCount
		} else {
			row.Level = "low"
		}
		out = append(out, row)
	}
	// сортируем по W desc — самые опасные сверху
	sort.Slice(out, func(i, j int) bool {
		if out[i].WMax != out[j].WMax {
			return out[i].WMax > out[j].WMax
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func (s *service) CriticalRisks(ctx context.Context, limit int) ([]domain.CriticalRisk, error) {
	if limit <= 0 {
		limit = 20
	}
	snaps, err := s.snapshot(ctx)
	if err != nil {
		return nil, err
	}
	all := make([]domain.CriticalRisk, 0, 64)
	for _, sp := range snaps {
		if sp.paths == nil {
			continue
		}
		for _, p := range sp.paths.Paths {
			all = append(all, domain.CriticalRisk{
				AssetID:    sp.asset.ID,
				AssetName:  sp.asset.Name,
				ThreatID:   p.Threat.ID,
				ThreatName: p.Threat.Name,
				BDUID:      p.Threat.BDUID,
				W:          p.W,
				Level:      p.Level,
			})
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].W > all[j].W })
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

type stdAcc struct {
	std   domain.ComplianceStandard
	sum   float64
	count int
	min   float64
	max   float64
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

