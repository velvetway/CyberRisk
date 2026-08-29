package optimizer

import (
	"testing"

	"Diplom/internal/domain"
)

// makePath строит сценарий с одним уязвимым звеном, которое закрывают
// перечисленные меры.
func makePath(qThreat, qSeverity, z float64, controls map[string]float64) domain.PTSZIAttackPath {
	cov := make([]domain.PTSZIControlCoverage, 0, len(controls))
	id := int16(0)
	for code, coverage := range controls {
		id++
		cov = append(cov, domain.PTSZIControlCoverage{
			Control:  domain.PTSZIControl{ID: id, Code: code, Name: code},
			Coverage: coverage,
		})
	}
	return domain.PTSZIAttackPath{
		QThreat:   qThreat,
		QSeverity: qSeverity,
		Z:         z,
		VulnerableLinks: []domain.PTSZIPathVL{
			{VulnerableLink: domain.PTSZIVulnerableLink{ID: 1, Code: "VL1"}, Controls: cov},
		},
	}
}

func TestRecalcW_NoControlsMatchesFormula(t *testing.T) {
	p := makePath(0.6, 0.8, 1.0, map[string]float64{"A": 0.9})

	// Без внедрённых мер Q_reaction = 0, значит W = (0.6+0.8+1)/3.
	got := recalcW(&p, nil)
	want := (0.6 + 0.8 + 1.0) / 3.0
	if diff := got - want; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("W без мер = %.6f, ожидалось %.6f", got, want)
	}
}

func TestRecalcW_AddedControlReducesW(t *testing.T) {
	p := makePath(0.6, 0.8, 1.0, map[string]float64{"A": 0.9})

	before := recalcW(&p, nil)
	after := recalcW(&p, map[string]float64{"A": 0.8})

	if after >= before {
		t.Fatalf("внедрение меры не снизило W: было %.4f, стало %.4f", before, after)
	}

	// Покрытие 0.9 * эффективность 0.8 = 0.72, значит Q_reaction = 0.72.
	want := (0.6 + 0.8 + (1 - 0.72)) / 3.0
	if diff := after - want; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("W с мерой = %.6f, ожидалось %.6f", after, want)
	}
}

// Две меры на одном звене не должны складываться: модель вероятностная,
// вторая мера добавляет лишь то, что пропустила первая.
func TestRecalcW_TwoControlsSaturate(t *testing.T) {
	p := makePath(0.5, 0.5, 1.0, map[string]float64{"A": 0.6, "FW": 0.6})

	base := recalcW(&p, nil)
	one := recalcW(&p, map[string]float64{"A": 1.0})
	two := recalcW(&p, map[string]float64{"A": 1.0, "FW": 1.0})

	gainFirst := base - one
	gainSecond := one - two

	if gainSecond >= gainFirst {
		t.Fatalf("вторая мера дала не меньше первой: первая %.4f, вторая %.4f", gainFirst, gainSecond)
	}
	if two <= 0 {
		t.Fatalf("W не может быть неположительным: %.4f", two)
	}
}

func TestGreedy_RespectsBudget(t *testing.T) {
	paths := []domain.PTSZIAttackPath{
		makePath(0.7, 0.7, 1.0, map[string]float64{"A": 0.8, "FW": 0.8, "IDS": 0.8}),
	}
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 100_000, Effectiveness: 0.8},
		{ControlCode: "FW", CostMax: 400_000, Effectiveness: 0.8},
		{ControlCode: "IDS", CostMax: 900_000, Effectiveness: 0.8},
	}

	steps, spent := greedy(paths, candidates, 500_000)

	if spent > 500_000 {
		t.Fatalf("бюджет превышен: потрачено %.0f", spent)
	}
	if len(steps) == 0 {
		t.Fatal("не выбрано ни одной меры, хотя бюджет позволяет")
	}
	for _, s := range steps {
		if s.DeltaW <= 0 {
			t.Fatalf("шаг %s не снижает риск: delta=%.6f", s.Candidate.ControlCode, s.DeltaW)
		}
	}
}

// Жадный выбор не обязан быть оптимальным, но не должен быть лучше точного
// оптимума: это означало бы ошибку в одном из них.
func TestGreedy_NeverBeatsExhaustive(t *testing.T) {
	paths := []domain.PTSZIAttackPath{
		makePath(0.8, 0.6, 1.0, map[string]float64{"A": 0.9, "FW": 0.5, "IDS": 0.7, "L": 0.4}),
		makePath(0.4, 0.9, 0.5, map[string]float64{"FW": 0.8, "L": 0.6}),
	}
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 120_000, Effectiveness: 0.8},
		{ControlCode: "FW", CostMax: 300_000, Effectiveness: 0.8},
		{ControlCode: "IDS", CostMax: 250_000, Effectiveness: 0.8},
		{ControlCode: "L", CostMax: 180_000, Effectiveness: 0.8},
	}
	sortCandidates(candidates)

	budget := 600_000.0
	baseline := totalW(paths, nil)
	steps, _ := greedy(paths, candidates, budget)

	greedyDelta := 0.0
	if len(steps) > 0 {
		greedyDelta = baseline - steps[len(steps)-1].WAfter
	}

	best, ok := exhaustive(paths, candidates, budget)
	if !ok {
		t.Fatal("перебор не выполнился на четырёх кандидатах")
	}
	if greedyDelta > best+1e-9 {
		t.Fatalf("жадный (%.6f) обогнал точный оптимум (%.6f) — ошибка в расчёте", greedyDelta, best)
	}
	t.Logf("жадный %.6f, оптимум %.6f, отставание %.2f%%", greedyDelta, best,
		100*(best-greedyDelta)/best)
}

func TestExhaustive_RejectsOverBudget(t *testing.T) {
	paths := []domain.PTSZIAttackPath{
		makePath(0.9, 0.9, 1.0, map[string]float64{"A": 0.9}),
	}
	candidates := []Candidate{{ControlCode: "A", CostMax: 1_000_000, Effectiveness: 0.9}}

	if best, ok := exhaustive(paths, candidates, 10_000); !ok || best != 0 {
		t.Fatalf("при недостаточном бюджете ожидалось нулевое снижение, получено %.6f", best)
	}
}
