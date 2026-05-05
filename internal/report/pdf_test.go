package report

import (
	"bytes"
	"testing"

	"Diplom/internal/domain"
)

// pdfHeader is the magic bytes every well-formed PDF must start with.
var pdfHeader = []byte("%PDF-")

func TestGenerateAttackPathPDF_ProducesValidPDF(t *testing.T) {
	path := &domain.AttackPath{
		Asset:  domain.AssetRef{ID: 1, Name: "Customer Database"},
		Threat: domain.ThreatRef{ID: 7, Name: "Угроза несанкционированного доступа", BDUID: "УБИ.001"},
		Sources: []domain.ThreatSource{
			{ID: 4, Code: "S4", Name: "Хакеры, мошенники"},
		},
		VulnerableLinks: []domain.VLNode{
			{
				VulnerabilityID: 11,
				Name:            "Открытые ОС / отсутствие средств защиты ЛВС",
				CoverageControls: []domain.ControlCoverage{
					{ID: 2, Name: "Межсетевой экран", Coverage: 1.0},
				},
			},
			{
				VulnerabilityID: 12,
				Name:            "Допустимость установки не декларируемого ПО",
				Uncovered:       true,
			},
		},
		DestructiveActions: []domain.DestructiveAction{
			{ID: 1, Code: "DA1", Name: "Копирование информации",
				AffectsConfidentiality: true},
		},
		QThreat:   0.7,
		QSeverity: 0.8,
		QReaction: 0.5,
		Z:         1.0,
		W:         0.55,
		Level:     "high",
	}

	pdfBytes, err := GenerateAttackPathPDF(path)
	if err != nil {
		t.Fatalf("GenerateAttackPathPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(pdfBytes, pdfHeader) {
		t.Fatalf("output does not look like a PDF (first bytes: %q)", pdfBytes[:min(8, len(pdfBytes))])
	}
	if len(pdfBytes) < 1024 {
		t.Errorf("PDF suspiciously small: %d bytes", len(pdfBytes))
	}
}

func TestGenerateAssetReportPDF_HappyPath(t *testing.T) {
	resp := &domain.AssetAttackPathsResponse{
		Asset: domain.AssetRef{ID: 1, Name: "Customer Database"},
		Aggregate: domain.AssetAggregate{
			WMax: 0.55, Level: "high", ThreatCount: 1, UncoveredCount: 1,
		},
		Paths: []domain.AttackPath{
			{
				Asset:     domain.AssetRef{ID: 1, Name: "Customer Database"},
				Threat:    domain.ThreatRef{ID: 7, Name: "Угроза НСД"},
				QThreat:   0.7,
				QSeverity: 0.8,
				QReaction: 0.5,
				Z:         1.0,
				W:         0.55,
				Level:     "high",
				VulnerableLinks: []domain.VLNode{
					{VulnerabilityID: 12, Name: "VL без покрытия", Uncovered: true},
				},
			},
		},
	}

	pdfBytes, err := GenerateAssetReportPDF(resp)
	if err != nil {
		t.Fatalf("GenerateAssetReportPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(pdfBytes, pdfHeader) {
		t.Fatalf("output does not look like a PDF")
	}
	if len(pdfBytes) < 2048 {
		t.Errorf("PDF suspiciously small: %d bytes", len(pdfBytes))
	}
}

func TestGenerateAttackPathPDF_NilInput(t *testing.T) {
	_, err := GenerateAttackPathPDF(nil)
	if err == nil {
		t.Fatal("expected error for nil input, got nil")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
