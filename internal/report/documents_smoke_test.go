package report

import (
	"bytes"
	"testing"
	"time"

	"Diplom/internal/domain"
)

// Smoke-тесты: PDF успешно генерируется и начинается с %PDF-.
// Не проверяем визуальное содержимое — это слишком хрупко; но если
// гарантировано не падает на минимально валидных входных данных, это
// ловит 80% регрессий типа nil-pointer / неинициализированных полей.

func TestGenerateAssetPassportPDF_Smoke(t *testing.T) {
	asset := domain.Asset{
		ID: 1, Name: "Test asset",
		Environment: "prod",
		IsIsolated:  false,
		CreatedAt:   time.Now(),
	}

	out, err := GenerateAssetPassportPDF(&AssetPassportData{
		Asset:            asset,
		AssetTypeName:    "Database",
		DataCategoryName: "Персональные данные",
		Software:         nil,
		Controls:         nil,
		Vulnerabilities:  nil,
		ComplianceScores: nil,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.HasPrefix(out, pdfHeader) {
		t.Errorf("результат не начинается с %%PDF-")
	}
	if len(out) < 1000 {
		t.Errorf("слишком маленький PDF: %d байт", len(out))
	}
}

func TestGenerateAssetPassportPDF_NilData(t *testing.T) {
	if _, err := GenerateAssetPassportPDF(nil); err == nil {
		t.Errorf("ожидали ошибку на nil data")
	}
}

func TestGenerateThreatModelPDF_NoThreats(t *testing.T) {
	asset := &domain.Asset{ID: 1, Name: "Test"}
	out, err := GenerateThreatModelPDF(asset, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.HasPrefix(out, pdfHeader) {
		t.Errorf("результат не начинается с %%PDF-")
	}
}

func TestGenerateThreatModelPDF_WithThreats(t *testing.T) {
	asset := &domain.Asset{ID: 1, Name: "Test"}
	resp := &domain.AssetAttackPathsResponse{
		Asset:     domain.AssetRef{ID: 1, Name: "Test"},
		Aggregate: domain.AssetAggregate{WMax: 0.7, Level: "high", ThreatCount: 1, UncoveredCount: 1},
		Paths: []domain.AttackPath{{
			Threat:    domain.ThreatRef{ID: 1, Name: "Какая-то угроза с очень длинным названием которое должно безопасно обработаться", BDUID: "УБИ.001"},
			QThreat:   0.6,
			QSeverity: 0.9,
			QReaction: 0.0,
			Z:         1.0,
			W:         0.83,
			Level:     "critical",
			VulnerableLinks: []domain.VLNode{
				{CategoryID: 1, Code: "VL1", Name: "Тестовая VL", Uncovered: true},
			},
		}},
	}
	out, err := GenerateThreatModelPDF(asset, resp)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.HasPrefix(out, pdfHeader) {
		t.Errorf("результат не начинается с %%PDF-")
	}
}

func TestGenerateThreatModelPDF_NilAsset(t *testing.T) {
	if _, err := GenerateThreatModelPDF(nil, nil); err == nil {
		t.Errorf("ожидали ошибку на nil asset")
	}
}

func TestGenerateProtectionPlanPDF_Smoke(t *testing.T) {
	out, err := GenerateProtectionPlanPDF(&ProtectionPlanData{
		Asset:            domain.Asset{ID: 1, Name: "Test"},
		Controls:         []domain.Control{{ID: 1, Name: "Антивирус"}},
		AttackPaths:      nil,
		ComplianceScores: nil,
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.HasPrefix(out, pdfHeader) {
		t.Errorf("результат не начинается с %%PDF-")
	}
}

func TestGenerateProtectionPlanPDF_NilData(t *testing.T) {
	if _, err := GenerateProtectionPlanPDF(nil); err == nil {
		t.Errorf("ожидали ошибку на nil data")
	}
}

func TestGenerateCompliancePDF_Smoke(t *testing.T) {
	standards := []*domain.AssetStandardCompliance{{
		Standard:       domain.ComplianceStandard{ID: 1, Code: "FSTEC_17", Name: "ФСТЭК №17", FullName: "Приказ ФСТЭК № 17"},
		OverallScore:   0.5,
		CoveredCount:   3,
		PartialCount:   2,
		UncoveredCount: 1,
		TotalCount:     6,
		Requirements: []domain.RequirementStatus{
			{
				Requirement: domain.ComplianceRequirement{ID: 1, Code: "ИАФ.1", Title: "Тест", Category: "ИАФ"},
				Coverage:    1.0,
			},
		},
	}}
	out, err := GenerateCompliancePDF("Test asset", standards)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.HasPrefix(out, pdfHeader) {
		t.Errorf("результат не начинается с %%PDF-")
	}
}

func TestGenerateCompliancePDF_EmptyStandards(t *testing.T) {
	if _, err := GenerateCompliancePDF("Test", nil); err == nil {
		t.Errorf("ожидали ошибку на пустом списке стандартов")
	}
}

func TestGenerateOrganizationReportPDF_Smoke(t *testing.T) {
	overview := &domain.OrganizationOverview{
		TotalAssets:      3,
		IsolatedAssets:   1,
		AssetsByEnvironment: map[string]int{"prod": 2, "dev": 1},
		RiskDistribution: map[string]int{"critical": 1, "high": 1, "medium": 1, "low": 0},
		WMax:             0.93,
		WMaxAsset:        "Test asset",
		WMaxThreat:       "Test threat",
		AvgWPerAsset:     0.7,
		TotalControls:    5,
		AssetsByType: []domain.AssetTypeBucket{
			{TypeName: "Server", Count: 2},
			{TypeName: "Database", Count: 1},
		},
	}
	out, err := GenerateOrganizationReportPDF(overview, nil, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !bytes.HasPrefix(out, pdfHeader) {
		t.Errorf("результат не начинается с %%PDF-")
	}
}

func TestGenerateOrganizationReportPDF_NilOverview(t *testing.T) {
	if _, err := GenerateOrganizationReportPDF(nil, nil, nil); err == nil {
		t.Errorf("ожидали ошибку на nil overview")
	}
}

// PackDocuments склеивает PDF в zip — проверяем что zip-magic присутствует.
func TestPackDocuments_Smoke(t *testing.T) {
	entries := []DocumentPackEntry{
		{Filename: "a.pdf", PDF: []byte("%PDF-fakecontent")},
		{Filename: "b.pdf", PDF: []byte("%PDF-fake2")},
	}
	out, err := PackDocuments(entries)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// ZIP magic: PK\x03\x04
	if len(out) < 4 || out[0] != 'P' || out[1] != 'K' {
		t.Errorf("результат не похож на ZIP-архив")
	}
}

func TestPackDocuments_SkipsEmptyEntries(t *testing.T) {
	entries := []DocumentPackEntry{
		{Filename: "a.pdf", PDF: []byte("%PDF-content")},
		{Filename: "empty.pdf", PDF: nil}, // должен быть пропущен
	}
	out, err := PackDocuments(entries)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(out) < 4 || out[0] != 'P' || out[1] != 'K' {
		t.Errorf("результат не похож на ZIP-архив")
	}
}
