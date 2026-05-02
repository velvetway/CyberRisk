# Risk Graph Visualization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Реализовать переработанную страницу `/risk/graph/:assetId` со stacked layout (hero / threat list / Sankey-граф `S→ST→VL→C→DA`), client-side what-if симуляцией контролей и новым bulk endpoint'ом.

**Architecture:** Бэкенд — добавить bulk endpoint, агрегатный helper и расширить `Service`. Фронтенд — выделить чистую логику в `riskFlow.ts` (TDD), затем шесть новых компонентов и переписанный orchestrator. Формула PTSZI и таблицы БД не меняются.

**Tech Stack:** Go 1.25 / Fiber v2 / pgx v5 / PostgreSQL · React 19 / TypeScript / d3-sankey / framer-motion · Jest (через react-scripts) для фронтенд-тестов · стандартный `testing` для Go.

**Spec:** `docs/superpowers/specs/2026-05-02-risk-graph-visualization-design.md`

**Module name (Go):** `Diplom`. Все импорты вида `Diplom/internal/...`.

**Frontend test command (single-run, без watch):** `CI=true npm test`. Запускать из `frontend/`.

**Backend test command:** `go test ./internal/...`

---

## File Structure

**NEW (backend):**
```
internal/service/risk/aggregate.go            // ComputeAssetAggregate helper
internal/service/risk/aggregate_test.go       // unit tests for aggregate
internal/transport/http/risk_handlers_test.go // handler tests with stubbed Service
```

**NEW (frontend):**
```
frontend/src/lib/riskFlow.ts                          // pure logic (sankey graph builder + recomputeW)
frontend/src/lib/riskFlow.test.ts                     // jest tests for riskFlow
frontend/src/components/risk/AssetRiskHero.tsx        // top hero card
frontend/src/components/risk/ThreatList.tsx           // sortable list of all threats
frontend/src/components/risk/AttackFlowSankey.tsx     // d3-sankey + custom React nodes
frontend/src/components/risk/ControlPopover.tsx       // popover for a control node
frontend/src/components/risk/WhatIfBar.tsx            // sticky simulation banner
frontend/src/components/risk/PtsziBreakdown.tsx       // PTSZI formula + breakdown panel
```

**MODIFIED:**
```
internal/domain/risk_graph.go                  // + AssetAggregate, AssetAttackPathsResponse
internal/service/risk/service.go               // + AssembleAssetAttackPaths method
internal/transport/http/risk_handlers.go       // + assetAttackPaths handler
internal/transport/http/server.go              // register new route
frontend/src/types/riskGraph.ts                // + AssetAggregate, AssetAttackPathsResponse
frontend/src/pages/RiskGraphPage.tsx           // полностью переписан (orchestrator)
```

**DELETED:**
```
frontend/src/components/RiskGraphSankey.tsx    // unused legacy
```

---

## Phase A · Backend

### Task 1: Add domain types `AssetAggregate` and `AssetAttackPathsResponse`

**Files:**
- Modify: `internal/domain/risk_graph.go`

- [ ] **Step 1: Append the new types to the existing file**

Add at the end of `internal/domain/risk_graph.go`:

```go
// AssetAggregate — сводные метрики по всем угрозам одного актива.
type AssetAggregate struct {
	WMax           float64 `json:"w_max"`
	Level          string  `json:"level"`
	ThreatCount    int     `json:"threat_count"`
	UncoveredCount int     `json:"uncovered_count"`
}

// AssetAttackPathsResponse — ответ bulk-эндпоинта /api/risk/asset/:asset_id/attack-paths.
type AssetAttackPathsResponse struct {
	Asset     AssetRef       `json:"asset"`
	Aggregate AssetAggregate `json:"aggregate"`
	Paths     []AttackPath   `json:"paths"`
}
```

- [ ] **Step 2: Verify the package still compiles**

Run: `go build ./internal/domain/...`
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add internal/domain/risk_graph.go
git commit -m "$(cat <<'EOF'
feat(domain): add AssetAggregate and AssetAttackPathsResponse types

Types for the new bulk attack-paths endpoint that returns all threats
of a single asset in one request.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Implement `ComputeAssetAggregate` (TDD)

**Files:**
- Create: `internal/service/risk/aggregate.go`
- Create: `internal/service/risk/aggregate_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/service/risk/aggregate_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/service/risk/ -run ComputeAssetAggregate -v`
Expected: FAIL with `undefined: ComputeAssetAggregate`.

- [ ] **Step 3: Write the implementation**

Create `internal/service/risk/aggregate.go`:

```go
package risk

import "Diplom/internal/domain"

// ComputeAssetAggregate возвращает сводные метрики по всем рассчитанным AttackPath
// одного актива:
//   - WMax — максимум path.W среди всех путей.
//   - Level — уровень соответствующий WMax (через LevelFromW); для пустого набора — "info".
//   - ThreatCount — длина среза.
//   - UncoveredCount — число путей, в которых есть хотя бы одна VL с Uncovered=true.
func ComputeAssetAggregate(paths []domain.AttackPath) domain.AssetAggregate {
	agg := domain.AssetAggregate{ThreatCount: len(paths)}
	if len(paths) == 0 {
		agg.Level = "info"
		return agg
	}
	for _, p := range paths {
		if p.W > agg.WMax {
			agg.WMax = p.W
		}
		for _, vl := range p.VulnerableLinks {
			if vl.Uncovered {
				agg.UncoveredCount++
				break
			}
		}
	}
	agg.Level = LevelFromW(agg.WMax)
	return agg
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/service/risk/ -run ComputeAssetAggregate -v`
Expected: PASS for all three subtests.

- [ ] **Step 5: Commit**

```bash
git add internal/service/risk/aggregate.go internal/service/risk/aggregate_test.go
git commit -m "$(cat <<'EOF'
feat(risk): add ComputeAssetAggregate helper

Pure function that summarizes a slice of AttackPath into WMax, Level,
ThreatCount and UncoveredCount for the asset-level hero block.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Extend `Service` interface with `AssembleAssetAttackPaths`

**Files:**
- Modify: `internal/service/risk/service.go`

- [ ] **Step 1: Add the method to the Service interface**

In `internal/service/risk/service.go`, find the `Service` interface (around line 13–26) and add this line **after** `AssembleAttackPath`:

```go
	// Bulk: собрать AttackPath для всех релевантных угроз актива + агрегат.
	AssembleAssetAttackPaths(ctx context.Context, assetID int64) (*domain.AssetAttackPathsResponse, error)
```

- [ ] **Step 2: Implement the method on `service`**

In the same file, append after the existing `AssembleAttackPath` implementation:

```go
// AssembleAssetAttackPaths — собирает AttackPath для всех угроз из репозитория
// для одного актива, отбрасывает полностью пустые (без S, VL и DA), считает
// агрегатные метрики и возвращает единый ответ.
func (s *service) AssembleAssetAttackPaths(ctx context.Context, assetID int64) (*domain.AssetAttackPathsResponse, error) {
	if assetID <= 0 {
		return nil, fmt.Errorf("assetID must be positive")
	}

	asset, err := s.assetsRepo.GetByID(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("get asset: %w", err)
	}
	if asset == nil {
		return nil, fmt.Errorf("asset not found")
	}

	threats, err := s.threatsRepo.List(ctx, repository.ThreatFilter{})
	if err != nil {
		return nil, fmt.Errorf("list threats: %w", err)
	}

	paths := make([]domain.AttackPath, 0, len(threats))
	for _, t := range threats {
		path, err := s.AssembleAttackPath(ctx, assetID, t.ID)
		if err != nil {
			// одна угроза не собралась — не валим весь ответ
			continue
		}
		// Пропускаем абсолютно пустые пути: ни источников, ни VL, ни DA.
		if len(path.Sources) == 0 && len(path.VulnerableLinks) == 0 && len(path.DestructiveActions) == 0 {
			continue
		}
		paths = append(paths, *path)
	}

	return &domain.AssetAttackPathsResponse{
		Asset:     domain.AssetRef{ID: asset.ID, Name: asset.Name},
		Aggregate: ComputeAssetAggregate(paths),
		Paths:     paths,
	}, nil
}
```

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/service/risk/`
Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add internal/service/risk/service.go
git commit -m "$(cat <<'EOF'
feat(risk): add AssembleAssetAttackPaths bulk method

