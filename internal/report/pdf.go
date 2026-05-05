// Package report builds PDF artefacts of the PTSZI risk model.
//
// The reports never reference the legacy 1..25 / impact×likelihood scale —
// every number on the page comes from the W formula or the S → ST → VL → DA
// graph as defined in docs/risk-model.md.
package report

import (
	"bytes"
	"fmt"
	"strings"

	"Diplom/internal/domain"

	"github.com/jung-kurt/gofpdf"
)

// GenerateAttackPathPDF renders one S → ST → VL → DA chain plus the
// W decomposition for a single (asset, threat) pair.
func GenerateAttackPathPDF(path *domain.AttackPath) ([]byte, error) {
	if path == nil {
		return nil, fmt.Errorf("nil attack path")
	}
	pdf := newPDF()
	pdf.AddPage()
	renderHeader(pdf, "Отчёт по риску ПТСЗИ", path.Asset.Name)
	renderThreatBlock(pdf, path)
	renderWBreakdown(pdf, path)
	renderChain(pdf, path)
	renderRecommendations(pdf, []domain.AttackPath{*path})

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// GenerateAssetReportPDF renders the full PTSZI report for one asset:
// the aggregate metric, then every applicable AttackPath in W-descending order.
func GenerateAssetReportPDF(resp *domain.AssetAttackPathsResponse) ([]byte, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil asset response")
	}
	pdf := newPDF()
	pdf.AddPage()
	renderHeader(pdf, "Сводный отчёт по активу", resp.Asset.Name)
	renderAggregate(pdf, resp.Aggregate, len(resp.Paths))

	for i := range resp.Paths {
		pdf.AddPage()
		path := resp.Paths[i]
		renderHeader(pdf, fmt.Sprintf("Угроза %d из %d", i+1, len(resp.Paths)), resp.Asset.Name)
		renderThreatBlock(pdf, &path)
		renderWBreakdown(pdf, &path)
		renderChain(pdf, &path)
	}

	if len(resp.Paths) > 0 {
		pdf.AddPage()
		pdf.SetFont("NotoSans", "B", 14)
		pdf.Cell(0, 10, "Рекомендации по снижению W")
		pdf.Ln(12)
		renderRecommendations(pdf, resp.Paths)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func newPDF() *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddUTF8FontFromBytes("NotoSans", "", notoSansRegular)
	pdf.AddUTF8FontFromBytes("NotoSans", "B", notoSansBold)
	return pdf
}

func renderHeader(pdf *gofpdf.Fpdf, title, asset string) {
	pdf.SetFont("NotoSans", "B", 18)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 10, title)
	pdf.Ln(10)
	pdf.SetFont("NotoSans", "", 11)
	pdf.SetTextColor(90, 90, 90)
	pdf.Cell(0, 6, "Актив: "+asset)
	pdf.Ln(4)
	pdf.Cell(0, 6, "Модель: ПТСЗИ — формула W = (Q^threat + q^threat + (1 − Q^reaction)) / 3 · Z")
	pdf.Ln(8)
	pdf.SetDrawColor(220, 220, 220)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(6)
}

func renderThreatBlock(pdf *gofpdf.Fpdf, path *domain.AttackPath) {
	pdf.SetFont("NotoSans", "B", 13)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 8, path.Threat.Name)
	pdf.Ln(7)
	if path.Threat.BDUID != "" {
		pdf.SetFont("NotoSans", "", 10)
		pdf.SetTextColor(110, 110, 110)
		pdf.Cell(0, 5, "БДУ ФСТЭК: "+path.Threat.BDUID)
		pdf.Ln(8)
	}
}

