package compliance

import (
	"testing"

	"Diplom/internal/domain"
)

// Helper: 3 требования, 5 рёбер для проверки 3 уровней покрытия.
//
//   req1 закрывается полностью (control 100, weight 1.0)
//   req2 закрывается частично (control 200, weight 0.5)
//   req3 имеет ребро на control 300 weight 1.0, но 300 не внедрён
//
// При attached={100, 200} ожидаем: 1 covered, 1 partial, 1 uncovered.
func sampleReqs() []domain.ComplianceRequirement {
	return []domain.ComplianceRequirement{
		{ID: 1, Code: "R1"},
		{ID: 2, Code: "R2"},
		{ID: 3, Code: "R3"},
	}
}

func sampleLinks() []domain.RequirementControlLink {
	return []domain.RequirementControlLink{
		{RequirementID: 1, ControlID: 100, CoverageWeight: 1.0},
		{RequirementID: 2, ControlID: 200, CoverageWeight: 0.5},
		// Для req3 — два ребра, оба ведут на не-внедрённые контроли.
		{RequirementID: 3, ControlID: 300, CoverageWeight: 1.0},
		{RequirementID: 3, ControlID: 400, CoverageWeight: 0.7},
	}
}

func TestCalcOverview_FullSpread(t *testing.T) {
	attached := map[int64]bool{100: true, 200: true}

	overall, covered, partial, uncov, perReq := calcOverview(sampleReqs(), sampleLinks(), attached)

	if covered != 1 || partial != 1 || uncov != 1 {
		t.Errorf("counts: covered=%d partial=%d uncov=%d, want 1/1/1", covered, partial, uncov)
	}

	expectedOverall := (1.0 + 0.5 + 0.0) / 3
	if absDiff(overall, expectedOverall) > 1e-9 {
		t.Errorf("overall = %v, want %v", overall, expectedOverall)
	}

	if perReq[1] != 1.0 || perReq[2] != 0.5 || perReq[3] != 0.0 {
		t.Errorf("per-req: %v, want {1:1.0, 2:0.5, 3:0.0}", perReq)
	}
}

// Если у требования несколько закрывающих контролей внедрено — берём
// MAX coverage_weight (а не сумму, не среднее).
func TestCalcOverview_MaxWins(t *testing.T) {
	reqs := []domain.ComplianceRequirement{{ID: 1, Code: "R1"}}
	links := []domain.RequirementControlLink{
		{RequirementID: 1, ControlID: 10, CoverageWeight: 0.4},
		{RequirementID: 1, ControlID: 20, CoverageWeight: 0.7},
		{RequirementID: 1, ControlID: 30, CoverageWeight: 0.3},
	}
	attached := map[int64]bool{10: true, 20: true, 30: true}

	overall, _, _, _, perReq := calcOverview(reqs, links, attached)

	if perReq[1] != 0.7 {
		t.Errorf("per-req max = %v, want 0.7 (max(0.4, 0.7, 0.3))", perReq[1])
	}
	if overall != 0.7 {
		t.Errorf("overall = %v, want 0.7", overall)
	}
}

// Когда ничего не внедрено — все требования uncovered, overall = 0.
func TestCalcOverview_NoControlsAttached(t *testing.T) {
	overall, covered, partial, uncov, _ := calcOverview(sampleReqs(), sampleLinks(), map[int64]bool{})

	if overall != 0.0 {
		t.Errorf("overall = %v, want 0", overall)
	}
	if covered != 0 || partial != 0 || uncov != 3 {
		t.Errorf("counts %d/%d/%d, want 0/0/3", covered, partial, uncov)
	}
}

// Когда все требования полностью закрыты — overall = 1, all covered.
func TestCalcOverview_FullCoverage(t *testing.T) {
	reqs := []domain.ComplianceRequirement{
		{ID: 1, Code: "R1"},
		{ID: 2, Code: "R2"},
	}
	links := []domain.RequirementControlLink{
		{RequirementID: 1, ControlID: 1, CoverageWeight: 1.0},
		{RequirementID: 2, ControlID: 2, CoverageWeight: 1.0},
	}
	attached := map[int64]bool{1: true, 2: true}

	overall, covered, partial, uncov, _ := calcOverview(reqs, links, attached)
	if overall != 1.0 {
		t.Errorf("overall = %v, want 1.0", overall)
	}
	if covered != 2 || partial != 0 || uncov != 0 {
		t.Errorf("counts %d/%d/%d, want 2/0/0", covered, partial, uncov)
	}
}

// Пустой набор требований — деление на ноль, overall=0, не паника.
func TestCalcOverview_EmptyRequirements(t *testing.T) {
	overall, covered, partial, uncov, perReq := calcOverview(nil, nil, map[int64]bool{1: true})
	if overall != 0.0 || covered != 0 || partial != 0 || uncov != 0 {
		t.Errorf("на пустом списке должны быть нули, получили %v/%d/%d/%d",
			overall, covered, partial, uncov)
	}
	if perReq == nil {
		t.Errorf("perReq должно быть не nil (пустая map)")
	}
}

// Требование без рёбер вообще — uncovered, не паникует.
func TestCalcOverview_RequirementWithoutLinks(t *testing.T) {
	reqs := []domain.ComplianceRequirement{
		{ID: 1, Code: "R1"},
		{ID: 2, Code: "R2-без-рёбер"},
	}
	links := []domain.RequirementControlLink{
		{RequirementID: 1, ControlID: 1, CoverageWeight: 1.0},
	}
	attached := map[int64]bool{1: true}

	overall, covered, partial, uncov, perReq := calcOverview(reqs, links, attached)
	if perReq[2] != 0.0 {
		t.Errorf("требование без рёбер должно иметь coverage 0, получили %v", perReq[2])
	}
	if covered != 1 || uncov != 1 || partial != 0 {
		t.Errorf("counts %d/%d/%d, want 1/0/1", covered, partial, uncov)
	}
	if overall != 0.5 {
		t.Errorf("overall = %v, want 0.5", overall)
	}
}

func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}
