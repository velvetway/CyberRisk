package optimizer

import (
	"testing"
	"time"

	"Diplom/internal/domain"
)

func TestRiskArea_NoPurchasesEqualsBaseline(t *testing.T) {
	paths := pathSet{makePath(0.6, 0.6, 1.0, map[string]float64{"A": 0.8})}
	baseline := paths.totalW(nil)

	area := riskArea(monthlyW(paths, nil, 36))
	want := baseline * 3

	if diff := area - want; diff > 1e-6 || diff < -1e-6 {
		t.Fatalf("площадь без закупок = %.6f, ожидалось %.6f", area, want)
	}
}

// Средство, купленное раньше, обязано давать большее сокращение площади:
// именно ради этого и вводилась временная метрика.
func TestRiskArea_EarlierPurchaseIsBetter(t *testing.T) {
	paths := pathSet{makePath(0.7, 0.7, 1.0, map[string]float64{"A": 0.9})}
	c := Candidate{ControlCode: "A", CostMax: 100, Effectiveness: 0.8, LicenseModel: "per_node"}

	early := riskArea(monthlyW(paths, []Purchase{{Candidate: c, ActiveFromMonth: 1}}, 36))
	late := riskArea(monthlyW(paths, []Purchase{{Candidate: c, ActiveFromMonth: 24}}, 36))

	if early >= late {
		t.Fatalf("ранняя закупка не выгоднее поздней: рано %.4f, поздно %.4f", early, late)
	}
	t.Logf("площадь: рано %.4f, поздно %.4f, разница %.4f", early, late, late-early)
}

// Срок внедрения должен учитываться: железо разворачивается дольше софта,
// и при равной цене и эффективности софт выгоднее по площади.
func TestDeployDelay_AffectsArea(t *testing.T) {
	paths := pathSet{makePath(0.7, 0.7, 1.0, map[string]float64{"A": 0.9})}
	c := Candidate{ControlCode: "A", CostMax: 100, Effectiveness: 0.8}

	software := riskArea(monthlyW(paths, []Purchase{{Candidate: c, ActiveFromMonth: deployDelay("per_node")}}, 36))
	hardware := riskArea(monthlyW(paths, []Purchase{{Candidate: c, ActiveFromMonth: deployDelay("appliance")}}, 36))

	if software >= hardware {
		t.Fatalf("срок внедрения не влияет: софт %.4f, железо %.4f", software, hardware)
	}
}

func TestPlanRoadmap_RespectsYearlyBudget(t *testing.T) {
	paths := pathSet{
		makePath(0.8, 0.7, 1.0, map[string]float64{"A": 0.9, "FW": 0.8, "IDS": 0.7, "L": 0.6}),
	}
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 80_000, TotalCost: 80_000, Effectiveness: 0.8, LicenseModel: "per_node"},
		{ControlCode: "FW", CostMax: 90_000, TotalCost: 90_000, Effectiveness: 0.8, LicenseModel: "appliance"},
		{ControlCode: "IDS", CostMax: 95_000, TotalCost: 95_000, Effectiveness: 0.8, LicenseModel: "per_server"},
		{ControlCode: "L", CostMax: 85_000, TotalCost: 85_000, Effectiveness: 0.8, LicenseModel: "per_node"},
	}
	sortCandidates(candidates)

	budget := 100_000.0
	periods, purchases := planRoadmap(paths, candidates, budget, 3, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))

	if len(periods) != 3 {
		t.Fatalf("ожидалось 3 периода, получено %d", len(periods))
	}
	for _, p := range periods {
		if p.Spent > budget+1e-9 {
			t.Fatalf("год %d: годовой бюджет превышен (%.0f > %.0f)", p.Year, p.Spent, budget)
		}
	}
	if len(purchases) == 0 {
		t.Fatal("за три года не куплено ничего")
	}

	// Риск обязан снижаться от года к году, раз закупки продолжаются.
	for i := 1; i < len(periods); i++ {
		if periods[i].WAtEnd > periods[i-1].WAtEnd+1e-9 {
			t.Fatalf("риск вырос: год %d конец %.4f, год %d конец %.4f",
				periods[i-1].Year, periods[i-1].WAtEnd, periods[i].Year, periods[i].WAtEnd)
		}
	}

	for _, p := range periods {
		t.Logf("год %d: потрачено %.0f, W %.4f → %.4f, площадь %.4f",
			p.Year, p.Spent, p.WAtStart, p.WAtEnd, p.RiskArea)
	}
}

