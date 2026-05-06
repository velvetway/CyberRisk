package main

import "testing"

// Q^угроза маппится из «Возможности нарушителя» БДУ ФСТЭК. Эти 3 значения
// потом попадают в формулу W; кейс «unknown → 0.6» особенно важен —
// БДУ часто даёт пустое поле и мы не должны падать.
func TestDeriveQThreat(t *testing.T) {
	tests := []struct {
		power string
		want  float64
	}{
		{"low", 0.3},
		{"medium", 0.6},
		{"high", 0.9},
		{"", 0.6},          // unknown → medium fallback
		{"UNKNOWN", 0.6},   // случайные значения тоже → medium
		{"VERY HIGH", 0.6}, // нестандартные → medium
	}
	for _, tt := range tests {
		t.Run(tt.power, func(t *testing.T) {
			got := deriveQThreat(tt.power)
			if got != tt.want {
				t.Errorf("deriveQThreat(%q) = %.2f, want %.2f", tt.power, got, tt.want)
			}
		})
	}
}

// q^серьёзность считается по числу нарушаемых из C/I/A.
// Дискретизация нашего изобретения, не диплома, поэтому критично что
// она стабильна — если поменять, поменяются W во всей системе.
func TestDeriveQSeverity(t *testing.T) {
	tests := []struct {
		c, i, a bool
		want    float64
		desc    string
	}{
		{false, false, false, 0.3, "ничего"},
		{true, false, false, 0.5, "только C"},
		{false, true, false, 0.5, "только I"},
		{false, false, true, 0.5, "только A"},
		{true, true, false, 0.7, "C+I"},
		{true, false, true, 0.7, "C+A"},
		{false, true, true, 0.7, "I+A"},
		{true, true, true, 0.9, "все три (C+I+A)"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := deriveQSeverity(tt.c, tt.i, tt.a)
			if got != tt.want {
				t.Errorf("deriveQSeverity(%v,%v,%v) = %.2f, want %.2f",
					tt.c, tt.i, tt.a, got, tt.want)
			}
		})
	}
}
