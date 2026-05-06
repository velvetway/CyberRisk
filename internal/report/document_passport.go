package report

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"Diplom/internal/domain"
)

// AssetPassportData — данные для «Технического паспорта АС».
type AssetPassportData struct {
	Asset             domain.Asset
	AssetTypeName     string
	DataCategoryName  string
	Software          []domain.AssetSoftwareWithSoftware
	Controls          []domain.Control
	Vulnerabilities   []domain.AssetVulnerability
	ComplianceScores  []*domain.AssetStandardCompliance
}

// GenerateAssetPassportPDF — «Технический паспорт автоматизированной системы».
// Базовый сводный документ, описывающий актив: общие сведения, состав ПО,
// внедрённые средства защиты, состояние соответствия стандартам.
//
// Формат соответствует разделу «Технический паспорт объекта» из 7.png диплома
// (блок «Формирование информационной модели КС»).
func GenerateAssetPassportPDF(d *AssetPassportData) ([]byte, error) {
	if d == nil {
		return nil, fmt.Errorf("nil passport data")
	}
	pdf := newPDF()
	pdf.AddPage()

	renderDocHeader(pdf, "Технический паспорт автоматизированной системы", d.Asset.Name)

	// 1. Общие сведения
	sectionTitle(pdf, "1. Общие сведения")
	envName := strings.ToUpper(string(d.Asset.Environment))
	if envName == "" {
		envName = "—"
	}
	owner := derefString(d.Asset.Owner, "—")
	descr := derefString(d.Asset.Description, "—")
	contour := "общий контур (Z = 1.0)"
	if d.Asset.IsIsolated {
		contour = "изолированный контур (Z = 0.5)"
	}

	row(pdf, "Наименование", d.Asset.Name)
	row(pdf, "Тип актива", coalesce(d.AssetTypeName, "—"))
	row(pdf, "Категория обрабатываемой информации", coalesce(d.DataCategoryName, "—"))
	row(pdf, "Среда эксплуатации", envName)
	row(pdf, "Принадлежность к контуру", contour)
	row(pdf, "Владелец / ответственный", owner)
	row(pdf, "Идентификатор в реестре", fmt.Sprintf("AS-%06d", d.Asset.ID))
	row(pdf, "Дата постановки на учёт", d.Asset.CreatedAt.Format("02.01.2006"))
	pdf.Ln(3)

	if descr != "—" {
		sectionTitle(pdf, "1.1. Назначение и описание")
		paragraph(pdf, descr)
	}

	// 1.2. Цели и задачи защиты информации (по ПТСЗИ из 7.png диплома).
	sectionTitle(pdf, "1.2. Цели и задачи защиты информации")
	paragraph(pdf, "Цели защиты информации в составе данной автоматизированной системы:")
	paragraph(pdf, "  • обеспечение конфиденциальности — предотвращение несанкционированного раскрытия защищаемой информации;")
	paragraph(pdf, "  • обеспечение целостности — предотвращение несанкционированного и непреднамеренного изменения защищаемой информации;")
	paragraph(pdf, "  • обеспечение доступности — гарантия своевременного предоставления защищаемой информации авторизованным пользователям;")
	paragraph(pdf, "  • обеспечение неотказуемости — фиксация авторов выполняемых действий с защищаемой информацией.")
	paragraph(pdf, "")
	paragraph(pdf, "Задачи защиты информации:")
	paragraph(pdf, "  • противодействие угрозам, актуальным для данного типа активов (см. модель угроз ИБ);")
	paragraph(pdf, "  • выполнение требований нормативной базы — Приказ ФСТЭК №17, ISO/IEC 27001:2022 (см. отчёт о состоянии защищённости);")
	paragraph(pdf, "  • реализация мероприятий по защите АРМ, ЛВС, электронного документооборота и конфиденциальной информации (см. перечень мер защиты).")

	// 2. Состав ПО
	sectionTitle(pdf, "2. Состав программного обеспечения")
	if len(d.Software) == 0 {
		paragraph(pdf, "На актив не привязано ПО из справочника.")
	} else {
		bulletHeader(pdf, []string{"№", "Наименование", "Производитель", "Версия", "Реестр Минцифры"}, []float64{10, 70, 45, 25, 30})
		for i, sw := range d.Software {
			ver := derefString(sw.Link.Version, "—")
			reg := "нет"
			if sw.Software.IsRussian && (sw.Software.RegistryNumber != nil && *sw.Software.RegistryNumber != "") {
				reg = "№" + *sw.Software.RegistryNumber
			} else if sw.Software.IsRussian {
				reg = "да"
			}
			bulletRow(pdf,
				[]string{fmt.Sprintf("%d", i+1), sw.Software.Name, sw.Software.Vendor, ver, reg},
				[]float64{10, 70, 45, 25, 30})
		}
	}
	pdf.Ln(3)

	// 3. Внедрённые контроли
	sectionTitle(pdf, "3. Внедрённые средства защиты информации")
	if len(d.Controls) == 0 {
		paragraph(pdf, "На активе не внедрено ни одного средства защиты. Q^reaction = 0; рекомендуется незамедлительное внедрение базового набора СЗИ.")
	} else {
		bulletHeader(pdf, []string{"№", "Средство защиты", "Описание"}, []float64{10, 60, 110})
		for i, c := range d.Controls {
			descr := derefString(c.Description, "")
			bulletRow(pdf, []string{fmt.Sprintf("%d", i+1), c.Name, descr}, []float64{10, 60, 110})
		}
	}
	pdf.Ln(3)

	// 4. Свидетельства уязвимостей (БДУ)
	sectionTitle(pdf, "4. Зарегистрированные уязвимости (БДУ ФСТЭК)")
	if len(d.Vulnerabilities) == 0 {
		paragraph(pdf, "Свидетельств уязвимостей в инвентаре актива не зафиксировано.")
	} else {
		paragraph(pdf, fmt.Sprintf("Всего записей: %d (часть может относиться к одной VL-категории).", len(d.Vulnerabilities)))
	}
	pdf.Ln(2)

	// 5. Соответствие стандартам
	sectionTitle(pdf, "5. Соответствие требованиям ИБ-стандартов")
	if len(d.ComplianceScores) == 0 {
		paragraph(pdf, "Стандарты соответствия не настроены.")
	} else {
		bulletHeader(pdf, []string{"Стандарт", "%", "✓", "◐", "✗", "Всего"}, []float64{90, 20, 18, 18, 18, 18})
		for _, st := range d.ComplianceScores {
			bulletRow(pdf, []string{
				st.Standard.Name,
				fmt.Sprintf("%.0f%%", st.OverallScore*100),
				fmt.Sprintf("%d", st.CoveredCount),
				fmt.Sprintf("%d", st.PartialCount),
				fmt.Sprintf("%d", st.UncoveredCount),
				fmt.Sprintf("%d", st.TotalCount),
			}, []float64{90, 20, 18, 18, 18, 18})
		}
	}

	// Подпись
	renderDocFooter(pdf, "Документ сформирован автоматически системой CyberRisk на основе данных учёта актива.")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// ------------------------ shared helpers ------------------------

func renderDocHeader(pdf gofpdfShim, title, assetName string) {
	pdf.SetFont("NotoSans", "B", 16)
	pdf.SetTextColor(20, 20, 20)
	pdf.Cell(0, 9, title)
	pdf.Ln(8)
	pdf.SetFont("NotoSans", "", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 5, fmt.Sprintf("Объект: %s   ·   дата формирования: %s", assetName, time.Now().Format("02.01.2006")))
	pdf.Ln(7)
	pdf.SetDrawColor(220, 220, 220)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(5)
}

func renderDocFooter(pdf gofpdfShim, text string) {
	pdf.Ln(8)
	pdf.SetDrawColor(220, 220, 220)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(2)
	pdf.SetFont("NotoSans", "", 8)
	pdf.SetTextColor(140, 140, 140)
	pdf.MultiCell(0, 4, text, "", "L", false)
}

func sectionTitle(pdf gofpdfShim, t string) {
	pdf.Ln(2)
	pdf.SetFont("NotoSans", "B", 12)
	pdf.SetTextColor(30, 30, 30)
	pdf.Cell(0, 7, t)
	pdf.Ln(7)
}

func paragraph(pdf gofpdfShim, t string) {
	pdf.SetFont("NotoSans", "", 10)
	pdf.SetTextColor(60, 60, 60)
	pdf.MultiCell(0, 5, t, "", "L", false)
	pdf.Ln(1)
}

// row рендерит строку «label — value».
func row(pdf gofpdfShim, label, value string) {
	pdf.SetFont("NotoSans", "B", 10)
	pdf.SetTextColor(40, 40, 40)
	pdf.CellFormat(60, 6, label, "", 0, "L", false, 0, "")
	pdf.SetFont("NotoSans", "", 10)
	pdf.SetTextColor(60, 60, 60)
	pdf.MultiCell(0, 6, value, "", "L", false)
}

// bulletHeader рисует «шапку» табличного вывода.
func bulletHeader(pdf gofpdfShim, cols []string, widths []float64) {
	pdf.SetFont("NotoSans", "B", 9)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFillColor(70, 90, 140)
	for i, c := range cols {
		pdf.CellFormat(widths[i], 6, c, "", 0, "L", true, 0, "")
	}
	pdf.Ln(6)
}

// bulletRow рисует одну строку табличного вывода.
func bulletRow(pdf gofpdfShim, cols []string, widths []float64) {
	pdf.SetFont("NotoSans", "", 9)
	pdf.SetTextColor(40, 40, 40)
	for i, c := range cols {
		pdf.CellFormat(widths[i], 5.5, " "+truncate(c, int(widths[i]/2.0)), "", 0, "L", false, 0, "")
	}
	pdf.Ln(5.5)
}

func derefString(s *string, fallback string) string {
	if s == nil || *s == "" {
		return fallback
	}
	return *s
}

func coalesce(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func truncate(s string, max int) string {
	if max < 4 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