Returns all relevant attack paths plus an aggregate (WMax, Level, counts)
for a single asset in one call. Skips paths with no sources, vulnerable
links and destructive actions to keep the response noise-free.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Add HTTP handler + route + handler test

**Files:**
- Modify: `internal/transport/http/risk_handlers.go`
- Modify: `internal/transport/http/server.go`
- Create: `internal/transport/http/risk_handlers_test.go`

- [ ] **Step 1: Add the handler method**

In `internal/transport/http/risk_handlers.go`, append after `riskGraph` (currently the last method before `listThreatSources`):

```go
// GET /api/risk/asset/:asset_id/attack-paths
func (h *RiskHandler) assetAttackPaths(c *fiber.Ctx) error {
	assetID, err := strconv.ParseInt(c.Params("asset_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid asset_id"})
	}

	res, err := h.svc.AssembleAssetAttackPaths(c.Context(), assetID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}
```

- [ ] **Step 2: Register the route**

In `internal/transport/http/server.go`, find line 148:

```go
	readOnly.Get("/risk/graph/:asset_id/:threat_id", riskHandler.riskGraph)
```

Insert **immediately after it** (so it sits between line 148 and the existing line 149):

```go
	readOnly.Get("/risk/asset/:asset_id/attack-paths", riskHandler.assetAttackPaths)
```

- [ ] **Step 3: Verify the server builds**

Run: `go build ./...`
Expected: no output.

- [ ] **Step 4: Write handler tests**

Create `internal/transport/http/risk_handlers_test.go`:

```go
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
```

- [ ] **Step 5: Run handler tests**

Run: `go test ./internal/transport/http/ -run TestAssetAttackPaths -v`
Expected: PASS for both `TestAssetAttackPaths_HappyPath` and `TestAssetAttackPaths_InvalidID`.

- [ ] **Step 6: Run the full backend test suite**

Run: `go test ./internal/...`
Expected: PASS (no regressions).

- [ ] **Step 7: Commit**

```bash
git add internal/transport/http/risk_handlers.go \
        internal/transport/http/server.go \
        internal/transport/http/risk_handlers_test.go
git commit -m "$(cat <<'EOF'
feat(api): GET /api/risk/asset/:id/attack-paths

Bulk endpoint returning all attack paths and aggregate metrics for an
asset in one request, plus handler unit tests using a stubbed Service.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Phase B · Frontend logic

### Task 5: Add frontend types

**Files:**
- Modify: `frontend/src/types/riskGraph.ts`

- [ ] **Step 1: Append the new interfaces**

Append at the end of `frontend/src/types/riskGraph.ts`:

```ts
export interface AssetAggregate {
  w_max: number;
  level: 'info' | 'low' | 'medium' | 'high' | 'critical';
  threat_count: number;
  uncovered_count: number;
}

export interface AssetAttackPathsResponse {
  asset: { id: number; name: string };
  aggregate: AssetAggregate;
  paths: AttackPath[];
}
```

- [ ] **Step 2: Verify TypeScript still compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/types/riskGraph.ts
git commit -m "$(cat <<'EOF'
feat(types): AssetAggregate and AssetAttackPathsResponse

Frontend types matching the new backend bulk endpoint.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Implement `riskFlow.ts` (TDD)

**Files:**
- Create: `frontend/src/lib/riskFlow.ts`
- Create: `frontend/src/lib/riskFlow.test.ts`

- [ ] **Step 1: Write the failing test file**

Create `frontend/src/lib/riskFlow.test.ts`:

```ts
import { buildSankeyGraph, recomputeW } from './riskFlow';
import type { AttackPath } from '../types/riskGraph';

const minimalPath = (overrides: Partial<AttackPath> = {}): AttackPath => ({
  asset: { id: 1, name: 'A' },
  threat: { id: 10, name: 'T' },
  sources: [{ id: 1, code: 'S1', name: 'Внешний нарушитель' }],
  vulnerable_links: [],
  destructive_actions: [{
    id: 1, code: 'DA1', name: 'Утечка',
    affects_confidentiality: true, affects_integrity: false, affects_availability: false,
  }],
  w: 0.5,
  q_threat: 0.5,
  q_severity: 0.5,
  q_reaction: 0,
  z: 1.0,
  level: 'high',
  ...overrides,
});

describe('buildSankeyGraph', () => {
  test('one VL with one half-coverage control: per-VL flow conserves', () => {
    const path = minimalPath({
      vulnerable_links: [{
        vulnerability_id: 1, name: 'CVE-1', severity: 5,
        coverage_controls: [{ id: 100, name: 'WAF', coverage: 0.5 }],
        uncovered: false,
      }],
    });
    const g = buildSankeyGraph(path);
    const vlIn = g.links.filter(l => l.target === 'VL1').reduce((a, l) => a + l.value, 0);
    const vlOut = g.links.filter(l => l.source === 'VL1').reduce((a, l) => a + l.value, 0);
    expect(Math.abs(vlIn - vlOut)).toBeLessThan(1e-9);
  });

  test('two controls with combined coverage > 1 are clamped, all flow blocked', () => {
    const path = minimalPath({
      vulnerable_links: [{
        vulnerability_id: 1, name: 'CVE', severity: 1,
        coverage_controls: [
          { id: 1, name: 'A', coverage: 0.7 },
          { id: 2, name: 'B', coverage: 0.6 },
        ],
        uncovered: false,
      }],
    });
    const g = buildSankeyGraph(path);
    const cFlow = g.links.filter(l => l.kind === 'VL->C').reduce((a, l) => a + l.value, 0);
    const daFlow = g.links.filter(l => l.kind === 'VL->DA').reduce((a, l) => a + l.value, 0);
    expect(cFlow + daFlow).toBeCloseTo(1, 9);
    expect(daFlow).toBeCloseTo(0, 9);
  });

  test('VL with no controls — all flow goes to DA', () => {
    const path = minimalPath({
      vulnerable_links: [{
        vulnerability_id: 1, name: 'CVE', severity: 1,
        coverage_controls: [], uncovered: true,
      }],
    });
    const g = buildSankeyGraph(path);
    const cFlow = g.links.filter(l => l.kind === 'VL->C').length;
    const daFlow = g.links.filter(l => l.kind === 'VL->DA').reduce((a, l) => a + l.value, 0);
    expect(cFlow).toBe(0);
    expect(daFlow).toBeCloseTo(1, 9);
  });

  test('disabled control does not produce VL->C edges; flow shifts to DA', () => {
    const path = minimalPath({
      vulnerable_links: [{
        vulnerability_id: 1, name: 'CVE', severity: 1,
        coverage_controls: [{ id: 1, name: 'A', coverage: 1.0 }],
        uncovered: false,
      }],
    });
    const g = buildSankeyGraph(path, new Set([1]));
    expect(g.links.filter(l => l.kind === 'VL->C').length).toBe(0);
    const daFlow = g.links.filter(l => l.kind === 'VL->DA').reduce((a, l) => a + l.value, 0);
    expect(daFlow).toBeCloseTo(1, 9);
  });

  test('S->ST equal split when multiple sources', () => {
    const path = minimalPath({
      sources: [
        { id: 1, code: 'S1', name: 'Внешний' },
        { id: 2, code: 'S2', name: 'Внутр.' },
      ],
    });
    const g = buildSankeyGraph(path);
    const stIn = g.links.filter(l => l.target === 'ST').map(l => l.value);
    expect(stIn).toHaveLength(2);
    expect(stIn[0]).toBeCloseTo(0.5, 9);
    expect(stIn[1]).toBeCloseTo(0.5, 9);
  });

  test('ST->VL split by relative severity sums to 1', () => {
    const path = minimalPath({
      vulnerable_links: [
        { vulnerability_id: 1, name: 'V1', severity: 8, coverage_controls: [], uncovered: true },
        { vulnerability_id: 2, name: 'V2', severity: 2, coverage_controls: [], uncovered: true },
      ],
    });
    const g = buildSankeyGraph(path);
    const total = g.links.filter(l => l.kind === 'ST->VL').reduce((a, l) => a + l.value, 0);
    expect(total).toBeCloseTo(1, 9);
    const v1 = g.links.find(l => l.target === 'VL1')!;
    expect(v1.value).toBeCloseTo(0.8, 9);
  });
});

