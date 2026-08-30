// Package compliance computes asset compliance against ИБ-стандарты
// (ФСТЭК-17, ISO 27001 …). Это покрывает блок «Оценка состояния
// защищённости» из 7.png диплома: для каждого актива считаем %
// соответствия каждому стандарту, какие требования закрыты внедрёнными
// контролями, какие нет.
//
// Алгоритм coverage(asset, requirement):
//   coverage = max(rc.coverage_weight) среди control_id, у которых есть
//   запись в asset_controls для этого актива. Если ни одного — 0.
//
// overall_score стандарта = avg(coverage по требованиям).
package compliance

import (
	"context"
	"fmt"

	"Diplom/internal/domain"
	"Diplom/internal/repository"
)

type Service interface {
	ListStandards(ctx context.Context) ([]domain.ComplianceStandard, error)
	AssetOverview(ctx context.Context, assetID int64) ([]domain.AssetComplianceOverview, error)
	AssetByStandard(ctx context.Context, assetID int64, standardCode string) (*domain.AssetStandardCompliance, error)
	AssetAllStandards(ctx context.Context, assetID int64) ([]*domain.AssetStandardCompliance, error)
}

type service struct {
	repo        repository.ComplianceRepository
	controlRepo repository.ControlRepository
	assetRepo   repository.AssetRepository
}

func NewService(repo repository.ComplianceRepository, controlRepo repository.ControlRepository, assetRepo repository.AssetRepository) Service {
	return &service{repo: repo, controlRepo: controlRepo, assetRepo: assetRepo}
}

func (s *service) ListStandards(ctx context.Context) ([]domain.ComplianceStandard, error) {
	return s.repo.ListStandards(ctx)
}

// assetOverviewCalc — внутренняя функция, считающая overview по одному стандарту.
// Принимает уже загруженные требования + ребра, чтобы не ходить за ними N раз.
func calcOverview(reqs []domain.ComplianceRequirement, links []domain.RequirementControlLink, attachedSet map[int64]bool) (overall float64, covered, partial, uncovered int, perReq map[int32]float64) {
	// Группируем links по requirement_id
	byReq := map[int32][]domain.RequirementControlLink{}
	for _, l := range links {
		byReq[l.RequirementID] = append(byReq[l.RequirementID], l)
	}

	perReq = make(map[int32]float64, len(reqs))
	if len(reqs) == 0 {
		return 0, 0, 0, 0, perReq
	}
	var sum float64
	for _, r := range reqs {
		var maxW float64
		for _, l := range byReq[r.ID] {
			if attachedSet[l.ControlID] && l.CoverageWeight > maxW {
				maxW = l.CoverageWeight
			}
		}
		perReq[r.ID] = maxW
		sum += maxW
		switch {
		case maxW <= 0:
			uncovered++
		case maxW >= 1.0:
			covered++
		default:
			partial++
		}
	}
	overall = sum / float64(len(reqs))
	return
}

func (s *service) AssetOverview(ctx context.Context, assetID int64) ([]domain.AssetComplianceOverview, error) {
	if assetID <= 0 {
		return nil, fmt.Errorf("invalid assetID")
	}

	asset, err := s.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset %d not found", assetID)
	}

	standards, err := s.repo.ListStandards(ctx)
	if err != nil {
		return nil, err
	}

	attached, err := s.controlRepo.ListAttached(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("list attached controls: %w", err)
	}
	attachedSet := make(map[int64]bool, len(attached))
	for _, c := range attached {
		attachedSet[c.ID] = true
	}

	out := make([]domain.AssetComplianceOverview, 0, len(standards))
	for _, st := range standards {
		reqs, err := s.repo.ListRequirements(ctx, st.ID)
		if err != nil {
			return nil, fmt.Errorf("list requirements for %s: %w", st.Code, err)
		}
		links, err := s.repo.ListRequirementControlLinks(ctx, st.ID)
		if err != nil {
			return nil, fmt.Errorf("list rc links for %s: %w", st.Code, err)
		}
		overall, covered, partial, uncov, _ := calcOverview(reqs, links, attachedSet)
		out = append(out, domain.AssetComplianceOverview{
			Standard:       st,
			OverallScore:   overall,
			CoveredCount:   covered,
			PartialCount:   partial,
			UncoveredCount: uncov,
			TotalCount:     len(reqs),
		})
	}
	return out, nil
}

func (s *service) AssetByStandard(ctx context.Context, assetID int64, standardCode string) (*domain.AssetStandardCompliance, error) {
	if assetID <= 0 {
		return nil, fmt.Errorf("invalid assetID")
	}
	if standardCode == "" {
		return nil, fmt.Errorf("standardCode required")
	}

	asset, err := s.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset %d not found", assetID)
	}

	st, err := s.repo.GetStandardByCode(ctx, standardCode)
	if err != nil {
		return nil, fmt.Errorf("standard %q not found: %w", standardCode, err)
	}

	reqs, err := s.repo.ListRequirements(ctx, st.ID)
	if err != nil {
		return nil, err
	}
	links, err := s.repo.ListRequirementControlLinks(ctx, st.ID)
	if err != nil {
		return nil, err
	}
	allControls, err := s.controlRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	controlByID := make(map[int64]domain.Control, len(allControls))
	for _, c := range allControls {
		controlByID[c.ID] = c
	}

	attached, err := s.controlRepo.ListAttached(ctx, assetID)
	if err != nil {
		return nil, err
	}
	attachedSet := make(map[int64]bool, len(attached))
	for _, c := range attached {
		attachedSet[c.ID] = true
	}

	overall, covered, partial, uncov, perReq := calcOverview(reqs, links, attachedSet)

	// Готовим RequirementStatus[] с разделением на covering / missing.
	byReq := map[int32][]domain.RequirementControlLink{}
	for _, l := range links {
		byReq[l.RequirementID] = append(byReq[l.RequirementID], l)
	}

	statuses := make([]domain.RequirementStatus, 0, len(reqs))
	for _, r := range reqs {
		rs := domain.RequirementStatus{
			Requirement: r,
			Coverage:    perReq[r.ID],
		}
		for _, l := range byReq[r.ID] {
			c, ok := controlByID[l.ControlID]
			if !ok {
				continue
			}
			if attachedSet[l.ControlID] {
				rs.CoveringControls = append(rs.CoveringControls, c)
			} else {
				rs.MissingControls = append(rs.MissingControls, c)
			}
		}
		statuses = append(statuses, rs)
	}

	return &domain.AssetStandardCompliance{
		Standard:       *st,
		OverallScore:   overall,
		CoveredCount:   covered,
		PartialCount:   partial,
		UncoveredCount: uncov,
		TotalCount:     len(reqs),
		Requirements:   statuses,
	}, nil
}

// AssetAllStandards — детализация по всем стандартам сразу. Используется
// генератором PDF-отчёта о состоянии защищённости.
func (s *service) AssetAllStandards(ctx context.Context, assetID int64) ([]*domain.AssetStandardCompliance, error) {
	standards, err := s.repo.ListStandards(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.AssetStandardCompliance, 0, len(standards))
	for _, st := range standards {
		c, err := s.AssetByStandard(ctx, assetID, st.Code)
		if err != nil {
			return nil, fmt.Errorf("asset compliance for %s: %w", st.Code, err)
		}
		out = append(out, c)
	}
	return out, nil
}
