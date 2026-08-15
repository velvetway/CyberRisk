package ptszi

import (
	"context"
	"fmt"
	"math"
	"sort"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

type Service interface {
	ListSources(ctx context.Context) ([]domain.ThreatSource, error)
	ListThreats(ctx context.Context) ([]domain.PTSZIThreat, error)
	ListVulnerableLinks(ctx context.Context) ([]domain.PTSZIVulnerableLink, error)
	ListControls(ctx context.Context) ([]domain.PTSZIControl, error)
	ListDestructiveActions(ctx context.Context) ([]domain.DestructiveAction, error)
	ListUBI(ctx context.Context, limit, offset int32, query string) ([]domain.PTSZIUBIThreat, error)
	AssetProfile(ctx context.Context, assetID int64) (*domain.PTSZIAssetProfile, error)
	UpdateAssetVulnerableLinks(ctx context.Context, assetID int64, ids []int16) error
	UpdateAssetControls(ctx context.Context, assetID int64, controls []AssetControlInput) error
	ApplicableThreats(ctx context.Context, assetID int64) ([]domain.PTSZIAttackPath, error)
	AttackPath(ctx context.Context, assetID, threatID int64) (*domain.PTSZIAttackPath, error)
}

type AssetControlInput struct {
	ControlID     int16   `json:"control_id"`
	Effectiveness float64 `json:"effectiveness"`
}

type service struct {
	assets repository.AssetRepository
	repo   repository.PTSZIRepository
}

func NewService(assets repository.AssetRepository, repo repository.PTSZIRepository) Service {
	return &service{assets: assets, repo: repo}
}

func (s *service) ListSources(ctx context.Context) ([]domain.ThreatSource, error) {
	return s.repo.ListSources(ctx)
}

func (s *service) ListThreats(ctx context.Context) ([]domain.PTSZIThreat, error) {
	return s.repo.ListThreats(ctx)
}

func (s *service) ListVulnerableLinks(ctx context.Context) ([]domain.PTSZIVulnerableLink, error) {
	return s.repo.ListVulnerableLinks(ctx)
}

func (s *service) ListControls(ctx context.Context) ([]domain.PTSZIControl, error) {
	return s.repo.ListControls(ctx)
}

func (s *service) ListDestructiveActions(ctx context.Context) ([]domain.DestructiveAction, error) {
	return s.repo.ListDestructiveActions(ctx)
}

func (s *service) ListUBI(ctx context.Context, limit, offset int32, query string) ([]domain.PTSZIUBIThreat, error) {
	return s.repo.ListUBI(ctx, limit, offset, query)
}

func (s *service) AssetProfile(ctx context.Context, assetID int64) (*domain.PTSZIAssetProfile, error) {
	asset, err := s.loadAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}
	vls, err := s.repo.ListAssetVulnerableLinks(ctx, assetID)
	if err != nil {
		return nil, err
	}
	controls, err := s.repo.ListAssetControls(ctx, assetID)
	if err != nil {
		return nil, err
	}
	threats, err := s.ApplicableThreats(ctx, assetID)
	if err != nil {
		return nil, err
	}
	contour, err := s.repo.AssetContour(ctx, assetID)
	if err != nil {
		return nil, err
	}
	return &domain.PTSZIAssetProfile{
		Asset:             *asset,
		SecurityContour:   contour,
		VulnerableLinks:   vls,
		Controls:          controls,
		ApplicableThreats: threats,
	}, nil
}

func (s *service) UpdateAssetVulnerableLinks(ctx context.Context, assetID int64, ids []int16) error {
	if _, err := s.loadAsset(ctx, assetID); err != nil {
		return err
	}
	return s.repo.SaveAssetVulnerableLinks(ctx, assetID, ids)
}

func (s *service) UpdateAssetControls(ctx context.Context, assetID int64, controls []AssetControlInput) error {
	if _, err := s.loadAsset(ctx, assetID); err != nil {
		return err
	}
	in := make([]repository.AssetPTSZIControlInput, 0, len(controls))
	for _, c := range controls {
		in = append(in, repository.AssetPTSZIControlInput{
			ControlID:     c.ControlID,
			Effectiveness: c.Effectiveness,
		})
	}
	return s.repo.SaveAssetControls(ctx, assetID, in)
}