func renderAggregate(pdf *gofpdf.Fpdf, agg domain.AssetAggregate, totalPaths int) {
	pdf.SetFont("NotoSans", "B", 13)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 8, "Агрегат по активу")
	pdf.Ln(8)

	r, g, b := levelRGB(agg.Level)
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("NotoSans", "B", 22)
	pdf.CellFormat(60, 18, fmt.Sprintf("W max = %.2f", agg.WMax), "", 0, "C", true, 0, "")
	pdf.Ln(20)

	pdf.SetTextColor(60, 60, 60)
	pdf.SetFont("NotoSans", "", 11)
	pdf.Cell(0, 6, fmt.Sprintf("Уровень риска: %s", levelLabel(agg.Level)))
	pdf.Ln(5)
	pdf.Cell(0, 6, fmt.Sprintf("Угроз учтено: %d (всего цепочек собрано: %d)", agg.ThreatCount, totalPaths))
	pdf.Ln(5)
	pdf.Cell(0, 6, fmt.Sprintf("Непокрытых уязвимых звеньев: %d", agg.UncoveredCount))
	pdf.Ln(10)
}

func renderWBreakdown(pdf *gofpdf.Fpdf, path *domain.AttackPath) {
	pdf.SetFont("NotoSans", "B", 12)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 7, "Разложение W")
	pdf.Ln(8)

	pdf.SetFont("NotoSans", "", 10)
	pdf.SetTextColor(60, 60, 60)
	rows := []struct {
		label string
		value string
	}{
		{"Q^threat (степень реализации)", fmt.Sprintf("%.2f", path.QThreat)},
		{"q^threat (степень опасности)", fmt.Sprintf("%.2f", path.QSeverity)},
		{"Q^reaction (степень предотвращения)", fmt.Sprintf("%.2f", path.QReaction)},
		{"Z (коэффициент контура)", fmt.Sprintf("%.2f", path.Z)},
	}
	for _, row := range rows {
		pdf.CellFormat(95, 6, row.label, "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, row.value, "", 0, "L", false, 0, "")
		pdf.Ln(6)
	}
	pdf.Ln(2)

	r, g, b := levelRGB(path.Level)
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("NotoSans", "B", 16)
	pdf.CellFormat(0, 12,
		fmt.Sprintf("Итог: W = %.2f  ·  %s", path.W, levelLabel(path.Level)),
		"", 0, "C", true, 0, "")
	pdf.Ln(16)
}

func renderChain(pdf *gofpdf.Fpdf, path *domain.AttackPath) {
	pdf.SetFont("NotoSans", "B", 12)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 7, "Цепочка атаки S → ST → VL → DA")
	pdf.Ln(8)

	pdf.SetFont("NotoSans", "", 10)
	pdf.SetTextColor(60, 60, 60)

	if len(path.Sources) > 0 {
		names := make([]string, len(path.Sources))
		for i, s := range path.Sources {
			names[i] = fmt.Sprintf("%s — %s", s.Code, s.Name)
		}
		pdf.MultiCell(0, 5, "Источники (S): "+strings.Join(names, "; "), "", "L", false)
		pdf.Ln(2)
	}

	if len(path.VulnerableLinks) > 0 {
		pdf.SetFont("NotoSans", "B", 10)
		pdf.Cell(0, 6, "Уязвимые звенья и покрытие контролями:")
		pdf.Ln(6)
		pdf.SetFont("NotoSans", "", 10)
		for _, vl := range path.VulnerableLinks {
			pdf.SetTextColor(20, 20, 20)
			marker := "[покрыто]"
			markerR, markerG, markerB := 30, 130, 60
			if vl.Uncovered || len(vl.CoverageControls) == 0 {
				marker = "[НЕ ПОКРЫТО]"
				markerR, markerG, markerB = 200, 50, 50
			}
			pdf.SetTextColor(markerR, markerG, markerB)
			pdf.CellFormat(35, 5, marker, "", 0, "L", false, 0, "")
			pdf.SetTextColor(20, 20, 20)
			pdf.MultiCell(0, 5, vl.Name, "", "L", false)
			if len(vl.CoverageControls) > 0 {
				pdf.SetTextColor(110, 110, 110)
				pdf.SetFont("NotoSans", "", 9)
				ctrlNames := make([]string, len(vl.CoverageControls))
				for i, c := range vl.CoverageControls {
					ctrlNames[i] = fmt.Sprintf("%s (%.0f%%)", c.Name, c.Coverage*100)
				}
				pdf.SetX(50)
				pdf.MultiCell(0, 4, "Контроли: "+strings.Join(ctrlNames, ", "), "", "L", false)
				pdf.SetFont("NotoSans", "", 10)
				pdf.SetTextColor(20, 20, 20)
			}
			pdf.Ln(1)
		}
		pdf.Ln(2)
	} else {
		pdf.SetTextColor(110, 110, 110)
		pdf.Cell(0, 6, "Уязвимые звенья: не найдены")
		pdf.Ln(8)
		pdf.SetTextColor(60, 60, 60)
	}

	if len(path.DestructiveActions) > 0 {
		pdf.SetFont("NotoSans", "B", 10)
		pdf.SetTextColor(20, 20, 20)
		pdf.Cell(0, 6, "Деструктивные действия (DA):")
		pdf.Ln(6)
		pdf.SetFont("NotoSans", "", 10)
		pdf.SetTextColor(60, 60, 60)
		for _, da := range path.DestructiveActions {
			cia := []string{}
			if da.AffectsConfidentiality {
				cia = append(cia, "C")
			}
			if da.AffectsIntegrity {
				cia = append(cia, "I")
			}
			if da.AffectsAvailability {
				cia = append(cia, "A")
			}
			line := fmt.Sprintf("• %s — %s", da.Code, da.Name)
			if len(cia) > 0 {
				line += fmt.Sprintf("  [%s]", strings.Join(cia, "/"))
			}
			pdf.MultiCell(0, 5, line, "", "L", false)
		}
	}
	pdf.Ln(4)
}

