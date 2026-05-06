package main

import "testing"

// resolveCategory — 3-проходный маппер subcategoryid Минцифры в наши 14 кодов.
// Ошибки здесь = 26k продуктов сваливаются в неправильные категории, и
// фильтры в UI «найти все антивирусы» / «найти все СУБД» начинают врать.
func TestResolveCategory(t *testing.T) {
	// Загрузим тестовый catByCode со всеми 14 нашими кодами + their fake ids.
	cat := map[string]int16{
		"os": 1, "dbms": 2, "erp": 3, "crm": 4, "office": 5,
		"antivirus": 6, "backup": 7, "monitoring": 8, "virtualization": 9,
		"network": 10, "development": 11, "web": 12, "mail": 13, "other": 14,
	}

	tests := []struct {
		desc string
		subs []apiSubcategory
		want int16
	}{
		// 1) Точный код subcategoryid.
		{"02.07 → СУБД", []apiSubcategory{{Code: "02.07", Name: "..."}}, 2},
		{"06.04 → почта", []apiSubcategory{{Code: "06.04", Name: "..."}}, 13},
		{"09.07 → ERP", []apiSubcategory{{Code: "09.07", Name: "..."}}, 3},
		{"09.09 → CRM", []apiSubcategory{{Code: "09.09", Name: "..."}}, 4},
		{"02.04 → виртуализация", []apiSubcategory{{Code: "02.04", Name: "..."}}, 9},
		{"01.02 → ОС", []apiSubcategory{{Code: "01.02", Name: "..."}}, 1},
		{"04.03 (особый случай: офисные внутри разработки)",
			[]apiSubcategory{{Code: "04.03", Name: "..."}}, 5},

		// 2) Префикс XX.* — точного кода нет, fallback на префикс.
		{"03.99 → antivirus (вся секция ИБ)",
			[]apiSubcategory{{Code: "03.99", Name: "Какое-то СЗИ"}}, 6},
		{"04.01 → development",
			[]apiSubcategory{{Code: "04.01", Name: "Средства подготовки исполнимого кода"}}, 11},

		// 3) Keyword по имени — для незнакомых кодов.
		{"неизвестный код, в названии 'почтовый' → mail",
			[]apiSubcategory{{Code: "99.99", Name: "Почтовая система"}}, 13},
		{"неизвестный код, 'виртуализация' → virtualization",
			[]apiSubcategory{{Code: "77.77", Name: "Средства виртуализации"}}, 9},
		{"неизвестный код, 'антивирус' → antivirus",
			[]apiSubcategory{{Code: "88.88", Name: "Антивирусная защита"}}, 6},

		// 4) Fallback на other.
		{"совсем неизвестные код и имя → other",
			[]apiSubcategory{{Code: "99.99", Name: "Стрижка единорогов"}}, 14},
		{"пустой список подкатегорий → other",
			[]apiSubcategory{}, 14},

		// Приоритет: первый совпавший точный код побеждает префикс.
		{"приоритет: точный 02.07 побеждает префикс 02.*",
			[]apiSubcategory{{Code: "02.99", Name: "..."}, {Code: "02.07", Name: "..."}}, 2},
		// При нескольких подкатегориях возвращается первая совпавшая.
		{"несколько подкатегорий — берётся первая по точному коду",
			[]apiSubcategory{
				{Code: "06.04", Name: "Почта"},   // exact → mail (id 13)
				{Code: "09.07", Name: "ERP"},     // exact → erp (id 3)
			}, 13},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := resolveCategory(tt.subs, cat)
			if got == nil {
				t.Fatalf("resolveCategory returned nil, want %d", tt.want)
			}
			if *got != tt.want {
				t.Errorf("resolveCategory = %d, want %d", *got, tt.want)
			}
		})
	}
}

// Если в catByCode нет «other», fallback должен мягко вернуть nil.
func TestResolveCategory_NoOtherFallback(t *testing.T) {
	cat := map[string]int16{"os": 1} // нет «other»
	got := resolveCategory([]apiSubcategory{{Code: "99.99", Name: "Что-то"}}, cat)
	if got != nil {
		t.Errorf("ожидался nil когда нечем fallback'нуть, получили %d", *got)
	}
}
