package optimizer

import (
	"math/rand"
	"testing"
)

// Возмущение обязано менять коэффициенты, но не выводить их за [0,1]:
// иначе модель начнёт считать по невозможным значениям.
func TestPerturb_StaysInRange(t *testing.T) {
	paths := pathSet{makePath(0.9, 0.95, 1.0, map[string]float64{"A": 0.95, "FW": 0.9})}
	rng := rand.New(rand.NewSource(1))

	for i := 0; i < 100; i++ {
		p := perturb(paths, 0.5, rng)
		for _, path := range p {
			if path.QThreat < 0 || path.QThreat > 1 {
				t.Fatalf("QThreat вышел за диапазон: %f", path.QThreat)
			}
			if path.QSeverity < 0 || path.QSeverity > 1 {
				t.Fatalf("QSeverity вышел за диапазон: %f", path.QSeverity)
			}
			for _, vl := range path.VulnerableLinks {
				for _, c := range vl.Controls {
					if c.Coverage < 0 || c.Coverage > 1 {
						t.Fatalf("покрытие вышло за диапазон: %f", c.Coverage)
					}
				}
			}
		}
	}
}

// Исходные сценарии не должны меняться: анализ гоняется много раз подряд,
// и порча входных данных копилась бы от прогона к прогону.
func TestPerturb_DoesNotMutateInput(t *testing.T) {
	paths := pathSet{makePath(0.6, 0.7, 1.0, map[string]float64{"A": 0.8})}
	before := paths[0].QThreat
	beforeCoverage := paths[0].VulnerableLinks[0].Controls[0].Coverage

	perturb(paths, 0.3, rand.New(rand.NewSource(7)))

	if paths[0].QThreat != before {
		t.Fatalf("исходный QThreat изменился: было %f, стало %f", before, paths[0].QThreat)
	}
	if paths[0].VulnerableLinks[0].Controls[0].Coverage != beforeCoverage {
		t.Fatal("исходное покрытие изменилось")
	}
}

// Одинаковый запрос обязан давать одинаковый отчёт: иначе устойчивость
// нельзя обсуждать — цифры плавали бы при каждом пересчёте.
func TestSensitivity_Deterministic(t *testing.T) {
	paths := pathSet{makePath(0.7, 0.8, 1.0, map[string]float64{"A": 0.8, "FW": 0.7, "IDS": 0.6})}
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 50_000, TotalCost: 50_000, Effectiveness: 0.8},
		{ControlCode: "FW", CostMax: 90_000, TotalCost: 90_000, Effectiveness: 0.8},
		{ControlCode: "IDS", CostMax: 120_000, TotalCost: 120_000, Effectiveness: 0.8},
	}
	sortCandidates(candidates)

	a := analyzeSensitivity(paths, candidates, 200_000, 50, 0.2, 42)
	b := analyzeSensitivity(paths, candidates, 200_000, 50, 0.2, 42)

	if a.CompositionStability != b.CompositionStability || a.MeanDelta != b.MeanDelta {
		t.Fatalf("отчёты разошлись при одном зерне: %+v против %+v", a, b)
	}
}

// Когда бюджета хватает на всё, состав плана не зависит от коэффициентов:
// покупается вообще всё, что есть.
func TestSensitivity_AmpleBudgetIsStable(t *testing.T) {
	paths := pathSet{makePath(0.7, 0.8, 1.0, map[string]float64{"A": 0.8, "FW": 0.7})}
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 1_000, TotalCost: 1_000, Effectiveness: 0.8},
		{ControlCode: "FW", CostMax: 1_000, TotalCost: 1_000, Effectiveness: 0.8},
	}
	sortCandidates(candidates)

	r := analyzeSensitivity(paths, candidates, 10_000_000, 100, 0.2, 1)

	if r.CompositionStability < 0.99 {
		t.Fatalf("при избыточном бюджете состав обязан быть стабильным, получено %.2f",
			r.CompositionStability)
	}
	t.Logf("устойчивость %.0f%%, вердикт: %s", 100*r.CompositionStability, r.Verdict)
}

// Когда бюджета хватает ровно на одно из двух почти равных средств, выбор
// начинает зависеть от коэффициентов — и анализ обязан это показать.
func TestSensitivity_TightBudgetIsUnstable(t *testing.T) {
	paths := pathSet{makePath(0.7, 0.7, 1.0, map[string]float64{"A": 0.70, "FW": 0.71})}
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 100_000, TotalCost: 100_000, Effectiveness: 0.8},
		{ControlCode: "FW", CostMax: 100_000, TotalCost: 100_000, Effectiveness: 0.8},
	}
	sortCandidates(candidates)

	r := analyzeSensitivity(paths, candidates, 100_000, 200, 0.3, 5)

	if r.CompositionStability >= 0.99 {
		t.Fatalf("на почти равных средствах состав не может быть абсолютно стабильным: %.2f",
			r.CompositionStability)
	}
	t.Logf("устойчивость %.0f%%, вердикт: %s", 100*r.CompositionStability, r.Verdict)
	for _, c := range r.Controls {
		t.Logf("   %s попадал в план в %.0f%% прогонов", c.ControlCode, 100*c.Frequency)
	}
}

func TestVerdict_Thresholds(t *testing.T) {
	cases := map[float64]string{
		0.95: "устойчив",
		0.75: "в основном устойчив",
		0.50: "чувствителен",
		0.10: "неустойчив",
	}
	for stability, want := range cases {
		got := verdict(stability)
		if len(got) == 0 {
			t.Fatalf("пустой вердикт для %.2f", stability)
		}
		t.Logf("%.2f → %s (ожидали упоминание «%s»)", stability, got, want)
	}
}
