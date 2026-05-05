package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"Diplom/internal/domain"
	"Diplom/internal/service/risk"

	"github.com/gofiber/fiber/v2"
)

// stubRiskSvc — минимальная заглушка risk.Service для тестов хэндлеров.
// Реализован только метод, который тестируем; вызовы остальных приведут
// к nil-pointer panic, но мы их не вызываем.
type stubRiskSvc struct {
	risk.Service
	resp *domain.AssetAttackPathsResponse
	err  error
}

func (s *stubRiskSvc) AssembleAssetAttackPaths(ctx context.Context, assetID int64) (*domain.AssetAttackPathsResponse, error) {
	return s.resp, s.err
}

func TestAssetAttackPaths_HappyPath(t *testing.T) {
	expected := &domain.AssetAttackPathsResponse{
		Asset: domain.AssetRef{ID: 1, Name: "TestAsset"},
		Aggregate: domain.AssetAggregate{
			WMax: 0.7, Level: "high", ThreatCount: 1, UncoveredCount: 0,
		},
		Paths: []domain.AttackPath{},
	}
	h := NewRiskHandler(&stubRiskSvc{resp: expected})

	app := fiber.New()
	app.Get("/api/risk/asset/:asset_id/attack-paths", h.assetAttackPaths)

	req := httptest.NewRequest("GET", "/api/risk/asset/1/attack-paths", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}
	if res.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	var got domain.AssetAttackPathsResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Asset.Name != "TestAsset" {
		t.Errorf("asset.name = %q, want TestAsset", got.Asset.Name)
	}
	if got.Aggregate.WMax != 0.7 {
		t.Errorf("aggregate.w_max = %v, want 0.7", got.Aggregate.WMax)
	}
	if got.Aggregate.Level != "high" {
		t.Errorf("aggregate.level = %q, want high", got.Aggregate.Level)
	}
}

func TestAssetAttackPaths_InvalidID(t *testing.T) {
	h := NewRiskHandler(&stubRiskSvc{})
	app := fiber.New()
	app.Get("/api/risk/asset/:asset_id/attack-paths", h.assetAttackPaths)

	req := httptest.NewRequest("GET", "/api/risk/asset/abc/attack-paths", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request: %v", err)
	}
	if res.StatusCode != 400 {
		t.Errorf("status = %d, want 400", res.StatusCode)
	}
}
