package risk

import (
	"math"
	"testing"

	"Diplom/internal/domain"
)

func approxEq(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestCalculateW_AllUncoveredMaxSeverity(t *testing.T) {
	// Q_threat=1, q_severity=1, Q_reaction=0, Z=1 → W = (1 + 1 + 1) / 3 * 1 = 1.0
	got := CalculateW(1.0, 1.0, 0.0, 1.0)
	if !approxEq(got, 1.0) {
		t.Fatalf("expected 1.0, got %v", got)
	}
}

func TestCalculateW_FullyCovered(t *testing.T) {
	// Q_threat=1, q_severity=1, Q_reaction=1, Z=1 → W = (1 + 1 + 0) / 3 * 1 = 2/3
	got := CalculateW(1.0, 1.0, 1.0, 1.0)
	if !approxEq(got, 2.0/3.0) {
		t.Fatalf("expected 0.6666..., got %v", got)
	}
}

func TestCalculateW_OneContour(t *testing.T) {
	// Z=0.5 halves the result
	got := CalculateW(1.0, 1.0, 0.0, 0.5)
	if !approxEq(got, 0.5) {
		t.Fatalf("expected 0.5, got %v", got)
	}
}

func TestCalculateW_ClampedToUnitInterval(t *testing.T) {
	// Inputs out of [0,1] are clamped. Z clamped to [0.5, 1.0].
	got := CalculateW(1.5, 1.5, -0.5, 2.0)
	// (1 + 1 + (1 - 0)) / 3 * 1 = 1.0
	if !approxEq(got, 1.0) {
		t.Fatalf("expected 1.0, got %v", got)
	}
}

func TestCalculateW_ClampsNegativeZ(t *testing.T) {
	// Z below 0.5 clamps to 0.5.
	got := CalculateW(1.0, 1.0, 0.0, -1.0)
	if !approxEq(got, 0.5) {
		t.Fatalf("expected 0.5, got %v", got)
	}
}

func TestLevelFromW(t *testing.T) {
	cases := map[float64]string{
		0.00: "low",
		0.24: "low",
		0.25: "medium",
		0.49: "medium",
		0.50: "high",
		0.74: "high",
		0.75: "critical",
		1.00: "critical",
	}
	for w, want := range cases {
		if got := LevelFromW(w); got != want {
			t.Errorf("LevelFromW(%v) = %q; want %q", w, got, want)
		}
	}
}

func TestQReactionFromVLs_AllCovered(t *testing.T) {
	vls := []domain.VLNode{
		{VulnerabilityID: 1, CoverageControls: []domain.ControlCoverage{{Coverage: 1.0}}},
		{VulnerabilityID: 2, CoverageControls: []domain.ControlCoverage{{Coverage: 1.0}}},
	}
	if got := QReactionFromVLs(vls); !approxEq(got, 1.0) {
		t.Fatalf("expected 1.0, got %v", got)
	}
}

func TestQReactionFromVLs_HalfCovered(t *testing.T) {
	vls := []domain.VLNode{
		{VulnerabilityID: 1, CoverageControls: []domain.ControlCoverage{{Coverage: 1.0}}},
		{VulnerabilityID: 2, CoverageControls: nil},
	}
	if got := QReactionFromVLs(vls); !approxEq(got, 0.5) {
		t.Fatalf("expected 0.5, got %v", got)
	}
}

func TestQReactionFromVLs_ZeroCoverageIgnored(t *testing.T) {
	// A control with coverage=0 does NOT count as "covering"
	vls := []domain.VLNode{
		{VulnerabilityID: 1, CoverageControls: []domain.ControlCoverage{{Coverage: 0.0}}},
	}
	if got := QReactionFromVLs(vls); !approxEq(got, 0.0) {
		t.Fatalf("expected 0.0, got %v", got)
	}
}

func TestQReactionFromVLs_Empty(t *testing.T) {
	if got := QReactionFromVLs(nil); !approxEq(got, 0.0) {
		t.Fatalf("empty VL list must yield 0.0, got %v", got)
	}
}

func TestZFromAsset_ProdNonIsolated(t *testing.T) {
	a := domain.Asset{Environment: "prod", IsIsolated: false}
	if got := ZFromAsset(a); !approxEq(got, 1.0) {
		t.Fatalf("prod+non-isolated must be 1.0, got %v", got)
	}
}

func TestZFromAsset_Isolated(t *testing.T) {
	a := domain.Asset{Environment: "prod", IsIsolated: true}
	if got := ZFromAsset(a); !approxEq(got, 0.5) {
		t.Fatalf("isolated asset must be 0.5, got %v", got)
	}
}

func TestZFromAsset_Stage(t *testing.T) {
	a := domain.Asset{Environment: "stage", IsIsolated: false}
	if got := ZFromAsset(a); !approxEq(got, 0.75) {
		t.Fatalf("stage asset must be 0.75, got %v", got)
	}
}

func TestZFromAsset_Dev(t *testing.T) {
	a := domain.Asset{Environment: "dev", IsIsolated: false}
	if got := ZFromAsset(a); !approxEq(got, 0.5) {
		t.Fatalf("dev asset must be 0.5, got %v", got)
	}
}