func renderRecommendations(pdf *gofpdf.Fpdf, paths []domain.AttackPath) {
	type rec struct {
		vlName  string
		threats map[string]struct{}
	}
	uncovered := map[int64]*rec{}
	for _, p := range paths {
		for _, vl := range p.VulnerableLinks {
			if !vl.Uncovered && len(vl.CoverageControls) > 0 {
				continue
			}
			r, ok := uncovered[vl.VulnerabilityID]
			if !ok {
				r = &rec{vlName: vl.Name, threats: map[string]struct{}{}}
				uncovered[vl.VulnerabilityID] = r
			}
			r.threats[p.Threat.Name] = struct{}{}
		}
	}

	pdf.SetFont("NotoSans", "B", 12)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 7, "Что повысит Q^reaction (и снизит W)")
	pdf.Ln(8)

	if len(uncovered) == 0 {
		pdf.SetFont("NotoSans", "", 10)
		pdf.SetTextColor(60, 60, 60)
		pdf.Cell(0, 6, "Все уязвимые звенья закрыты внедрёнными контролями.")
		return
	}

	pdf.SetFont("NotoSans", "", 10)
	pdf.SetTextColor(60, 60, 60)
	idx := 1
	for _, r := range uncovered {
		threats := make([]string, 0, len(r.threats))
		for t := range r.threats {
			threats = append(threats, t)
		}
		pdf.SetFont("NotoSans", "B", 10)
		pdf.SetTextColor(20, 20, 20)
		pdf.MultiCell(0, 6, fmt.Sprintf("%d. %s", idx, r.vlName), "", "L", false)
		pdf.SetFont("NotoSans", "", 9)
		pdf.SetTextColor(110, 110, 110)
		pdf.MultiCell(0, 4, "  Затрагивает угрозы: "+strings.Join(threats, "; "), "", "L", false)
		pdf.MultiCell(0, 4, "  Действие: внедрить хотя бы один контроль из vulnerability_controls для этого VL на активе.", "", "L", false)
		pdf.Ln(2)
		idx++
	}
}

func levelRGB(level string) (int, int, int) {
	switch strings.ToLower(level) {
	case "critical":
		return 200, 50, 50
	case "high":
		return 235, 130, 40
	case "medium":
		return 235, 195, 60
	case "low":
		return 80, 170, 90
	default:
		return 110, 110, 110
	}
}

func levelLabel(level string) string {
	switch strings.ToLower(level) {
	case "critical":
		return "Критический"
	case "high":
		return "Высокий"
	case "medium":
		return "Средний"
	case "low":
		return "Низкий"
	default:
		return level
	}
}
