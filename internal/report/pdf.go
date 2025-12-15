package report

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"Diplom/internal/service/risk"

	"github.com/jung-kurt/gofpdf"
)

// getFontsDir возвращает путь к директории со шрифтами
func getFontsDir() string {
	// Пробуем найти директорию относительно исполняемого файла
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		fontsDir := filepath.Join(dir, "internal", "report", "fonts")
		if _, err := os.Stat(fontsDir); err == nil {
			return fontsDir
		}
	}

	// Пробуем найти относительно рабочей директории
	wd, err := os.Getwd()
	if err == nil {
		fontsDir := filepath.Join(wd, "internal", "report", "fonts")
		if _, err := os.Stat(fontsDir); err == nil {
			return fontsDir
		}
	}

	// Fallback - стандартный относительный путь
	return "internal/report/fonts"
}

func GenerateRiskPDF(assetID, threatID int64, res risk.RiskResult, recs []risk.RiskRecommendation) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")

	useCustomFonts := tryAddEmbeddedFonts(pdf)
	if !useCustomFonts {
		fontsDir := getFontsDir()
		useCustomFonts = tryAddFontsFromFS(pdf,
			filepath.Join(fontsDir, "NotoSans-Regular.ttf"),
			filepath.Join(fontsDir, "NotoSans-Bold.ttf"),
		)
	}

	if useCustomFonts {
		renderReportWithCustomFonts(pdf, assetID, threatID, res, recs)
	} else {
		renderReportWithDefaultFonts(pdf, assetID, threatID, res, recs)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func tryAddEmbeddedFonts(pdf *gofpdf.Fpdf) bool {
	if len(notoSansRegular) == 0 || len(notoSansBold) == 0 {
		return false
	}

	pdf.AddUTF8FontFromBytes("NotoSans", "", notoSansRegular)
	pdf.AddUTF8FontFromBytes("NotoSans", "B", notoSansBold)
	if pdf.Ok() {
		return true
	}

	pdf.ClearError()
	return false
}

func tryAddFontsFromFS(pdf *gofpdf.Fpdf, regularPath, boldPath string) bool {
	if _, err := os.Stat(regularPath); err != nil {
		return false
	}
	if _, err := os.Stat(boldPath); err != nil {
		return false
	}

	pdf.AddUTF8Font("NotoSans", "", regularPath)
	pdf.AddUTF8Font("NotoSans", "B", boldPath)
	if pdf.Ok() {
		return true
	}

	pdf.ClearError()
	return false
}

func renderReportWithDefaultFonts(pdf *gofpdf.Fpdf, assetID, threatID int64, res risk.RiskResult, recs []risk.RiskRecommendation) {
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 20)
	pdf.Cell(0, 14, "Cyber Risk Report")
	pdf.Ln(16)

	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(0, 7, fmt.Sprintf("Asset ID: %d", assetID))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Threat ID: %d", threatID))
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "B", 14)
	pdf.Cell(0, 8, "Risk Parameters:")
	pdf.Ln(10)

	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(0, 6, fmt.Sprintf("Impact: %d", res.Impact))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Likelihood: %d", res.Likelihood))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Risk Score: %d", res.Score))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Risk Level: %s", res.Level))
	pdf.Ln(6)
	if res.RegulatoryFactor > 0 {
		pdf.Cell(0, 6, fmt.Sprintf("Regulatory Factor: %.2f", res.RegulatoryFactor))
		pdf.Ln(6)
		pdf.Cell(0, 6, fmt.Sprintf("Adjusted Score: %.1f", res.AdjustedScore))
		pdf.Ln(6)
	}
	pdf.Ln(6)

	pdf.SetFont("Helvetica", "B", 14)
	pdf.Cell(0, 8, "Recommendations:")
	pdf.Ln(10)

	if len(recs) == 0 {
		pdf.SetFont("Helvetica", "", 12)
		pdf.Cell(0, 6, "No recommendations.")
		pdf.Ln(6)
	} else {
		for _, r := range recs {
			pdf.SetFont("Helvetica", "B", 12)
			pdf.Cell(0, 6, fmt.Sprintf("[%s] %s", r.Priority, r.Title))
			pdf.Ln(5)

			pdf.SetFont("Helvetica", "", 11)
			pdf.MultiCell(0, 5,
				fmt.Sprintf("Category: %s\nDescription: %s\n", r.Category, r.Description),
				"", "L", false,
			)
			pdf.Ln(3)
		}
	}
}

func renderReportWithCustomFonts(pdf *gofpdf.Fpdf, assetID, threatID int64, res risk.RiskResult, recs []risk.RiskRecommendation) {
	pdf.AddPage()

	pdf.SetFont("NotoSans", "B", 20)
	pdf.Cell(0, 14, "Отчёт по киберриску")
	pdf.Ln(16)

	pdf.SetFont("NotoSans", "", 12)
	pdf.Cell(0, 7, fmt.Sprintf("ID актива: %d", assetID))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("ID угрозы: %d", threatID))
	pdf.Ln(10)

	pdf.SetFont("NotoSans", "B", 14)
	pdf.Cell(0, 8, "Параметры риска:")
	pdf.Ln(10)

	pdf.SetFont("NotoSans", "", 12)
	pdf.Cell(0, 6, fmt.Sprintf("Влияние: %d", res.Impact))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Вероятность: %d", res.Likelihood))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Оценка риска: %d", res.Score))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Уровень риска: %s", res.Level))
	pdf.Ln(6)
	if res.RegulatoryFactor > 0 {
		pdf.Cell(0, 6, fmt.Sprintf("Регуляторный множитель: %.2f", res.RegulatoryFactor))
		pdf.Ln(6)
		pdf.Cell(0, 6, fmt.Sprintf("Скорректированная оценка: %.1f", res.AdjustedScore))
		pdf.Ln(6)
	}
	pdf.Ln(6)

	pdf.SetFont("NotoSans", "B", 14)
	pdf.Cell(0, 8, "Рекомендации:")
	pdf.Ln(10)

	if len(recs) == 0 {
		pdf.SetFont("NotoSans", "", 12)
		pdf.Cell(0, 6, "Рекомендации отсутствуют.")
		pdf.Ln(6)
	} else {
		for _, r := range recs {
			pdf.SetFont("NotoSans", "B", 12)
			pdf.Cell(0, 6, fmt.Sprintf("[%s] %s", r.Priority, r.Title))
			pdf.Ln(5)

			pdf.SetFont("NotoSans", "", 11)
			pdf.MultiCell(0, 5,
				fmt.Sprintf("Категория: %s\nОписание: %s\n", r.Category, r.Description),
				"", "L", false,
			)
			pdf.Ln(3)
		}
	}
}
