package risk

import (
	"testing"

	"Diplom/internal/domain"
)

func TestComputeAssetAggregate_Empty(t *testing.T) {
	agg := ComputeAssetAggregate(nil)
	if agg.ThreatCount != 0 {
		t.Errorf("ThreatCount = %d, want 0", agg.ThreatCount)
	}
	if agg.WMax != 0 {
		t.Errorf("WMax = %v, want 0", agg.WMax)
	}
	if agg.Level != "info" {
		t.Errorf("Level = %q, want info", agg.Level)
	}
	if agg.UncoveredCount != 0 {
		t.Errorf("UncoveredCount = %d, want 0", agg.UncoveredCount)
	}
}

func TestComputeAssetAggregate_TwoPaths(t *testing.T) {
	paths := []domain.AttackPath{
		{
			W: 0.6, Level: "high",
			VulnerableLinks: []domain.VLNode{{Uncovered: false}, {Uncovered: false}},
		},
		{
			W: 0.84, Level: "critical",
			VulnerableLinks: []domain.VLNode{{Uncovered: true}, {Uncovered: false}},
		},
	}
	agg := ComputeAssetAggregate(paths)
	if agg.ThreatCount != 2 {
		t.Errorf("ThreatCount = %d, want 2", agg.ThreatCount)
	}
	if agg.WMax != 0.84 {
		t.Errorf("WMax = %v, want 0.84", agg.WMax)
	}
	if agg.Level != "critical" {
		t.Errorf("Level = %q, want critical", agg.Level)
	}
	if agg.UncoveredCount != 1 {
		t.Errorf("UncoveredCount = %d, want 1 (only the second path has any uncovered VL)", agg.UncoveredCount)
	}
}

func TestComputeAssetAggregate_MultipleUncoveredVLsInOnePath(t *testing.T) {
	paths := []domain.AttackPath{
		{
			W: 0.5, Level: "high",
			VulnerableLinks: []domain.VLNode{{Uncovered: true}, {Uncovered: true}, {Uncovered: true}},
		},
	}
	agg := ComputeAssetAggregate(paths)
	if agg.UncoveredCount != 1 {
		t.Errorf("UncoveredCount = %d, want 1 (count of paths with >=1 uncovered VL, not VLs)", agg.UncoveredCount)
	}
}

func TestComputeAssetAggregate_NoneUncovered(t *testing.T) {
	paths := []domain.AttackPath{
		{W: 0.3, VulnerableLinks: []domain.VLNode{{Uncovered: false}, {Uncovered: false}}},
		{W: 0.5, VulnerableLinks: []domain.VLNode{{Uncovered: false}}},
	}
	agg := ComputeAssetAggregate(paths)
	if agg.UncoveredCount != 0 {
		t.Errorf("UncoveredCount = %d, want 0 (no uncovered VLs on any path)", agg.UncoveredCount)
	}
	if agg.WMax != 0.5 {
		t.Errorf("WMax = %v, want 0.5", agg.WMax)
	}
}

func TestComputeAssetAggregate_PathWithNoVLs(t *testing.T) {
	paths := []domain.AttackPath{
		{W: 0.7, VulnerableLinks: nil},
	}
	agg := ComputeAssetAggregate(paths)
	if agg.ThreatCount != 1 {
		t.Errorf("ThreatCount = %d, want 1", agg.ThreatCount)
	}
	if agg.UncoveredCount != 0 {
		t.Errorf("UncoveredCount = %d, want 0 (no VLs to be uncovered)", agg.UncoveredCount)
	}
}
