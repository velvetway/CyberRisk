package report

import "testing"

// statusGlyph — текстовая метка для PDF (вместо ✓◐✗ которые NotoSans
// рендерит как ⬛). Если случайно переименуют константы, тесты упадут.
func TestStatusGlyph(t *testing.T) {
	tests := []struct {
		coverage float64
		want     string
	}{
		{1.0, "OK"},
		{1.5, "OK"}, // > 1 тоже как полностью покрыто
		{0.99, "Част."},
		{0.5, "Част."},
		{0.01, "Част."},
		{0.0, "Нет"},
		{-1.0, "Нет"},
	}
	for _, tt := range tests {
		got := statusGlyph(tt.coverage)
		if got != tt.want {
			t.Errorf("statusGlyph(%v) = %q, want %q", tt.coverage, got, tt.want)
		}
	}
}

// complianceLevelRGB используется для подсветки в PDF: зелёный/жёлтый/
// оранжевый/красный по баллу соответствия.
func TestComplianceLevelRGB(t *testing.T) {
	tests := []struct {
		score    float64
		wantR, wantG, wantB int
		desc     string
	}{
		{1.0, 80, 170, 90, "100% — зелёный"},
		{0.85, 80, 170, 90, "85% — зелёный"},
		{0.8, 80, 170, 90, "ровно 0.8 — зелёный (граница включительно)"},
		{0.79, 235, 195, 60, "79% — жёлтый"},
		{0.5, 235, 195, 60, "50% — жёлтый (граница включительно)"},
		{0.49, 235, 130, 40, "49% — оранжевый"},
		{0.25, 235, 130, 40, "25% — оранжевый"},
		{0.24, 200, 50, 50, "24% — красный"},
		{0.0, 200, 50, 50, "0% — красный"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			r, g, b := complianceLevelRGB(tt.score)
			if r != tt.wantR || g != tt.wantG || b != tt.wantB {
				t.Errorf("complianceLevelRGB(%.2f) = (%d,%d,%d), want (%d,%d,%d)",
					tt.score, r, g, b, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

// levelRGB — для уровней риска critical/high/medium/low. Использовалось
// в renderAggregate для цветной плашки W max.
func TestLevelRGB(t *testing.T) {
	tests := []struct {
		level    string
		wantR, wantG, wantB int
	}{
		{"critical", 200, 50, 50},
		{"high", 235, 130, 40},
		{"medium", 235, 195, 60},
		{"low", 80, 170, 90},
		{"unknown", 110, 110, 110},
		{"", 110, 110, 110},
		{"CRITICAL", 200, 50, 50}, // case-insensitive
	}
	for _, tt := range tests {
		r, g, b := levelRGB(tt.level)
		if r != tt.wantR || g != tt.wantG || b != tt.wantB {
			t.Errorf("levelRGB(%q) = (%d,%d,%d), want (%d,%d,%d)",
				tt.level, r, g, b, tt.wantR, tt.wantG, tt.wantB)
		}
	}
}

func TestLevelLabel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"critical", "Критический"},
		{"CRITICAL", "Критический"},
		{"high", "Высокий"},
		{"medium", "Средний"},
		{"low", "Низкий"},
		{"misc", "misc"}, // unknown — passthrough
	}
	for _, tt := range tests {
		if got := levelLabel(tt.in); got != tt.want {
			t.Errorf("levelLabel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