func (s *service) ApplicableThreats(ctx context.Context, assetID int64) ([]domain.PTSZIAttackPath, error) {
	asset, err := s.loadAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}
	threats, err := s.repo.ListThreats(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.PTSZIAttackPath, 0, len(threats))
	for _, threat := range threats {
		path, err := s.repo.LoadAttackPath(ctx, asset, threat.ID)
		if err != nil {
			return nil, err
		}
		if path == nil || !path.Applicable {
			continue
		}
		s.calculate(path)
		out = append(out, *path)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].W > out[j].W
	})
	return out, nil
}

func (s *service) AttackPath(ctx context.Context, assetID, threatID int64) (*domain.PTSZIAttackPath, error) {
	asset, err := s.loadAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}
	path, err := s.repo.LoadAttackPath(ctx, asset, threatID)
	if err != nil {
		return nil, err
	}
	if path == nil {
		return nil, fmt.Errorf("ptszi threat not found")
	}
	s.calculate(path)
	return path, nil
}

func (s *service) loadAsset(ctx context.Context, assetID int64) (*domain.Asset, error) {
	if assetID <= 0 {
		return nil, fmt.Errorf("asset id must be positive")
	}
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}
	return asset, nil
}

func (s *service) calculate(path *domain.PTSZIAttackPath) {
	path.QReaction = calculateQReaction(path.VulnerableLinks)
	path.W = CalculateW(path.QThreat, path.QSeverity, path.QReaction, path.Z)
	path.Level = LevelFromW(path.W)
	path.MissingControls = missingControls(path.VulnerableLinks)
	path.Recommendations = recommendationsFromMissing(path.MissingControls)
}

func CalculateW(qThreat, qSeverity, qReaction, z float64) float64 {
	qThreat = clamp01(qThreat)
	qSeverity = clamp01(qSeverity)
	qReaction = clamp01(qReaction)
	if z < 0.5 {
		z = 0.5
	}
	if z > 1 {
		z = 1
	}
	return (qThreat + qSeverity + (1 - qReaction)) / 3 * z
}

func LevelFromW(w float64) string {
	switch {
	case w >= 0.75:
		return "critical"
	case w >= 0.50:
		return "high"
	case w >= 0.25:
		return "medium"
	default:
		return "low"
	}
}

func calculateQReaction(vls []domain.PTSZIPathVL) float64 {
	if len(vls) == 0 {
		return 0
	}
	sum := 0.0
	for i := range vls {
		product := 1.0
		for j := range vls[i].Controls {
			c := &vls[i].Controls[j]
			if !c.Implemented {
				continue
			}
			c.ResultingCoverage = clamp01(c.Coverage * c.Effectiveness)
			product *= 1 - c.ResultingCoverage
		}
		vls[i].Coverage = clamp01(1 - product)
		vls[i].Uncovered = vls[i].Coverage == 0
		sum += vls[i].Coverage
	}
	return clamp01(sum / float64(len(vls)))
}

func missingControls(vls []domain.PTSZIPathVL) []domain.PTSZIControl {
	seen := map[int16]bool{}
	out := make([]domain.PTSZIControl, 0)
	for _, vl := range vls {
		for _, c := range vl.Controls {
			if c.Implemented || seen[c.Control.ID] {
				continue
			}
			seen[c.Control.ID] = true
			out = append(out, c.Control)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Code < out[j].Code
	})
	return out
}

func recommendationsFromMissing(controls []domain.PTSZIControl) []domain.PTSZIRecommendation {
	out := make([]domain.PTSZIRecommendation, 0, len(controls))
	for _, c := range controls {
		out = append(out, domain.PTSZIRecommendation{
			ControlID:   c.ID,
			ControlCode: c.Code,
			Category:    recommendationCategory(c.Code),
			Title:       "Внедрить метод " + c.Code + ": " + c.Name,
			Description: "Метод закрывает актуальные уязвимые звенья сценария и повышает Qreaction.",
			Priority:    recommendationPriority(c.Code),
		})
	}
	return out
}

func recommendationCategory(code string) string {
	switch code {
	case "FW", "IDS", "HP", "DZ", "DD":
		return "Защита ЛВС"
	case "A", "AD":
		return "Защита АРМ"
	case "L", "TE", "DS":
		return "Защита конфиденциальной информации"
	case "R":
		return "Защита информации"
	default:
		return "Организационно-технические меры"
	}
}

func recommendationPriority(code string) string {
	switch code {
	case "FW", "IDS", "L", "DD":
		return "high"
	case "A", "AD", "R":
		return "medium"
	default:
		return "low"
	}
}

func clamp01(v float64) float64 {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