// Тот же набор средств при большем годовом бюджете покупается раньше,
// значит площадь под кривой должна быть меньше.
func TestPlanRoadmap_BiggerBudgetReducesArea(t *testing.T) {
	build := func() []Candidate {
		c := []Candidate{
			{ControlCode: "A", CostMax: 100_000, TotalCost: 100_000, Effectiveness: 0.8, LicenseModel: "per_node"},
			{ControlCode: "FW", CostMax: 100_000, TotalCost: 100_000, Effectiveness: 0.8, LicenseModel: "per_node"},
			{ControlCode: "IDS", CostMax: 100_000, TotalCost: 100_000, Effectiveness: 0.8, LicenseModel: "per_node"},
		}
		sortCandidates(c)
		return c
	}
	paths := pathSet{makePath(0.8, 0.8, 1.0, map[string]float64{"A": 0.8, "FW": 0.8, "IDS": 0.8})}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, slow := planRoadmap(paths, build(), 100_000, 3, now)
	_, fast := planRoadmap(paths, build(), 300_000, 3, now)

	slowArea := riskArea(monthlyW(paths, slow, 36))
	fastArea := riskArea(monthlyW(paths, fast, 36))

	if fastArea >= slowArea {
		t.Fatalf("больший бюджет не сократил площадь: медленно %.4f, быстро %.4f", slowArea, fastArea)
	}
	t.Logf("площадь: бюджет 100к — %.4f, бюджет 300к — %.4f", slowArea, fastArea)
}

func TestPlanRoadmap_SkipsCandidatesDeployingBeyondHorizon(t *testing.T) {
	paths := pathSet{makePath(0.9, 0.9, 1.0, map[string]float64{"A": 0.9})}
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 1000, TotalCost: 1000, Effectiveness: 0.9, LicenseModel: "appliance"},
	}

	// Горизонт в один год: железо со сроком внедрения 3 месяца всё ещё успевает.
	_, purchases := planRoadmap(paths, candidates, 10_000, 1, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if len(purchases) != 1 {
		t.Fatalf("средство должно было попасть в план, куплено %d", len(purchases))
	}
	if purchases[0].ActiveFromMonth != 3 {
		t.Fatalf("ожидался старт с 3-го месяца, получен %d", purchases[0].ActiveFromMonth)
	}
}

var _ = domain.PTSZIAttackPath{}

// После истечения сертификата средство перестаёт защищать, и риск
// возвращается к исходному уровню — это должно быть видно в помесячном ряду.
func TestExpiry_RiskReturnsAfterCertificateEnds(t *testing.T) {
	paths := pathSet{makePath(0.7, 0.7, 1.0, map[string]float64{"A": 0.9})}
	baseline := paths.totalW(nil)

	expires := 12
	purchases := []Purchase{{
		Candidate:       Candidate{ControlCode: "A", Effectiveness: 0.8},
		ActiveFromMonth: 0,
		ExpiresAtMonth:  &expires,
	}}

	series := monthlyW(paths, purchases, 24)

	if series[6] >= baseline {
		t.Fatalf("до истечения риск должен быть ниже базового: %.4f против %.4f", series[6], baseline)
	}
	if diff := series[12] - baseline; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("после истечения риск должен вернуться к %.4f, получено %.4f", baseline, series[12])
	}
	t.Logf("W: месяц 6 = %.4f, месяц 12 (истёк) = %.4f", series[6], series[12])
}

// Средство с истекающим сертификатом хуже по площади, чем бессрочное:
// планировщик обязан это различать.
func TestExpiry_PerpetualBeatsExpiring(t *testing.T) {
	paths := pathSet{makePath(0.7, 0.7, 1.0, map[string]float64{"A": 0.9})}
	c := Candidate{ControlCode: "A", Effectiveness: 0.8}

	ends := 18
	perpetual := riskArea(monthlyW(paths, []Purchase{{Candidate: c, ActiveFromMonth: 1}}, 36))
	expiring := riskArea(monthlyW(paths, []Purchase{{Candidate: c, ActiveFromMonth: 1, ExpiresAtMonth: &ends}}, 36))

	if perpetual >= expiring {
		t.Fatalf("бессрочный сертификат не выгоднее истекающего: %.4f против %.4f", perpetual, expiring)
	}
	t.Logf("площадь: бессрочный %.4f, истекает на 18-м месяце %.4f", perpetual, expiring)
}

func TestMonthsUntil(t *testing.T) {
	now := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		date string
		want int
	}{
		{"2027-01-15", 12},
		{"2026-07-15", 6},
		{"2025-01-15", 0}, // уже истёк
	}
	for _, c := range cases {
		d := c.date
		got := monthsUntil(&d, now)
		if got == nil || *got != c.want {
			t.Errorf("monthsUntil(%s) = %v, ожидалось %d", c.date, got, c.want)
		}
	}
	if got := monthsUntil(nil, now); got != nil {
		t.Errorf("бессрочный сертификат должен давать nil, получено %v", got)
	}
}

// Покупать средство, сертификат которого истекает раньше окончания
// внедрения, бессмысленно: оно не отработает ни дня.
func TestPlanRoadmap_SkipsCertificateExpiringBeforeDeploy(t *testing.T) {
	paths := pathSet{makePath(0.9, 0.9, 1.0, map[string]float64{"A": 0.9})}
	soon := "2026-02-01"
	candidates := []Candidate{{
		ControlCode:   "A",
		CostMax:       1000,
		TotalCost:     1000,
		Effectiveness: 0.9,
		LicenseModel:  "appliance", // внедрение 3 месяца
		ValidUntil:    &soon,       // а сертификат кончится через месяц
	}}

	_, purchases := planRoadmap(paths, candidates, 10_000, 1, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if len(purchases) != 0 {
		t.Fatalf("средство не должно было попасть в план, куплено %d", len(purchases))
	}
}