describe('recomputeW', () => {
  test('baseline (no disabled): W matches formula', () => {
    const path = minimalPath({
      q_threat: 0.6, q_severity: 0.4, q_reaction: 1.0, z: 1.0,
      w: (0.6 + 0.4 + 0) / 3,
      vulnerable_links: [{
        vulnerability_id: 1, name: 'V', severity: 1,
        coverage_controls: [{ id: 1, name: 'A', coverage: 1.0 }],
        uncovered: false,
      }],
    });
    const r = recomputeW(path, new Set());
    expect(r.qReaction).toBeCloseTo(1.0, 9);
    expect(r.w).toBeCloseTo((0.6 + 0.4 + 0) / 3, 9);
    expect(Math.abs(r.delta)).toBeLessThan(1e-9);
  });

  test('disabling the only covering control drops q_reaction to 0', () => {
    const path = minimalPath({
      q_threat: 0.6, q_severity: 0.4, q_reaction: 1.0, z: 1.0, w: (0.6 + 0.4) / 3,
      vulnerable_links: [{
        vulnerability_id: 1, name: 'V', severity: 1,
        coverage_controls: [{ id: 1, name: 'A', coverage: 1.0 }],
        uncovered: false,
      }],
    });
    const r = recomputeW(path, new Set([1]));
    expect(r.qReaction).toBe(0);
    expect(r.w).toBeCloseTo((0.6 + 0.4 + 1.0) / 3, 9);
    expect(r.delta).toBeGreaterThan(0);
  });

  test('disabling one of two controls on same VL keeps it covered', () => {
    const path = minimalPath({
      q_threat: 0.5, q_severity: 0.5, q_reaction: 1.0, z: 1.0, w: (0.5 + 0.5) / 3,
      vulnerable_links: [{
        vulnerability_id: 1, name: 'V', severity: 1,
        coverage_controls: [
          { id: 1, name: 'A', coverage: 0.4 },
          { id: 2, name: 'B', coverage: 0.6 },
        ],
        uncovered: false,
      }],
    });
    const r = recomputeW(path, new Set([1])); // disable A; B (cov=0.6>0) still covers
    expect(r.qReaction).toBe(1.0);
    expect(Math.abs(r.delta)).toBeLessThan(1e-9);
  });

  test('empty VLs: q_reaction=0, W is full pass-through', () => {
    const path = minimalPath({
      q_threat: 0.4, q_severity: 0.3, q_reaction: 0, z: 1.0,
      vulnerable_links: [],
      w: (0.4 + 0.3 + 1) / 3,
    });
    const r = recomputeW(path, new Set());
    expect(r.qReaction).toBe(0);
    expect(r.w).toBeCloseTo((0.4 + 0.3 + 1) / 3, 9);
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run from `frontend/`: `CI=true npm test -- --testPathPattern=riskFlow`
Expected: FAIL with `Cannot find module './riskFlow'` (or similar — module does not exist).

- [ ] **Step 3: Write the implementation**

Create `frontend/src/lib/riskFlow.ts`:

```ts
import type { AttackPath, ControlCoverage } from '../types/riskGraph';

export type SankeyNodeKind = 'S' | 'ST' | 'VL' | 'C' | 'DA';
export type SankeyLinkKind = 'S->ST' | 'ST->VL' | 'VL->C' | 'VL->DA';

export interface SankeyNodeData {
  id: string;
  kind: SankeyNodeKind;
  label: string;
  meta?: {
    coverage?: number;
    disabled?: boolean;
    severity?: number;
    uncovered?: boolean;
    code?: string;
  };
}

export interface SankeyLinkData {
  source: string;
  target: string;
  value: number;
  kind: SankeyLinkKind;
}

export interface SankeyGraph {
  nodes: SankeyNodeData[];
  links: SankeyLinkData[];
}

export interface RecomputeResult {
  qReaction: number;
  w: number;
  delta: number; // w − path.w; положительная дельта = риск вырос
}

const clamp01 = (v: number): number => (v < 0 ? 0 : v > 1 ? 1 : v);
const clampZ = (z: number): number => (z < 0.5 ? 0.5 : z > 1 ? 1 : z);

/**
 * Строит Sankey-граф S → ST → VL → {C | DA} с честной flow-семантикой.
 * Поток нормирован на 1.0. Подробное описание формул — в spec'е.
 */
export function buildSankeyGraph(
  path: AttackPath,
  disabledControlIds: Set<number> = new Set(),
): SankeyGraph {
  const nodes: SankeyNodeData[] = [];
  const links: SankeyLinkData[] = [];

  // S nodes
  for (const s of path.sources) {
    nodes.push({ id: `S${s.id}`, kind: 'S', label: s.code, meta: { code: s.code } });
  }

  // ST node (single)
  nodes.push({
    id: 'ST',
    kind: 'ST',
    label: path.threat.name,
    meta: { code: path.threat.bdu_id },
  });

  // VL nodes
  for (const vl of path.vulnerable_links) {
    nodes.push({
      id: `VL${vl.vulnerability_id}`,
      kind: 'VL',
      label: vl.name,
      meta: { severity: vl.severity, uncovered: vl.uncovered },
    });
  }

  // C nodes — дедуп по id (один контроль может покрывать несколько VL)
  const allControls = new Map<number, ControlCoverage>();
  for (const vl of path.vulnerable_links) {
    for (const c of vl.coverage_controls) {
      if (!allControls.has(c.id)) allControls.set(c.id, c);
    }
  }
  for (const c of allControls.values()) {
    nodes.push({
      id: `C${c.id}`,
      kind: 'C',
      label: c.name,
      meta: { coverage: c.coverage, disabled: disabledControlIds.has(c.id) },
    });
  }

  // DA nodes
  for (const da of path.destructive_actions) {
    nodes.push({ id: `DA${da.id}`, kind: 'DA', label: da.name, meta: { code: da.code } });
  }

  // S → ST: равномерное распределение
  if (path.sources.length > 0) {
    const w = 1 / path.sources.length;
    for (const s of path.sources) {
      links.push({ source: `S${s.id}`, target: 'ST', value: w, kind: 'S->ST' });
    }
  }

  // ST → VL: по относительной severity (сумма = 1)
  const totalSev = path.vulnerable_links.reduce((a, v) => a + (v.severity || 1), 0);
  if (path.vulnerable_links.length > 0) {
    for (const vl of path.vulnerable_links) {
      const sev = vl.severity || 1;
      const value = totalSev > 0 ? sev / totalSev : 1 / path.vulnerable_links.length;
      links.push({
        source: 'ST',
        target: `VL${vl.vulnerability_id}`,
        value,
        kind: 'ST->VL',
      });
    }
  }

  // Per-VL: VL → C (только enabled), VL → DA (passthrough)
  for (const vl of path.vulnerable_links) {
    const inflow = totalSev > 0
      ? (vl.severity || 1) / totalSev
      : 1 / Math.max(1, path.vulnerable_links.length);
    const enabled = vl.coverage_controls.filter(c => !disabledControlIds.has(c.id));
    const sumC = enabled.reduce((a, c) => a + c.coverage, 0);
    const cov = Math.min(1, sumC);

    if (sumC > 0) {
      for (const c of enabled) {
        const share = c.coverage / sumC;
        const value = share * cov * inflow;
        if (value > 0) {
          links.push({
            source: `VL${vl.vulnerability_id}`,
            target: `C${c.id}`,
            value,
            kind: 'VL->C',
          });
        }
      }
    }

    const passthrough = inflow * (1 - cov);
    if (passthrough > 0 && path.destructive_actions.length > 0) {
      const perDa = passthrough / path.destructive_actions.length;
      for (const da of path.destructive_actions) {
        links.push({
          source: `VL${vl.vulnerability_id}`,
          target: `DA${da.id}`,
          value: perDa,
          kind: 'VL->DA',
        });
      }
    }
  }

  return { nodes, links };
}

/**
 * Пересчитывает W при отключении набора контролей. Зеркалит бэкенд-формулу
 * QReactionFromVLs: VL считается «закрытой», если у неё есть хотя бы один
 * включённый контроль с coverage > 0.
 */
export function recomputeW(
  path: AttackPath,
  disabledControlIds: Set<number>,
): RecomputeResult {
  const vls = path.vulnerable_links;
  let qReaction = 0;
  if (vls.length > 0) {
    let covered = 0;
    for (const vl of vls) {
      const hasActive = vl.coverage_controls.some(
        c => c.coverage > 0 && !disabledControlIds.has(c.id),
      );
      if (hasActive) covered++;
    }
    qReaction = covered / vls.length;
  }

  const qT = clamp01(path.q_threat);
  const qS = clamp01(path.q_severity);
  const qR = clamp01(qReaction);
  const z = clampZ(path.z);
  const w = ((qT + qS + (1 - qR)) / 3) * z;

  return {
    qReaction,
    w,
    delta: w - path.w,
  };
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run from `frontend/`: `CI=true npm test -- --testPathPattern=riskFlow`
Expected: PASS for all 10 tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/riskFlow.ts frontend/src/lib/riskFlow.test.ts
git commit -m "$(cat <<'EOF'
feat(frontend): pure riskFlow lib (Sankey builder + recomputeW)

Pure functions for the new RiskGraphPage:
- buildSankeyGraph: S→ST→VL→{C|DA} graph with normalized flow
- recomputeW: client-side what-if mirroring backend QReactionFromVLs

10 unit tests cover flow conservation, coverage clamping, disabled
controls, severity-weighted splits and edge cases.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Phase C · Frontend components

> All components live under `frontend/src/components/risk/`. Style via existing design tokens (`var(--risk-*)`, `var(--bg-elev-*)`, `var(--font-mono)`, etc.) — same approach as the current `RiskGraphPage`.

### Task 7: `AssetRiskHero`

**Files:**
- Create: `frontend/src/components/risk/AssetRiskHero.tsx`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/risk/AssetRiskHero.tsx`:

```tsx
import React from "react";
import { motion } from "framer-motion";
import { Btn, Card, Icon, RiskBadge } from "../design";
import type { AssetAggregate } from "../../types/riskGraph";

type Level = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface AssetRiskHeroProps {
  assetId: number;
  assetName: string;
  aggregate: AssetAggregate;
  onPdf?: () => void;
  onParams?: () => void;
  onBack?: () => void;
}

export const AssetRiskHero: React.FC<AssetRiskHeroProps> = ({
  assetId, assetName, aggregate, onPdf, onParams, onBack,
}) => {
  const level = aggregate.level as Level;

  return (
    <motion.div
      initial={{ opacity: 0, y: -4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <Card pad={0}>
        <div style={{
          padding: '16px 20px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 16,
        }}>
          <div>
            <div style={{
              fontSize: 'var(--text-xs)', color: 'var(--fg-dim)',
              fontFamily: 'var(--font-mono)',
              textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4,
            }}>
              Граф атаки · ПТСЗИ
            </div>
            <h1 style={{
              margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600,
              letterSpacing: '-0.02em',
            }}>
              <span className="mono" style={{ color: 'var(--fg-muted)' }}>
                A-{String(assetId).padStart(3, '0')}
              </span>{' '}
              «{assetName}»
            </h1>
            <div style={{
              marginTop: 8, display: 'flex', gap: 16, alignItems: 'center',
              fontSize: 'var(--text-sm)', color: 'var(--fg-muted)',
            }}>
              <span className="num" style={{ color: 'var(--fg)', fontWeight: 600 }}>
                W = {aggregate.w_max.toFixed(2)}
              </span>
              <RiskBadge level={level} />
              <span>{aggregate.threat_count} угроз</span>
              <span style={{ color: aggregate.uncovered_count > 0 ? 'var(--risk-critical)' : 'var(--fg-muted)' }}>
                {aggregate.uncovered_count} непокрыто
              </span>
            </div>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            {onBack && (
              <Btn variant="outline" icon={<Icon name="arrowL" size={13} />} onClick={onBack}>
                Назад
              </Btn>
            )}
            {onParams && (
              <Btn variant="outline" icon={<Icon name="sliders" size={13} />} onClick={onParams}>
                Параметры
              </Btn>
            )}
            {onPdf && (
              <Btn variant="primary" icon={<Icon name="file" size={13} />} onClick={onPdf}>
                PDF сценария
              </Btn>
            )}
          </div>
        </div>
      </Card>
    </motion.div>
  );
};
```

- [ ] **Step 2: Verify it compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/risk/AssetRiskHero.tsx
git commit -m "$(cat <<'EOF'
feat(frontend): AssetRiskHero component

Top hero card showing asset id, name, aggregate W, level badge, threat
count and uncovered count, with PDF / Параметры / Back action buttons.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: `ThreatList`

**Files:**
- Create: `frontend/src/components/risk/ThreatList.tsx`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/risk/ThreatList.tsx`:

```tsx
import React, { useMemo, useState } from "react";
import { Card, RiskBadge } from "../design";
import type { AttackPath } from "../../types/riskGraph";

type SortKey = 'w' | 'uncovered' | 'name';
type Level = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface ThreatListProps {
  paths: AttackPath[];
  selectedThreatId: number | null;
  onSelect: (threatId: number) => void;
}

export const ThreatList: React.FC<ThreatListProps> = ({ paths, selectedThreatId, onSelect }) => {
  const [sortKey, setSortKey] = useState<SortKey>('w');

  const sorted = useMemo(() => {
    const copy = [...paths];
    copy.sort((a, b) => {
      switch (sortKey) {
        case 'w':
          return b.w - a.w;
        case 'uncovered': {
          const au = a.vulnerable_links.filter(v => v.uncovered).length;
          const bu = b.vulnerable_links.filter(v => v.uncovered).length;
          return bu - au;
        }
        case 'name':
          return a.threat.name.localeCompare(b.threat.name, 'ru');
      }
    });
    return copy;
  }, [paths, sortKey]);

  if (paths.length === 0) {
    return (
      <Card title="Угрозы актива" dense>
        <div style={{ padding: 12, color: 'var(--fg-muted)' }}>
          У этого актива нет связанных угроз с непустыми путями.
        </div>
      </Card>
    );
  }

  return (
    <Card pad={0}>
      <div style={{
        padding: '12px 16px',
        borderBottom: '1px solid var(--border)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
      }}>
        <div style={{
          fontSize: 'var(--text-xs)', color: 'var(--fg-dim)',
          textTransform: 'uppercase', letterSpacing: '0.08em', fontWeight: 500,
        }}>
          Угрозы актива · {paths.length}
        </div>
        <div style={{ display: 'flex', gap: 6, fontSize: 'var(--text-xs)' }}>
          {(['w', 'uncovered', 'name'] as SortKey[]).map(k => (
            <button
              key={k}
              onClick={() => setSortKey(k)}
              style={{
                padding: '4px 8px',
                background: sortKey === k ? 'var(--bg-elev-2)' : 'transparent',
                border: '1px solid var(--border)',
                borderRadius: 'var(--r-sm)',
                color: sortKey === k ? 'var(--fg)' : 'var(--fg-muted)',
                cursor: 'pointer',
                fontFamily: 'var(--font-mono)',
              }}
            >
              {k === 'w' ? 'W ↓' : k === 'uncovered' ? 'непокрыт.' : 'имя'}
            </button>
          ))}
        </div>
      </div>

      <div role="listbox" aria-label="Угрозы актива">
        {sorted.map(p => {
          const isSelected = p.threat.id === selectedThreatId;
          const uncoveredCount = p.vulnerable_links.filter(v => v.uncovered).length;
          const level = p.level as Level;

          return (
            <div
              key={p.threat.id}
              role="option"
              aria-selected={isSelected}
              tabIndex={0}
              onClick={() => onSelect(p.threat.id)}
              onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(p.threat.id); } }}
              style={{
                display: 'grid',
                gridTemplateColumns: '24px 110px 1fr 60px 80px 64px',
                gap: 12,
                alignItems: 'center',
                padding: '10px 16px',
                borderBottom: '1px solid var(--border)',
                background: isSelected ? 'var(--bg-elev-2)' : 'transparent',
                cursor: 'pointer',
                borderLeft: `3px solid ${isSelected ? 'var(--accent)' : 'transparent'}`,
              }}
            >
              <span style={{ color: isSelected ? 'var(--accent)' : 'var(--fg-faint)' }}>{isSelected ? '▶' : ''}</span>
              <span className="mono" style={{
                fontSize: 'var(--text-xs)',
                color: 'var(--fg-muted)',
                fontFamily: 'var(--font-mono)',
              }}>
                {p.threat.bdu_id || `T-${p.threat.id}`}
              </span>
              <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {p.threat.name}
              </span>
              <div style={{
                height: 6, background: 'var(--bg-elev-3)', borderRadius: 999, overflow: 'hidden',
              }}>
                <div style={{
                  width: `${Math.round(p.w * 100)}%`,
                  height: '100%',
                  background: `var(--risk-${level})`,
                  transition: 'width 200ms',
                }} />
              </div>
              <span className="num" style={{
                fontSize: 'var(--text-sm)',
                fontFamily: 'var(--font-mono)',
                color: 'var(--fg)',
                textAlign: 'right',
              }}>
                {p.w.toFixed(2)}
              </span>
              <span style={{ textAlign: 'right' }}>
                {uncoveredCount > 0 ? (
                  <span className="mono" style={{
                    fontSize: 'var(--text-xs)',
                    color: 'var(--risk-critical)',
                    fontFamily: 'var(--font-mono)',
                  }}>
                    {uncoveredCount} откр.
                  </span>
                ) : (
                  <RiskBadge level={level} />
                )}
              </span>
            </div>
          );
        })}
      </div>
    </Card>
  );
};
```

- [ ] **Step 2: Verify it compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/risk/ThreatList.tsx
git commit -m "$(cat <<'EOF'
feat(frontend): ThreatList component

Sortable list of all threats for an asset with W bars, uncovered count
and selection highlighting. Sortable by W (default), uncovered count
or name. Keyboard-navigable.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 9: `PtsziBreakdown` (extract from current page)

**Files:**
- Create: `frontend/src/components/risk/PtsziBreakdown.tsx`

- [ ] **Step 1: Create the component (extracted from existing RiskGraphPage)**

Create `frontend/src/components/risk/PtsziBreakdown.tsx`:

```tsx
import React from "react";
import { Card, PtsziFormula, RiskBadge } from "../design";
import type { AttackPath } from "../../types/riskGraph";

type Level = 'critical' | 'high' | 'medium' | 'low' | 'info';

const levelFromW = (w: number): Level => {
  if (w >= 0.75) return 'critical';
  if (w >= 0.50) return 'high';
  if (w >= 0.25) return 'medium';
  return 'low';
};

export interface PtsziBreakdownProps {
  path: AttackPath;
  /** Optional what-if simulation overrides for q_reaction and w. */
  simulated?: { qReaction: number; w: number; delta: number };
}

export const PtsziBreakdown: React.FC<PtsziBreakdownProps> = ({ path, simulated }) => {
  const effectiveQR = simulated?.qReaction ?? path.q_reaction;
  const effectiveW = simulated?.w ?? path.w;
  const level = levelFromW(effectiveW);

  const rows = [
    { label: 'Q^th · потенциал угрозы', v: path.q_threat, info: path.threat.name, color: 'var(--risk-critical)' },
    { label: 'q · опасность уязвимостей', v: path.q_severity, info: `${path.vulnerable_links.length} VL`, color: 'var(--risk-high)' },
    { label: 'Q^re · покрытие СЗИ', v: effectiveQR, info: 'Ниже = риск выше (1−Q^re)', color: 'var(--risk-info)' },
    { label: 'Z · вес контура', v: path.z, info: 'По окружению актива', color: 'var(--risk-medium)' },
  ];

  return (
    <Card title="Формула ПТСЗИ" subtitle="Выбранный сценарий" dense>
      <div style={{ marginBottom: 14 }}>
        <PtsziFormula size="lg" align="center" />
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {rows.map(p => (
          <div key={p.label} style={{
            padding: '8px 10px', background: 'var(--bg-elev-2)',
            border: '1px solid var(--border)', borderRadius: 'var(--r-sm)',
          }}>
            <div style={{
              display: 'flex', justifyContent: 'space-between', alignItems: 'baseline',
              marginBottom: 4,
            }}>
              <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)' }}>{p.label}</span>
              <span className="num" style={{ fontSize: 'var(--text-sm)', color: p.color, fontWeight: 600 }}>
                {p.v.toFixed(2)}
              </span>
            </div>
            <div style={{
              height: 3, background: 'var(--bg-elev-3)', borderRadius: 999, overflow: 'hidden',
            }}>
              <div style={{
                width: `${Math.max(0, Math.min(1, p.v)) * 100}%`,
                height: '100%', background: p.color, transition: 'width 300ms',
              }} />
            </div>
            <div style={{ fontSize: 10, color: 'var(--fg-faint)', marginTop: 4 }}>{p.info}</div>
          </div>
        ))}
      </div>

      <div style={{
        marginTop: 14, padding: '12px 14px',
        background: 'var(--accent-ghost)',
        border: '1px solid oklch(0.62 var(--accent-c) var(--accent-h) / 0.35)',
        borderRadius: 'var(--r-md)',
      }}>
        <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', marginBottom: 4 }}>
          {simulated ? 'W (оценка симуляции, клиент-сайд)' : 'Итоговый вес сценария W'}
        </div>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
          <span className="num" style={{
            fontSize: 28, fontWeight: 600, color: 'var(--accent)',
            letterSpacing: '-0.02em',
          }}>
            {effectiveW.toFixed(3)}
          </span>
          <RiskBadge level={level} />
          {simulated && simulated.delta !== 0 && (
            <span className="num" style={{
              fontSize: 'var(--text-sm)',
              color: simulated.delta > 0 ? 'var(--risk-critical)' : 'var(--risk-low)',
              fontFamily: 'var(--font-mono)',
            }}>
              {simulated.delta > 0 ? '+' : ''}{simulated.delta.toFixed(2)}
            </span>
          )}
        </div>
        <div style={{
          fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', marginTop: 6,
          fontFamily: 'var(--font-mono)',
        }}>
          ({path.q_threat.toFixed(2)} + {path.q_severity.toFixed(2)} + {(1 - effectiveQR).toFixed(2)}) / 3 × {path.z.toFixed(2)}
        </div>
      </div>
    </Card>
  );
};
```

- [ ] **Step 2: Verify it compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/risk/PtsziBreakdown.tsx
git commit -m "$(cat <<'EOF'
feat(frontend): PtsziBreakdown component

Extracted formula breakdown panel with optional what-if simulation
overrides (effective q_reaction and W from client-side recompute).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 10: `AttackFlowSankey`

**Files:**
- Create: `frontend/src/components/risk/AttackFlowSankey.tsx`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/risk/AttackFlowSankey.tsx`:

