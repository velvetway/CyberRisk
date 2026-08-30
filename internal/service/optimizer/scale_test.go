package optimizer

import "testing"

func TestPricingUnit(t *testing.T) {
	cases := map[string]string{
		"per_node":   "node",
		"per_server": "server",
		"appliance":  "appliance",
		"bundle":     "bundle",
		// Срочные модели описывают не единицу, а срок. В собранных данных
		// и Dallas Lock (perpetual), и Secret Net Studio (yearly) стоят
		// за рабочее место.
		"perpetual": "node",
		"yearly":    "node",
		// Неизвестное безопаснее не умножать на масштаб.
		"":         "bundle",
		"неведомо": "bundle",
	}
	for model, want := range cases {
		if got := pricingUnit(model); got != want {
			t.Errorf("pricingUnit(%q) = %q, ожидалось %q", model, got, want)
		}
	}
}

func TestScaledCost_MultipliesByUnits(t *testing.T) {
	scale := AssetScale{Workstations: 200, Servers: 4, Appliances: 2}

	cases := []struct {
		model string
		want  float64
	}{
		{"per_node", 1000 * 200},
		{"per_server", 1000 * 4},
		{"appliance", 1000 * 2},
		{"bundle", 1000},
	}
	for _, c := range cases {
		got := scaledCost(Candidate{CostMax: 1000, LicenseModel: c.model}, scale)
		if got != c.want {
			t.Errorf("модель %s: стоимость %.0f, ожидалось %.0f", c.model, got, c.want)
		}
	}
}

// Нулевой масштаб не должен обнулять стоимость: пустые параметры запроса
// означают «одна единица», а не «ничего не покупаем».
func TestScale_ZeroBecomesOne(t *testing.T) {
	got := AssetScale{}.normalized()
	if got.Workstations != 1 || got.Servers != 1 || got.Appliances != 1 {
		t.Fatalf("нулевой масштаб должен стать единичным, получено %+v", got)
	}
}

// Главное, ради чего вводился масштаб: на большом активе поштучная лицензия
// перестаёт быть дешёвой, и выбор смещается к аппаратному решению.
func TestScale_ChangesCheapestChoice(t *testing.T) {
	perNode := Candidate{ControlCode: "FW", CostMax: 7_000, LicenseModel: "per_node"}
	appliance := Candidate{ControlCode: "FW", CostMax: 300_000, LicenseModel: "appliance"}

	small := AssetScale{Workstations: 5, Servers: 1, Appliances: 1}
	if scaledCost(perNode, small) >= scaledCost(appliance, small) {
		t.Fatalf("на пяти станциях поштучная лицензия должна быть дешевле: %.0f против %.0f",
			scaledCost(perNode, small), scaledCost(appliance, small))
	}

	large := AssetScale{Workstations: 200, Servers: 1, Appliances: 1}
	if scaledCost(perNode, large) <= scaledCost(appliance, large) {
		t.Fatalf("на двухстах станциях аппаратное решение должно быть дешевле: %.0f против %.0f",
			scaledCost(perNode, large), scaledCost(appliance, large))
	}

	t.Logf("5 станций: поштучно %.0f ₽, шасси %.0f ₽",
		scaledCost(perNode, small), scaledCost(appliance, small))
	t.Logf("200 станций: поштучно %.0f ₽, шасси %.0f ₽",
		scaledCost(perNode, large), scaledCost(appliance, large))
}

// Регрессия: бюджет должен расходоваться по итоговой стоимости, а не по цене
// за единицу. Пока это было не так, план на масштабном активе молча выходил
// за бюджет во столько раз, сколько единиц закупалось.
func TestGreedy_SpendsScaledCostNotUnitPrice(t *testing.T) {
	paths := pathSet{makePath(0.8, 0.8, 1.0, map[string]float64{"A": 0.9, "FW": 0.9})}

	// Цена за единицу мала, но единиц сто: в бюджет влезает только одно средство.
	candidates := []Candidate{
		{ControlCode: "A", CostMax: 1_000, Units: 100, TotalCost: 100_000, Effectiveness: 0.8, LicenseModel: "per_node"},
		{ControlCode: "FW", CostMax: 1_000, Units: 100, TotalCost: 100_000, Effectiveness: 0.8, LicenseModel: "per_node"},
	}
	sortCandidates(candidates)

	steps, spent := greedy(paths, candidates, 150_000)

	if spent > 150_000 {
		t.Fatalf("бюджет превышен: потрачено %.0f при лимите 150000", spent)
	}
	if len(steps) != 1 {
		t.Fatalf("в бюджет должно было влезть ровно одно средство, выбрано %d", len(steps))
	}

	sum := 0.0
	for _, s := range steps {
		sum += s.Candidate.TotalCost
	}
	if diff := sum - spent; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("сумма шагов %.0f не совпадает с потраченным %.0f", sum, spent)
	}
}
