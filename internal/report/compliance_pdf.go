package report

import (
	"bytes"
	"fmt"
	"strings"

	"Diplom/internal/domain"
)

// GenerateCompliancePDF строит сводный отчёт о состоянии защищённости актива
// (ОСЗ) по нескольким стандартам ИБ — ФСТЭК, ISO 27001 и т.д.
//
// Структура страниц:
//   1. Сводка: блок с названием актива + «плитки» по каждому стандарту
//      с overall_score и счётчиками ✓/◐/✗.
//   2. По одному отдельному разделу на стандарт: список требований
//      сгруппированный по категориям; каждое требование показывает
//      coverage %, список покрывающих контролей и список рекомендуемых
//      (которые ещё закрыли бы это требование).
//
// `assetName` — то, что попадёт в заголовок отчёта; обычно asset.Name.
func GenerateCompliancePDF(assetName string, standards []*domain.AssetStandardCompliance) ([]byte, error) {
	if len(standards) == 0 {
		return nil, fmt.Errorf("compliance pdf: empty standards list")
	}
	pdf := newPDF()
	pdf.AddPage()
	renderHeaderCompliance(pdf, "Отчёт о состоянии защищённости (ОСЗ)", assetName)
	renderComplianceSummary(pdf, standards)

	for _, st := range standards {
		pdf.AddPage()
		renderHeaderCompliance(pdf, st.Standard.Name, assetName)
		renderStandardDetail(pdf, st)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func renderHeaderCompliance(pdf interface {
	SetFont(family, style string, size float64)
	SetTextColor(r, g, b int)
	Cell(w, h float64, txt string)
	Ln(h float64)
	SetDrawColor(r, g, b int)
	Line(x1, y1, x2, y2 float64)
	GetY() float64
}, title, asset string) {
	pdf.SetFont("NotoSans", "B", 18)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 10, title)
	pdf.Ln(10)
	pdf.SetFont("NotoSans", "", 11)
	pdf.SetTextColor(90, 90, 90)
	pdf.Cell(0, 6, "Актив: "+asset)
	pdf.Ln(4)
	pdf.Cell(0, 6, "Источник: Приказ ФСТЭК №17 + ISO/IEC 27001:2022 (Annex A)")
	pdf.Ln(8)
	pdf.SetDrawColor(220, 220, 220)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(6)
}

func renderComplianceSummary(pdf gofpdfShim, standards []*domain.AssetStandardCompliance) {
	pdf.SetFont("NotoSans", "B", 13)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 8, "Сводка по стандартам")
	pdf.Ln(10)

	for _, st := range standards {
		renderComplianceTile(pdf, st)
		pdf.Ln(4)
	}
}

func renderComplianceTile(pdf gofpdfShim, st *domain.AssetStandardCompliance) {
	pct := st.OverallScore * 100
	r, g, b := complianceLevelRGB(st.OverallScore)

	// Цветной прямоугольник со score
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("NotoSans", "B", 18)
	pdf.CellFormat(40, 16, fmt.Sprintf("%.0f%%", pct), "", 0, "C", true, 0, "")

	// Название стандарта + счётчики справа
	pdf.SetTextColor(20, 20, 20)
	pdf.SetFont("NotoSans", "B", 12)
	pdf.CellFormat(0, 8, "  "+st.Standard.Name, "", 1, "L", false, 0, "")

	pdf.SetX(55)
	pdf.SetFont("NotoSans", "", 10)
	pdf.SetTextColor(90, 90, 90)
	pdf.CellFormat(0, 8, fmt.Sprintf("  закрыто: %d   частично: %d   не закрыто: %d   всего: %d",
		st.CoveredCount, st.PartialCount, st.UncoveredCount, st.TotalCount), "", 1, "L", false, 0, "")

	pdf.Ln(3)
	pdf.SetFont("NotoSans", "", 9)
	pdf.SetTextColor(110, 110, 110)
	pdf.MultiCell(0, 4, st.Standard.FullName, "", "L", false)
	pdf.Ln(2)
}

