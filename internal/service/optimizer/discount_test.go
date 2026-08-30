package optimizer

import (
	"math"
	"testing"
	"time"
)

func TestDiscountFactor(t *testing.T) {
	// Затраты первого года не приводятся: они тратятся сейчас.
	if f := discountFactor(0.1, 0); f != 1 {
		t.Fatalf("нулевой год не должен дисконтироваться, получено %f", f)
	}
	// 1 / 1.1 ≈ 0.909
	if f := discountFactor(0.1, 1); math.Abs(f-0.9091) > 1e-4 {
		t.Fatalf("год 1 при ставке 10%%: ожидалось ≈0.9091, получено %f", f)
	}
	// Выключенная ставка ничего не меняет.
	if f := discountFactor(0, 3); f != 1 {
		t.Fatalf("без ставки приведения не должно быть, получено %f", f)
	}
}

// Приведённая стоимость плана, растянутого во времени, меньше номинальной.
func TestPresentValue_LaterSpendingIsCheaper(t *testing.T) {
	periods := []Period{
		{Year: 1, Spent: 100_000},
		{Year: 2, Spent: 100_000},
		{Year: 3, Spent: 100_000},
	}

	nominal := 300_000.0
	pv := presentValue(periods, 0.1)

	if pv >= nominal {
		t.Fatalf("приведённая стоимость должна быть меньше номинальной: %.0f против %.0f", pv, nominal)
	}
	if off := presentValue(periods, 0); off != nominal {
		t.Fatalf("без ставки приведённая обязана совпадать с номинальной: %.0f", off)
	}
	t.Logf("номинал %.0f ₽, приведённая при 10%% — %.0f ₽", nominal, pv)
}

func TestDegradedEffectiveness(t *testing.T) {
	// Без ставки старения эффективность не меняется.
	if got := degradedEffectiveness(0.8, 0, 36); got != 0.8 {
		t.Fatalf("без деградации ожидалось 0.8, получено %f", got)
	}
	// 0.8 × 0.9 = 0.72 через год при ставке 10%.
	if got := degradedEffectiveness(0.8, 0.1, 12); math.Abs(got-0.72) > 1e-9 {
		t.Fatalf("через год при 10%%: ожидалось 0.72, получено %f", got)
	}
	// Со временем только убывает.
	year1 := degradedEffectiveness(0.8, 0.1, 12)
	year3 := degradedEffectiveness(0.8, 0.1, 36)
	if year3 >= year1 {
		t.Fatalf("защита должна слабеть: год 1 %.4f, год 3 %.4f", year1, year3)
	}
}

// Старение защиты увеличивает площадь под кривой: та же закупка со временем
// удерживает риск хуже.
func TestDegradation_IncreasesRiskArea(t *testing.T) {
	paths := pathSet{makePath(0.7, 0.7, 1.0, map[string]float64{"A": 0.9})}
	purchases := []Purchase{{
		Candidate:       Candidate{ControlCode: "A", Effectiveness: 0.8},
		ActiveFromMonth: 0,
	}}

	fresh := riskArea(monthlyW(paths, purchases, 36, 0))
	aging := riskArea(monthlyW(paths, purchases, 36, 0.15))

	if aging <= fresh {
		t.Fatalf("со старением площадь обязана расти: без %.4f, со старением %.4f", fresh, aging)
	}
	t.Logf("площадь за 3 года: без старения %.4f, при 15%%/год %.4f", fresh, aging)
}

// Некорректные значения трактуются как «не задано»: расчёт по мусору хуже,
// чем расчёт без поправки.
func TestNormalizeRate_RejectsOutOfRange(t *testing.T) {
	cases := []float64{-0.5, 0.9, 100}
	for _, v := range cases {
		if got := normalizeRate(v, maxDiscountRate); got != 0 {
			t.Errorf("значение %f должно обнуляться, получено %f", v, got)
		}
	}
	if got := normalizeRate(0.12, maxDiscountRate); got != 0.12 {
		t.Errorf("допустимая ставка должна проходить, получено %f", got)
	}
}

// При старении защиты планировщик по-прежнему укладывается в годовой бюджет.
func TestPlanRoadmap_WithDegradationRespectsBudget(t *testing.T) {
	paths := pathSet{makePath(0.8, 0.8, 1.0, map[string]float64{"A": 0.9, "FW": 0.8})}
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 60_000, TotalCost: 60_000, Effectiveness: 0.8, LicenseModel: "per_node"},
		{ControlCode: "FW", CostMax: 70_000, TotalCost: 70_000, Effectiveness: 0.8, LicenseModel: "per_node"},
	}
	sortCandidates(candidates)

	periods, _ := planRoadmap(paths, candidates, 100_000, 3,
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 0.1)

	for _, p := range periods {
		if p.Spent > 100_000+1e-9 {
			t.Fatalf("год %d: бюджет превышен (%.0f)", p.Year, p.Spent)
		}
	}
}