```tsx
import React, { useMemo, useState } from "react";
import { sankey, sankeyLinkHorizontal, SankeyNode, SankeyLink } from "d3-sankey";
import { buildSankeyGraph, SankeyNodeData, SankeyLinkData, SankeyNodeKind } from "../../lib/riskFlow";
import type { AttackPath } from "../../types/riskGraph";

const WIDTH = 960;
const HEIGHT = 460;
const NODE_WIDTH = 18;
const NODE_PAD = 12;

type LayoutNode = SankeyNode<SankeyNodeData, SankeyLinkData>;
type LayoutLink = SankeyLink<SankeyNodeData, SankeyLinkData>;

const colorFor = (kind: SankeyNodeKind, dimmed: boolean): string => {
  if (dimmed) return 'var(--bg-elev-3)';
  switch (kind) {
    case 'S': return 'var(--risk-info)';
    case 'ST': return 'var(--risk-critical)';
    case 'VL': return 'var(--risk-high)';
    case 'C': return 'var(--risk-low, #58a463)';
    case 'DA': return 'var(--risk-critical)';
  }
};

const linkColor = (l: LayoutLink): string => {
  const k = (l as unknown as SankeyLinkData).kind;
  switch (k) {
    case 'S->ST': return 'var(--risk-info)';
    case 'ST->VL': return 'var(--risk-high)';
    case 'VL->C': return 'var(--risk-low, #58a463)';
    case 'VL->DA': return 'var(--risk-critical)';
  }
};

export interface AttackFlowSankeyProps {
  path: AttackPath;
  disabledControlIds: Set<number>;
  onControlClick: (controlId: number, anchor: { x: number; y: number }) => void;
}

export const AttackFlowSankey: React.FC<AttackFlowSankeyProps> = ({
  path, disabledControlIds, onControlClick,
}) => {
  const [hoverNode, setHoverNode] = useState<string | null>(null);

  const layout = useMemo(() => {
    const graph = buildSankeyGraph(path, disabledControlIds);
    const nodeIndex = new Map<string, number>();
    graph.nodes.forEach((n, i) => nodeIndex.set(n.id, i));

    const sk = sankey<SankeyNodeData, SankeyLinkData>()
      .nodeWidth(NODE_WIDTH)
      .nodePadding(NODE_PAD)
      .extent([[12, 28], [WIDTH - 12, HEIGHT - 12]])
      .nodeId(d => d.id);

    return sk({
      nodes: graph.nodes.map(n => ({ ...n })),
      links: graph.links.map(l => ({ ...l })),
    });
  }, [path, disabledControlIds]);

  const isHighlighted = (l: LayoutLink): boolean => {
    if (!hoverNode) return false;
    const s = typeof l.source === 'object' ? (l.source as LayoutNode).id : l.source;
    const t = typeof l.target === 'object' ? (l.target as LayoutNode).id : l.target;
    return s === hoverNode || t === hoverNode;
  };

  return (
    <div style={{ position: 'relative' }}>
      <div style={{
        padding: '8px 12px',
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-muted)',
        background: 'var(--bg-elev-2)',
        borderBottom: '1px solid var(--border)',
      }}>
        Толщина <code>VL→C</code> — coverage контроля. <code>q_reaction</code> в формуле — доля VL, прикрытых хотя бы одним контролем.
      </div>

      <svg
        viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
        style={{ width: '100%', height: 'auto', display: 'block', background: 'var(--bg-elev-1)' }}
      >
        {/* Column headers — approximate positions matching d3-sankey's even spread */}
        {([
          { label: 'Источник',         x: 24,  anchor: 'start'  as const },
          { label: 'Угроза (БДУ)',     x: 288, anchor: 'middle' as const },
          { label: 'Уязвимость',       x: 480, anchor: 'middle' as const },
          { label: 'СЗИ (контроль)',   x: 672, anchor: 'middle' as const },
          { label: 'Действие',         x: 936, anchor: 'end'    as const },
        ]).map(h => (
          <text
            key={h.label}
            x={h.x}
            y={18}
            fontSize={10}
            fill="var(--fg-faint)"
            textAnchor={h.anchor}
            fontFamily="var(--font-mono)"
            style={{ textTransform: 'uppercase', letterSpacing: '0.08em' }}
          >
            {h.label}
          </text>
        ))}

        {/* Links */}
        {(layout.links as LayoutLink[]).map((l, i) => {
          const linkData = l as unknown as SankeyLinkData;
          const path = sankeyLinkHorizontal()(l) ?? '';
          const highlighted = isHighlighted(l);
          const op = hoverNode ? (highlighted ? 0.85 : 0.15) : 0.5;
          return (
            <path
              key={`l-${i}`}
              d={path}
              fill="none"
              stroke={linkColor(l)}
              strokeOpacity={op}
              strokeWidth={Math.max(1, l.width ?? 1)}
              style={{ transition: 'stroke-opacity 200ms' }}
            >
              <title>{`${linkData.kind}: ${(l.value ?? 0).toFixed(3)} (${((l.value ?? 0) * 100).toFixed(1)}%)`}</title>
            </path>
          );
        })}

        {/* Nodes */}
        {(layout.nodes as LayoutNode[]).map(n => {
          const x0 = n.x0 ?? 0, x1 = n.x1 ?? 0, y0 = n.y0 ?? 0, y1 = n.y1 ?? 0;
          const w = x1 - x0, h = Math.max(2, y1 - y0);
          const dimmed = !!(n.meta?.disabled);
          const fill = colorFor(n.kind, dimmed);
          const labelLeft = x0 < WIDTH / 2;
          const isHover = hoverNode === n.id;
          const clickable = n.kind === 'C';

          return (
            <g
              key={n.id}
              tabIndex={clickable ? 0 : -1}
              role={clickable ? 'button' : undefined}
              aria-label={clickable ? `Контроль ${n.label}` : undefined}
              onMouseEnter={() => setHoverNode(n.id)}
              onMouseLeave={() => setHoverNode(null)}
              onClick={(e) => {
                if (clickable) {
                  const rect = (e.currentTarget.ownerSVGElement as SVGSVGElement).getBoundingClientRect();
                  const id = parseInt(n.id.replace('C', ''), 10);
                  onControlClick(id, { x: rect.left + (x0 + x1) / 2, y: rect.top + y0 });
                }
              }}
              onKeyDown={(e) => {
                if (clickable && (e.key === 'Enter' || e.key === ' ')) {
                  e.preventDefault();
                  const rect = (e.currentTarget.ownerSVGElement as SVGSVGElement).getBoundingClientRect();
                  const id = parseInt(n.id.replace('C', ''), 10);
                  onControlClick(id, { x: rect.left + (x0 + x1) / 2, y: rect.top + y0 });
                }
              }}
              style={{ cursor: clickable ? 'pointer' : 'default' }}
            >
              <rect
                x={x0} y={y0} width={w} height={h}
                fill={fill}
                opacity={dimmed ? 0.4 : 1}
                stroke={isHover ? 'var(--fg)' : 'transparent'}
                strokeWidth={1}
                rx={3}
              />
              <text
                x={labelLeft ? x1 + 6 : x0 - 6}
                y={(y0 + y1) / 2}
                dy="0.35em"
                textAnchor={labelLeft ? 'start' : 'end'}
                fontSize={11}
                fill={dimmed ? 'var(--fg-faint)' : 'var(--fg)'}
                style={{ pointerEvents: 'none' }}
              >
                {n.label.length > 28 ? n.label.slice(0, 27) + '…' : n.label}
              </text>
              {dimmed && (
                <text
                  x={labelLeft ? x1 + 6 : x0 - 6}
                  y={(y0 + y1) / 2 + 12}
                  textAnchor={labelLeft ? 'start' : 'end'}
                  fontSize={9}
                  fill="var(--fg-faint)"
                  style={{ pointerEvents: 'none' }}
                >
                  выкл в симуляции
                </text>
              )}
              <title>{`${n.kind}: ${n.label}`}</title>
            </g>
          );
        })}
      </svg>
    </div>
  );
};
```

