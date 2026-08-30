package optimizer

import "testing"

// Мера без своей опоры обязана работать слабее: обнаружение вторжений
// без администрирования фиксирует атаку, но не останавливает её.
func TestCompatibility_DependencyPenalty(t *testing.T) {
	planned := map[string]float64{"IDS": 0.8}

	alone := applyCompatibility(planned, map[string]bool{"IDS": true})
	withSupport := applyCompatibility(planned, map[string]bool{"IDS": true, "AD": true})

	if alone["IDS"] >= withSupport["IDS"] {
		t.Fatalf("IDS без AD должен быть слабее: без %.3f, с %.3f", alone["IDS"], withSupport["IDS"])
	}
	// 0.8 × 0.6 = 0.48
	if diff := alone["IDS"] - 0.48; diff > 1e-9 || diff < -1e-9 {
		t.Fatalf("ожидался штраф до 0.48, получено %.3f", alone["IDS"])
	}
}

// Пара, работающая вместе, усиливает обе стороны.
func TestCompatibility_SynergyBonus(t *testing.T) {
	planned := map[string]float64{"FW": 0.7, "DZ": 0.7}

	separate := applyCompatibility(map[string]float64{"FW": 0.7}, map[string]bool{"FW": true})
	together := applyCompatibility(planned, map[string]bool{"FW": true, "DZ": true})

	if together["FW"] <= separate["FW"] {
		t.Fatalf("FW вместе с DZ должен быть сильнее: один %.3f, вместе %.3f",
			separate["FW"], together["FW"])
	}
	t.Logf("FW: отдельно %.3f, в паре с DZ %.3f", separate["FW"], together["FW"])
}

// Синергия IDS+AD должна перекрывать штраф за зависимость: вместе они
// работают лучше, чем IDS в одиночку.
func TestCompatibility_SynergyOutweighsPenalty(t *testing.T) {
	alone := applyCompatibility(map[string]float64{"IDS": 0.8}, map[string]bool{"IDS": true})
	paired := applyCompatibility(
		map[string]float64{"IDS": 0.8, "AD": 0.8},
		map[string]bool{"IDS": true, "AD": true},
	)

	if paired["IDS"] <= alone["IDS"] {
		t.Fatalf("IDS с AD должен быть сильнее одинокого: %.3f против %.3f",
			paired["IDS"], alone["IDS"])
	}
	t.Logf("IDS: в одиночку %.3f, вместе с AD %.3f", alone["IDS"], paired["IDS"])
}

// Эффективность не должна вылезать за единицу даже при нескольких бонусах.
func TestCompatibility_StaysInRange(t *testing.T) {
	planned := map[string]float64{"FW": 0.98, "DZ": 0.98, "A": 0.99, "L": 0.99}
	active := map[string]bool{"FW": true, "DZ": true, "A": true, "L": true}

	for code, v := range applyCompatibility(planned, active) {
		if v < 0 || v > 1 {
			t.Fatalf("%s вышел за диапазон: %f", code, v)
		}
	}
}

// Исходная карта не должна меняться: расчёт вызывается многократно, и
// накопленные поправки исказили бы результат.
func TestCompatibility_DoesNotMutateInput(t *testing.T) {
	planned := map[string]float64{"FW": 0.7, "DZ": 0.7}
	applyCompatibility(planned, map[string]bool{"FW": true, "DZ": true})

	if planned["FW"] != 0.7 {
		t.Fatalf("исходная эффективность изменилась: %f", planned["FW"])
	}
}

// Уже внедрённые на активе меры учитываются наравне с планируемыми:
// если администрирование есть, IDS не должен получать штраф.
func TestActiveControls_IncludesImplemented(t *testing.T) {
	path := makePath(0.7, 0.7, 1.0, map[string]float64{"AD": 0.8, "IDS": 0.7})
	// Отмечаем AD как уже внедрённый.
	for i := range path.VulnerableLinks[0].Controls {
		if path.VulnerableLinks[0].Controls[i].Control.Code == "AD" {
			path.VulnerableLinks[0].Controls[i].Implemented = true
			path.VulnerableLinks[0].Controls[i].Effectiveness = 0.8
		}
	}
	paths := pathSet{path}

	active := activeControls(paths, map[string]float64{"IDS": 0.8})
	if !active["AD"] {
		t.Fatal("внедрённый AD не попал в число действующих методов")
	}

	adjusted := applyCompatibility(map[string]float64{"IDS": 0.8}, active)
	if adjusted["IDS"] < 0.8 {
		t.Fatalf("при наличии AD штрафа быть не должно, получено %.3f", adjusted["IDS"])
	}
}

func TestCompatibilityNotes_Explains(t *testing.T) {
	notes := compatibilityNotes(
		map[string]bool{"IDS": true},
		map[string]bool{"IDS": true},
	)
	if len(notes) == 0 {
		t.Fatal("ожидалось пояснение про отсутствие опоры для IDS")
	}
	for _, n := range notes {
		if n.Reason == "" {
			t.Fatalf("правило без объяснения: %+v", n)
		}
		t.Logf("%s: %s ← %s (×%.2f) — %s", n.Kind, n.Control, n.Related, n.Factor, n.Reason)
	}
}
