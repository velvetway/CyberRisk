package risk

import (
	"testing"

	"Diplom/internal/domain"
)

func i16ptr(v int16) *int16 { return &v }

func TestIsApplicable_EmptyTargets_AppliesEverywhere(t *testing.T) {
	asset := domain.Asset{AssetTypeID: i16ptr(2)}
	threat := domain.Threat{AppliesToAssetTypes: nil}
	if !IsApplicable(asset, threat) {
		t.Fatal("threat without targets must apply to any asset")
	}
}

func TestIsApplicable_AssetTypeMatches(t *testing.T) {
	asset := domain.Asset{AssetTypeID: i16ptr(2)}
	threat := domain.Threat{AppliesToAssetTypes: []int16{1, 2, 3}}
	if !IsApplicable(asset, threat) {
		t.Fatal("expected applicable")
	}
}

func TestIsApplicable_AssetTypeDoesNotMatch(t *testing.T) {
	asset := domain.Asset{AssetTypeID: i16ptr(7)}
	threat := domain.Threat{AppliesToAssetTypes: []int16{1, 2, 3}}
	if IsApplicable(asset, threat) {
		t.Fatal("expected NOT applicable: asset type 7 not in {1,2,3}")
	}
}

func TestIsApplicable_AssetWithoutType_PassesThrough(t *testing.T) {
	// Conservative behavior: if asset has no type but threat targets some,
	// we still show the pair (we can't prove non-applicability).
	asset := domain.Asset{AssetTypeID: nil}
	threat := domain.Threat{AppliesToAssetTypes: []int16{1, 2}}
	if !IsApplicable(asset, threat) {
		t.Fatal("expected applicable when asset type is unknown")
	}
}