- [ ] **Step 2: Verify it compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/risk/AttackFlowSankey.tsx
git commit -m "$(cat <<'EOF'
feat(frontend): AttackFlowSankey component

d3-sankey layout with custom React-rendered nodes for the 5-column
S→ST→VL→C→DA graph. Hover highlights, control nodes are clickable
(opens ControlPopover via onControlClick), disabled controls are
visually dimmed.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 11: `ControlPopover`

**Files:**
- Create: `frontend/src/components/risk/ControlPopover.tsx`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/risk/ControlPopover.tsx`:

```tsx
import React, { useEffect, useRef } from "react";
import { Btn } from "../design";
import type { ControlCoverage } from "../../types/riskGraph";

export interface ControlPopoverProps {
  control: ControlCoverage;
  disabled: boolean;
  anchor: { x: number; y: number };
  onToggle: (id: number) => void;
  onClose: () => void;
}

export const ControlPopover: React.FC<ControlPopoverProps> = ({
  control, disabled, anchor, onToggle, onClose,
}) => {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const onDocClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') onClose(); };
    document.addEventListener('mousedown', onDocClick);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDocClick);
      document.removeEventListener('keydown', onKey);
    };
  }, [onClose]);

  return (
    <div
      ref={ref}
      role="dialog"
      aria-label={`Контроль ${control.name}`}
      style={{
        position: 'fixed',
        left: anchor.x - 130,
        top: anchor.y - 140,
        width: 260,
        background: 'var(--bg-elev-2)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--r-md)',
        padding: 14,
        boxShadow: '0 8px 24px rgba(0,0,0,0.35)',
        zIndex: 1000,
        fontSize: 'var(--text-sm)',
      }}
    >
      <div style={{
        fontSize: 'var(--text-xs)',
        color: 'var(--fg-dim)',
        fontFamily: 'var(--font-mono)',
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
      }}>
        СЗИ · C-{control.id}
      </div>
      <div style={{
        marginTop: 6,
        fontWeight: 600,
        fontSize: 'var(--text-md)',
        lineHeight: 1.3,
      }}>
        {control.name}
      </div>
      <div style={{ marginTop: 12, display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
        <span style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)' }}>Coverage</span>
        <span className="num" style={{
          fontFamily: 'var(--font-mono)',
          fontSize: 'var(--text-md)',
          color: 'var(--risk-low, #58a463)',
        }}>
          {control.coverage.toFixed(2)}
        </span>
      </div>
      <div style={{
        marginTop: 4,
        height: 4,
        background: 'var(--bg-elev-3)',
        borderRadius: 999,
        overflow: 'hidden',
      }}>
        <div style={{
          width: `${Math.max(0, Math.min(1, control.coverage)) * 100}%`,
          height: '100%',
          background: 'var(--risk-low, #58a463)',
        }} />
      </div>
      <div style={{ marginTop: 14, display: 'flex', gap: 8 }}>
        <Btn
          variant={disabled ? 'primary' : 'outline'}
          onClick={() => onToggle(control.id)}
        >
          {disabled ? 'Включить' : 'Выкл в симуляции'}
        </Btn>
        <Btn variant="outline" onClick={onClose}>Закрыть</Btn>
      </div>
    </div>
  );
};
```

- [ ] **Step 2: Verify it compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/risk/ControlPopover.tsx
git commit -m "$(cat <<'EOF'
feat(frontend): ControlPopover component

Modal popover for a single control: shows id, name, coverage value,
and a toggle button to disable/re-enable in client-side simulation.
Click-outside and Escape close the popover.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 12: `WhatIfBar`

**Files:**
- Create: `frontend/src/components/risk/WhatIfBar.tsx`

- [ ] **Step 1: Create the component**

Create `frontend/src/components/risk/WhatIfBar.tsx`:

```tsx
import React from "react";
import { Btn } from "../design";

