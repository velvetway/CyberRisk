// internal/service/risk/service.go
package risk

import (
	"context"
	"fmt"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

// Service describes the public methods of the PTSZI risk service.
//
// All risk numbers are computed strictly via the W formula
// (see docs/risk-model.md). There is no legacy 1..25 / impact×likelihood
// engine.
type Service interface {
	// Per-pair PTSZI attack path: full S → ST → VL → DA chain plus W decomposition.
	AssembleAttackPath(ctx context.Context, assetID, threatID int64) (*domain.AttackPath, error)
	// Bulk: AttackPath for every threat applicable to an asset, plus the asset-level aggregate.
	AssembleAssetAttackPaths(ctx context.Context, assetID int64) (*domain.AssetAttackPathsResponse, error)
	// W for every {asset × threat} pair — feeds the risk overview map.
	Overview(ctx context.Context) ([]OverviewPoint, error)

	// Reference dictionaries used by the PTSZI graph UI.
	ListThreatSources(ctx context.Context) ([]domain.ThreatSource, error)
	ListDestructiveActions(ctx context.Context) ([]domain.DestructiveAction, error)
}

// OverviewPoint — one row of the global PTSZI risk map.
type OverviewPoint struct {
	AssetID    int64  `json:"asset_id"`
	AssetName  string `json:"asset_name"`
	ThreatID   int64  `json:"threat_id"`
	ThreatName string `json:"threat_name"`

	// PTSZI W-model
	W         float64 `json:"w"`
	Level     string  `json:"level"`
	QThreat   float64 `json:"q_threat"`
	QSeverity float64 `json:"q_severity"`
	QReaction float64 `json:"q_reaction"`
	Z         float64 `json:"z"`
}

type service struct {
	assetsRepo  repository.AssetRepository
	threatsRepo repository.ThreatRepository
	sourceRepo  repository.ThreatSourceRepository
	daRepo      repository.DestructiveActionRepository
	graphRepo   repository.RiskGraphRepository
}

// NewService wires the PTSZI risk service.
func NewService(
	assets repository.AssetRepository,
	threats repository.ThreatRepository,
	sources repository.ThreatSourceRepository,
	das repository.DestructiveActionRepository,
	graph repository.RiskGraphRepository,
) Service {
	return &service{
		assetsRepo:  assets,
		threatsRepo: threats,
		sourceRepo:  sources,
		daRepo:      das,
		graphRepo:   graph,
	}
}

// Overview builds one OverviewPoint per (asset, threat) pair, filtered
// down to applicable pairs only (см. IsApplicable). Без фильтра матрица
// ~9×227 = 2000+ строк забивалась бы «угрозами для СУБД на рабочей
// станции» — формально риск ноль, но шум огромный.
func (s *service) Overview(ctx context.Context) ([]OverviewPoint, error) {
	assets, err := s.assetsRepo.List(ctx, repository.AssetFilter{})
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}

	threats, err := s.threatsRepo.List(ctx, repository.ThreatFilter{Limit: 500})
	if err != nil {
		return nil, fmt.Errorf("list threats: %w", err)
	}

	result := make([]OverviewPoint, 0, len(assets)*len(threats))
	for _, a := range assets {
		for _, t := range threats {
			if !IsApplicable(a, t) {
				continue
			}
			path, err := s.AssembleAttackPath(ctx, a.ID, t.ID)
			if err != nil {
				continue
			}
			result = append(result, OverviewPoint{
				AssetID:    a.ID,
				AssetName:  a.Name,
				ThreatID:   t.ID,
				ThreatName: t.Name,

				W:         path.W,
				Level:     path.Level,
				QThreat:   path.QThreat,
				QSeverity: path.QSeverity,
				QReaction: path.QReaction,
				Z:         path.Z,
			})
		}
	}
	return result, nil
}

// AssembleAttackPath builds the full S → ST → VL → DA chain for one (asset, threat)
// pair and computes W per the PTSZI formula.
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

// AssembleAssetAttackPaths returns AttackPath for every threat for one asset,
// drops fully empty paths (no S, no VL, no DA), and computes the asset aggregate.
func (s *service) AssembleAssetAttackPaths(ctx context.Context, assetID int64) (*domain.AssetAttackPathsResponse, error) {
	if assetID <= 0 {
		return nil, fmt.Errorf("assetID must be positive")
	}

	asset, err := s.assetsRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}

	threats, err := s.threatsRepo.List(ctx, repository.ThreatFilter{Limit: 500})
	if err != nil {
		return nil, fmt.Errorf("list threats: %w", err)
	}

	paths := make([]domain.AttackPath, 0, len(threats))
	for _, t := range threats {
		if !IsApplicable(*asset, t) {
			continue
		}
		path, err := s.AssembleAttackPath(ctx, assetID, t.ID)
		if err != nil {
			continue
		}
		if len(path.Sources) == 0 && len(path.VulnerableLinks) == 0 && len(path.DestructiveActions) == 0 {
			continue
		}
		paths = append(paths, *path)
	}

	return &domain.AssetAttackPathsResponse{
		Asset:     domain.AssetRef{ID: asset.ID, Name: asset.Name},
		Aggregate: ComputeAssetAggregate(paths),
		Paths:     paths,
	}, nil
}

func (s *service) ListThreatSources(ctx context.Context) ([]domain.ThreatSource, error) {
	return s.sourceRepo.List(ctx)
}

func (s *service) ListDestructiveActions(ctx context.Context) ([]domain.DestructiveAction, error) {
	return s.daRepo.List(ctx)
}
