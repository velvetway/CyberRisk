package report

import (
	"testing"

	"Diplom/internal/domain"
)

// joinControls — простая вариативная склейка, но это в PDF попадает,
// поэтому проверяем edge-кейсы (пустой / 1 / N).
func TestJoinControls(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, "—"},
		{[]string{}, "—"},
		{[]string{"Антивирус"}, "Антивирус"},
		{[]string{"Антивирус", "МСЭ"}, "Антивирус, МСЭ"},
		{[]string{"A", "B", "C", "D"}, "A, B, C, D"},
	}
	for _, tt := range tests {
		got := joinControls(tt.in)
		if got != tt.want {
			t.Errorf("joinControls(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// countUncoveredVL агрегирует уникальные uncovered VL по всем угрозам.
// Это число попадает в шапку «Сводное состояние защиты» отчёта —
// важно что одна и та же VL у двух угроз считается ОДИН раз.
func TestCountUncoveredVL(t *testing.T) {
	t.Run("пустой ввод", func(t *testing.T) {
		if got := countUncoveredVL(nil); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("одна угроза, две uncovered VL", func(t *testing.T) {
		paths := []domain.AttackPath{{
			VulnerableLinks: []domain.VLNode{
				{CategoryID: 1, Uncovered: true},
				{CategoryID: 2, Uncovered: true},
				{CategoryID: 3, CoverageControls: []domain.ControlCoverage{{ID: 1}}}, // покрыта
			},
		}}
		if got := countUncoveredVL(paths); got != 2 {
			t.Errorf("got %d, want 2", got)
		}
	})

	t.Run("дедуп по category_id между угрозами", func(t *testing.T) {
		paths := []domain.AttackPath{
			{VulnerableLinks: []domain.VLNode{
				{CategoryID: 1, Uncovered: true},
				{CategoryID: 2, Uncovered: true},
			}},
			{VulnerableLinks: []domain.VLNode{
				{CategoryID: 1, Uncovered: true}, // та же VL — должна посчитаться один раз
				{CategoryID: 4, Uncovered: true},
			}},
		}
		// Уникальных uncovered VL: 1, 2, 4 → 3
		if got := countUncoveredVL(paths); got != 3 {
			t.Errorf("got %d, want 3 (deduplicated)", got)
		}
	})

	t.Run("VL с CoverageControls пустыми считается uncovered", func(t *testing.T) {
		paths := []domain.AttackPath{{
			VulnerableLinks: []domain.VLNode{
				{CategoryID: 1, Uncovered: false, CoverageControls: nil}, // нет покрытия — uncovered
			},
		}}
		if got := countUncoveredVL(paths); got != 1 {
			t.Errorf("got %d, want 1", got)
		}
	})
}

// groupControlsByMeasure разносит 11 контролей по 4 «мероприятиям»
// из 8.png диплома. Проверяем что состав групп стабилен.
func TestGroupControlsByMeasure(t *testing.T) {
	allControls := []domain.Control{
		{ID: 1, Name: "Антивирус"},
		{ID: 2, Name: "Межсетевой экран"},
		{ID: 3, Name: "Honeypot"},
		{ID: 4, Name: "Демилитаризованная зона"},
		{ID: 5, Name: "Система обнаружения вторжений"},
		{ID: 6, Name: "Системы администрирования"},
		{ID: 7, Name: "Резервное копирование"},
		{ID: 8, Name: "Программная защита от НСД"},
		{ID: 9, Name: "Шифрование трафика"},
		{ID: 10, Name: "Цифровая подпись"},
		{ID: 11, Name: "DDoS-фильтры"},
	}

	// Внедрены только Антивирус и МСЭ.
	deployed := []domain.Control{allControls[0], allControls[1]}
	groups := groupControlsByMeasure(deployed)

	if len(groups) != 4 {
		t.Fatalf("ожидали 4 группы (АРМ/ЛВС/ЭДО/конф.инф.), получили %d", len(groups))
	}

	// АРМ: Антивирус внедрён, остальные из группы — в missing.
	arm := findGroup(groups, "АРМ")
	if arm == nil {
		t.Fatal("группа АРМ не найдена")
	}
	if !contains(arm.deployed, "Антивирус") {
		t.Errorf("Антивирус должен быть в АРМ.deployed, deployed: %v", arm.deployed)
	}

	// ЛВС: только МСЭ внедрён.
	lvs := findGroup(groups, "ЛВС")
	if lvs == nil {
		t.Fatal("группа ЛВС не найдена")
	}
	if !contains(lvs.deployed, "Межсетевой экран") {
		t.Errorf("МСЭ должен быть в ЛВС.deployed, deployed: %v", lvs.deployed)
	}
	// В ЛВС.missing должны быть Honeypot, ДМЗ, IDS, DDoS-фильтры
	for _, expected := range []string{"Honeypot", "Демилитаризованная зона", "Система обнаружения вторжений", "DDoS-фильтры"} {
		if !contains(lvs.missing, expected) {
			t.Errorf("%q должен быть в ЛВС.missing", expected)
		}
	}

	// Документооборот без внедрённых — все в missing.
	doc := findGroup(groups, "электронного документооборота")
	if doc == nil {
		t.Fatal("группа документооборота не найдена")
	}
	if len(doc.deployed) != 0 {
		t.Errorf("в документообороте ничего не внедрено, но deployed=%v", doc.deployed)
	}
	for _, expected := range []string{"Шифрование трафика", "Цифровая подпись"} {
		if !contains(doc.missing, expected) {
			t.Errorf("%q должен быть в документооборот.missing", expected)
		}
	}
}

func findGroup(groups []measureSection, contains string) *measureSection {
	for i := range groups {
		if substr(groups[i].name, contains) {
			return &groups[i]
		}
	}
	return nil
}

func contains(arr []string, s string) bool {
	for _, x := range arr {
		if x == s {
			return true
		}
	}
	return false
}

func substr(s, sub string) bool {
	return len(s) >= len(sub) && stringIndex(s, sub) >= 0
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