export interface WhatIfChip {
  id: number;
  name: string;
}

export interface WhatIfBarProps {
  disabledChips: WhatIfChip[];
  baselineW: number;
  simulatedW: number;
  delta: number;
  onReset: () => void;
  onRemoveChip: (id: number) => void;
  onSaveNote: () => void;
}

export const WhatIfBar: React.FC<WhatIfBarProps> = ({
  disabledChips, baselineW, simulatedW, delta, onReset, onRemoveChip, onSaveNote,
}) => {
  if (disabledChips.length === 0) return null;

  const sign = delta > 0 ? '+' : '';
  const deltaColor = delta > 0 ? 'var(--risk-critical)' : delta < 0 ? 'var(--risk-low, #58a463)' : 'var(--fg-muted)';

  return (
    <div
      role="status"
      aria-live="polite"
      style={{
        position: 'sticky',
        bottom: 12,
        marginTop: 12,
        padding: '12px 16px',
        background: 'var(--bg-elev-2)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--r-md)',
        boxShadow: '0 4px 16px rgba(0,0,0,0.25)',
        display: 'flex',
        alignItems: 'center',
        gap: 12,
        flexWrap: 'wrap',
      }}
    >
      <span style={{
        fontSize: 'var(--text-xs)',
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        color: 'var(--fg-dim)',
        fontFamily: 'var(--font-mono)',
      }}>
        Симуляция
      </span>

      <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
        {disabledChips.map(c => (
          <button
            key={c.id}
            onClick={() => onRemoveChip(c.id)}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 6,
              padding: '4px 8px',
              fontSize: 'var(--text-xs)',
              background: 'var(--bg-elev-3)',
              border: '1px solid var(--border)',
              borderRadius: 'var(--r-sm)',
              cursor: 'pointer',
              color: 'var(--fg)',
            }}
          >
            {c.name} <span aria-hidden="true">✕</span>
          </button>
        ))}
      </div>

      <span style={{ marginLeft: 'auto', fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>
        <span className="num" style={{ color: 'var(--fg-muted)' }}>{baselineW.toFixed(2)}</span>
        {' → '}
        <span className="num" style={{ color: 'var(--fg)', fontWeight: 600 }}>{simulatedW.toFixed(2)}</span>
        <span className="num" style={{ marginLeft: 8, color: deltaColor }}>
          ΔW {sign}{delta.toFixed(2)}
        </span>
      </span>

      <Btn variant="outline" onClick={onReset}>Сбросить</Btn>
      <Btn variant="outline" onClick={onSaveNote}>Сохранить заметку</Btn>
    </div>
  );
};
```

- [ ] **Step 2: Verify it compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/risk/WhatIfBar.tsx
git commit -m "$(cat <<'EOF'
feat(frontend): WhatIfBar component

Sticky bottom banner that surfaces an active what-if simulation:
chips for each disabled control (click to re-enable), baseline → simulated
W with ΔW, plus Сбросить and Сохранить заметку actions. Hidden when no
controls are disabled.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Phase D · Wire-up + cleanup

### Task 13: Rewrite `RiskGraphPage` as orchestrator

**Files:**
- Modify (full rewrite): `frontend/src/pages/RiskGraphPage.tsx`

- [ ] **Step 1: Replace the file contents**

Overwrite `frontend/src/pages/RiskGraphPage.tsx` with:

```tsx
import React, { useEffect, useMemo, useState } from "react";
import { useParams, useSearchParams, useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { authFetch } from "../api/client";
import { Btn, Card, Icon } from "../components/design";
import { AssetRiskHero } from "../components/risk/AssetRiskHero";
import { ThreatList } from "../components/risk/ThreatList";
import { AttackFlowSankey } from "../components/risk/AttackFlowSankey";
import { ControlPopover } from "../components/risk/ControlPopover";
import { WhatIfBar } from "../components/risk/WhatIfBar";
import { PtsziBreakdown } from "../components/risk/PtsziBreakdown";
import { recomputeW } from "../lib/riskFlow";
import type { AssetAttackPathsResponse, AttackPath, ControlCoverage } from "../types/riskGraph";

export const RiskGraphPage: React.FC = () => {
  const { assetId } = useParams<{ assetId: string }>();
  const [params, setParams] = useSearchParams();
  const navigate = useNavigate();

  const [response, setResponse] = useState<AssetAttackPathsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [disabledControls, setDisabledControls] = useState<Set<number>>(new Set());
  const [popoverFor, setPopoverFor] = useState<{ control: ControlCoverage; anchor: { x: number; y: number } } | null>(null);

  // Fetch all attack paths for the asset
  useEffect(() => {
    if (!assetId) return;
    setLoading(true);
    setError(null);
    authFetch(`/api/risk/asset/${assetId}/attack-paths`)
      .then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`)))
      .then((data: AssetAttackPathsResponse) => {
        setResponse(data);
        setLoading(false);
      })
      .catch(e => {
        setError(e.message);
        setLoading(false);
      });
  }, [assetId]);

  // Resolve selected threat from query or auto-pick max-W
  const requestedThreatId = params.get('threat');
  const selectedThreatId = useMemo<number | null>(() => {
    if (!response || response.paths.length === 0) return null;
    if (requestedThreatId) {
      const id = parseInt(requestedThreatId, 10);
      if (response.paths.some(p => p.threat.id === id)) return id;
      // requested threat is not in the asset's paths
      toast.error(`Угроза id=${id} не найдена для этого актива`);
      return response.paths[0].threat.id;
    }
    return response.paths.reduce((acc, p) => p.w > acc.w ? p : acc, response.paths[0]).threat.id;
  }, [response, requestedThreatId]);

  // Auto-sync ?threat= without push-state
  useEffect(() => {
    if (selectedThreatId !== null && String(selectedThreatId) !== requestedThreatId) {
      const next = new URLSearchParams(params);
      next.set('threat', String(selectedThreatId));
      setParams(next, { replace: true });
    }
  }, [selectedThreatId, requestedThreatId, params, setParams]);

  // Reset what-if when switching threats
  useEffect(() => {
    setDisabledControls(new Set());
    setPopoverFor(null);
  }, [selectedThreatId]);

  const selectedPath: AttackPath | null = useMemo(() => {
    if (!response || selectedThreatId === null) return null;
    return response.paths.find(p => p.threat.id === selectedThreatId) ?? null;
  }, [response, selectedThreatId]);

  const simulation = useMemo(() => {
    if (!selectedPath) return null;
    if (disabledControls.size === 0) return null;
    return recomputeW(selectedPath, disabledControls);
  }, [selectedPath, disabledControls]);

  const allControlsForPath = useMemo(() => {
    if (!selectedPath) return new Map<number, ControlCoverage>();
    const m = new Map<number, ControlCoverage>();
    for (const vl of selectedPath.vulnerable_links) {
      for (const c of vl.coverage_controls) {
        if (!m.has(c.id)) m.set(c.id, c);
      }
    }
    return m;
  }, [selectedPath]);

  // ─────────────────── Render states ───────────────────

  if (!assetId) {
    return (
      <div style={{ padding: 40, textAlign: 'center', color: 'var(--fg-muted)' }}>
        Не указан актив.
      </div>
    );
  }

  if (loading) {
    return (
      <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ height: 80, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-md)' }} />
        <div style={{ height: 200, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-md)' }} />
        <div style={{ height: 460, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-md)' }} />
      </div>
    );
  }

  if (error || !response) {
    return (
      <div style={{ padding: 40 }}>
        <Card title="Ошибка загрузки" dense>
          <div style={{ color: 'var(--risk-critical)' }}>{error ?? 'Нет данных'}</div>
          <Btn style={{ marginTop: 14 }} variant="outline" onClick={() => navigate(-1)} icon={<Icon name="arrowL" size={13} />}>
            Назад
          </Btn>
        </Card>
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.3 }}
      style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}
    >
      <AssetRiskHero
        assetId={response.asset.id}
        assetName={response.asset.name}
        aggregate={response.aggregate}
        onBack={() => navigate(-1)}
      />

      <ThreatList
        paths={response.paths}
        selectedThreatId={selectedThreatId}
        onSelect={(id) => {
          const next = new URLSearchParams(params);
          next.set('threat', String(id));
          setParams(next, { replace: true });
        }}
      />

      {selectedPath && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 340px', gap: 16, alignItems: 'start' }}>
          <Card pad={0}>
            <AttackFlowSankey
              path={selectedPath}
              disabledControlIds={disabledControls}
              onControlClick={(id, anchor) => {
                const c = allControlsForPath.get(id);
                if (c) setPopoverFor({ control: c, anchor });
              }}
            />
          </Card>

          <PtsziBreakdown
            path={selectedPath}
            simulated={simulation ?? undefined}
          />
        </div>
      )}

      {selectedPath && simulation && (
        <WhatIfBar
          disabledChips={Array.from(disabledControls).map(id => ({
            id,
            name: allControlsForPath.get(id)?.name ?? `C-${id}`,
          }))}
          baselineW={selectedPath.w}
          simulatedW={simulation.w}
          delta={simulation.delta}
          onReset={() => setDisabledControls(new Set())}
          onRemoveChip={(id) => setDisabledControls(s => {
            const next = new Set(s); next.delete(id); return next;
          })}
          onSaveNote={() => {
            const note = {
              assetId: response.asset.id,
              threatId: selectedPath.threat.id,
              disabledIds: Array.from(disabledControls),
              w: simulation.w,
              wBaseline: selectedPath.w,
              ts: Date.now(),
            };
            const key = 'risk:notes';
            const prev = JSON.parse(localStorage.getItem(key) ?? '[]');
            localStorage.setItem(key, JSON.stringify([note, ...prev]));
            toast.success('Заметка сохранена локально');
          }}
        />
      )}

      {popoverFor && (
        <ControlPopover
          control={popoverFor.control}
          disabled={disabledControls.has(popoverFor.control.id)}
          anchor={popoverFor.anchor}
          onToggle={(id) => {
            setDisabledControls(s => {
              const next = new Set(s);
              if (next.has(id)) next.delete(id); else next.add(id);
              return next;
            });
            setPopoverFor(null);
          }}
          onClose={() => setPopoverFor(null)}
        />
      )}
    </motion.div>
  );
};
```

- [ ] **Step 2: Verify TypeScript compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 3: Run all frontend tests**

Run from `frontend/`: `CI=true npm test`
Expected: PASS (App.test.tsx + riskFlow.test.ts), no failures.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/RiskGraphPage.tsx
git commit -m "$(cat <<'EOF'
feat(frontend): rewrite RiskGraphPage as stacked orchestrator

New layout: AssetRiskHero / ThreatList / (AttackFlowSankey + PtsziBreakdown).
Auto-selects max-W threat when ?threat= absent, mirrors selection to
the URL via replaceState, and keeps a what-if simulation in client state
(disabledControls Set) — recompute mirrors backend QReactionFromVLs.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 14: Delete legacy `RiskGraphSankey` + manual smoke test

**Files:**
- Delete: `frontend/src/components/RiskGraphSankey.tsx`

- [ ] **Step 1: Confirm no references remain**

Run: `grep -rn "RiskGraphSankey" /Users/velvetway/Downloads/CyberRisk/frontend/src/ 2>&1 || echo "no references"`
Expected: only the file itself, no other references.

- [ ] **Step 2: Delete the file**

Run: `git rm frontend/src/components/RiskGraphSankey.tsx`
Expected: file staged for deletion.

- [ ] **Step 3: Verify TypeScript still compiles**

Run from `frontend/`: `npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 4: Manual smoke test (UI)**

This is a manual verification step. The agent must execute it and report observations.

```bash
# Terminal 1 — backend
go run ./cmd/server &
SERVER_PID=$!

# Terminal 2 — frontend dev server
cd frontend && npm start
```

Open browser at the dev URL (usually `http://localhost:3000`). Login if needed. Navigate to `/risk/graph/1` (substitute a real asset id from your DB; e.g., the first one in `/assets`).

Verify:
- Hero shows W, level badge, threat count, uncovered count.
- ThreatList shows ≥1 threat, sortable; clicking selects, URL updates with `?threat=<id>`.
- Sankey renders 5 columns when there are VLs and controls; column labels visible.
- Hover on a node highlights its links.
- Click on a control (`C` node) opens `ControlPopover`. Toggle «Выкл в симуляции».
- WhatIfBar appears at the bottom; ΔW updates; chip remove and Reset both work.
- Refresh with `?threat=<id>` deep-link works.
- Reload without `?threat=` auto-picks max-W.

Stop both servers (`kill $SERVER_PID` and Ctrl-C the dev server).

If anything fails, fix it before committing — Step 5 only commits on green.

- [ ] **Step 5: Commit deletion**

```bash
git commit -m "$(cat <<'EOF'
chore(frontend): drop legacy RiskGraphSankey component

Replaced by AttackFlowSankey on the rewritten RiskGraphPage; the legacy
component was no longer referenced by any route or page.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

- [ ] **Step 6: Final verification**

Run: `go test ./internal/...`
Expected: PASS.

Run from `frontend/`: `CI=true npm test`
Expected: PASS.

Run: `git status --short`
Expected: clean tree (no uncommitted changes).

---

## Self-Review

This plan addresses every section of the spec:

- **Цели (4 боли)** → Tasks 6 (riskFlow honest semantics + visible C nodes) + 8 (ThreatList sorts → comparison) + 10 (clickable nodes) ✓
- **Маршруты** → Task 13 (URL handling: optional `?threat=`, replaceState) ✓
- **Архитектура страницы (Stacked)** → Task 13 (orchestrator stitches Hero/List/Detail) ✓
- **Компоненты (7 шт)** → Tasks 7–13 ✓
- **Sankey flow-семантика** → Task 6 (`buildSankeyGraph`) + tests ✓
- **What-if симуляция** → Task 6 (`recomputeW`) + Task 13 (state) + Task 11 (popover) + Task 12 (bar) ✓
- **API изменения** → Tasks 1–4 ✓
- **Бэкенд: реализация** → Tasks 1–4 (handler, route, service, helper) ✓
- **Фронтенд: реализация** → Tasks 5–14 ✓
- **Тесты** → ComputeAssetAggregate (Task 2), handler tests (Task 4), riskFlow (Task 6), final smoke (Task 14) ✓
- **Состояния страницы** → Task 13 (loading skeleton, error card, empty handled by ThreatList, deep-link validation with toast) ✓
- **Удаление RiskGraphSankey** → Task 14 ✓

Type/method consistency check: `buildSankeyGraph(path, disabledControlIds)` signature matches between `riskFlow.ts` definition (Task 6), its tests (same task), and consumer `AttackFlowSankey` (Task 10). `recomputeW` returns `{ qReaction, w, delta }` consistently. `ControlPopover` props match the call site in `RiskGraphPage`. ✓

No placeholders, every step has executable content. Plan ready to execute.