func renderStandardDetail(pdf gofpdfShim, st *domain.AssetStandardCompliance) {
	// Подзаголовок-сводка
	pdf.SetFont("NotoSans", "B", 13)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 8, fmt.Sprintf("Соответствие: %.1f%% (%d из %d требований)",
		st.OverallScore*100, st.CoveredCount, st.TotalCount))
	pdf.Ln(8)

	pdf.SetFont("NotoSans", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.Cell(0, 5, fmt.Sprintf("Полностью закрыто: %d   ·   Частично: %d   ·   Не выполнено: %d",
		st.CoveredCount, st.PartialCount, st.UncoveredCount))
	pdf.Ln(8)

	// Группируем требования по категории, сохраняя порядок появления.
	var catOrder []string
	byCat := map[string][]domain.RequirementStatus{}
	for _, r := range st.Requirements {
		k := r.Requirement.Category
		if k == "" {
			k = "Прочее"
		}
		if _, seen := byCat[k]; !seen {
			catOrder = append(catOrder, k)
		}
		byCat[k] = append(byCat[k], r)
	}

	for _, cat := range catOrder {
		items := byCat[cat]
		pdf.SetFont("NotoSans", "B", 11)
		pdf.SetTextColor(40, 40, 40)
		pdf.Cell(0, 7, fmt.Sprintf("%s (%d)", cat, len(items)))
		pdf.Ln(7)

		for _, r := range items {
			renderRequirementRow(pdf, r)
		}
		pdf.Ln(2)
	}
}

func renderRequirementRow(pdf gofpdfShim, rs domain.RequirementStatus) {
	r, g, b := complianceLevelRGB(rs.Coverage)

	// Метка статуса (текстовая — NotoSans не имеет ✓◐✗).
	pdf.SetFillColor(r, g, b)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("NotoSans", "B", 8)
	pdf.CellFormat(14, 5.5, statusGlyph(rs.Coverage), "", 0, "C", true, 0, "")

	// Код + заголовок
	pdf.SetTextColor(30, 30, 30)
	pdf.SetFont("NotoSans", "B", 9)
	pdf.CellFormat(20, 5.5, " "+rs.Requirement.Code, "", 0, "L", false, 0, "")

	pdf.SetFont("NotoSans", "", 9)
	pdf.SetTextColor(60, 60, 60)
	pdf.CellFormat(0, 5.5, rs.Requirement.Title+fmt.Sprintf("   [%.0f%%]", rs.Coverage*100), "", 1, "L", false, 0, "")

	// Внедрённые контроли
	if names := controlNames(rs.CoveringControls); names != "" {
		pdf.SetFont("NotoSans", "", 8)
		pdf.SetTextColor(70, 130, 80)
		pdf.SetX(40)
		pdf.MultiCell(0, 3.5, "  закрыто: "+names, "", "L", false)
	}
	// Что бы ещё закрыло
	if rs.Coverage < 1.0 {
		if names := controlNames(rs.MissingControls); names != "" {
			pdf.SetFont("NotoSans", "", 8)
			pdf.SetTextColor(170, 110, 50)
			pdf.SetX(40)
			pdf.MultiCell(0, 3.5, "  рекомендуем: "+names, "", "L", false)
		}
	}
	pdf.Ln(1)
}

func controlNames(cs []domain.Control) string {
	if len(cs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(cs))
	for _, c := range cs {
		parts = append(parts, c.Name)
	}
	return strings.Join(parts, ", ")
}

func complianceLevelRGB(score float64) (int, int, int) {
	switch {
	case score >= 0.8:
		return 80, 170, 90 // зелёный
	case score >= 0.5:
		return 235, 195, 60 // жёлтый
	case score >= 0.25:
		return 235, 130, 40 // оранжевый
	default:
		return 200, 50, 50 // красный
	}
}

// gofpdfShim — узкий интерфейс используемых нами методов *gofpdf.Fpdf,
// чтобы render-функции можно было тестировать с минимальной поверхностью.
type gofpdfShim interface {
	SetFont(family, style string, size float64)
	SetTextColor(r, g, b int)
	SetFillColor(r, g, b int)
	SetDrawColor(r, g, b int)
	Cell(w, h float64, txt string)
	CellFormat(w, h float64, txt, border string, ln int, align string, fill bool, link int, linkStr string)
	Ln(h float64)
	MultiCell(w, h float64, txt, border, align string, fill bool)
	SetX(x float64)
	GetY() float64
	Line(x1, y1, x2, y2 float64)
}
