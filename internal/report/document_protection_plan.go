package report

import (
	"bytes"
	"fmt"

	"Diplom/internal/domain"
)

// ProtectionPlanData — данные для «Перечня мер защиты».
type ProtectionPlanData struct {
	Asset            domain.Asset
	Controls         []domain.Control
	AttackPaths      []domain.AttackPath
	ComplianceScores []*domain.AssetStandardCompliance
}

// GenerateProtectionPlanPDF — «Перечень мер защиты информации и
// рекомендации по совершенствованию системы ИБ». Закрывает блок
// «Формирование рекомендаций по совершенствованию системы ИБ» из 7.png
// и «Рекомендации по усовершенствованию» из 8.png.
//
// Документ группирует контроли и рекомендации по 4 «мероприятиям» из
// 8.png диплома: защита АРМ, защита ЛВС, защита электронного
// документооборота, защита конфиденциальной информации.
func GenerateProtectionPlanPDF(d *ProtectionPlanData) ([]byte, error) {
	if d == nil {
		return nil, fmt.Errorf("nil protection plan data")
	}
	pdf := newPDF()
	pdf.AddPage()

	renderDocHeader(pdf, "Перечень мер защиты информации", d.Asset.Name)

	sectionTitle(pdf, "1. Назначение документа")
	paragraph(pdf, "Документ систематизирует внедрённые на активе средства защиты информации (СЗИ) и формирует адресные рекомендации по их совершенствованию. Группировка ведётся по четырём блокам мероприятий, принятых в дипломной работе ПТСЗИ: защита автоматизированных рабочих мест, защита локальной вычислительной сети, защита электронного документооборота, защита конфиденциальной информации.")

	// 2. Сводка
	sectionTitle(pdf, "2. Сводное состояние защиты")
	row(pdf, "Внедрено СЗИ", fmt.Sprintf("%d из 11", len(d.Controls)))
	if len(d.ComplianceScores) > 0 {
		for _, st := range d.ComplianceScores {
			row(pdf, "Соответствие "+st.Standard.Name, fmt.Sprintf("%.0f%%  (✓%d / ◐%d / ✗%d из %d)",
				st.OverallScore*100, st.CoveredCount, st.PartialCount, st.UncoveredCount, st.TotalCount))
		}
	}
	if uncovered := countUncoveredVL(d.AttackPaths); uncovered > 0 {
		row(pdf, "Непокрытых VL по угрозам", fmt.Sprintf("%d", uncovered))
	}
	pdf.Ln(2)

	// 3. Мероприятия — группировка контролей
	sectionTitle(pdf, "3. Сгруппированные мероприятия защиты")

	groups := groupControlsByMeasure(d.Controls)
	for _, g := range groups {
		renderMeasureBlock(pdf, g)
	}

	// 4. Рекомендации (адресные, по непокрытым VL)
	pdf.AddPage()
	renderDocHeader(pdf, "Рекомендации по совершенствованию системы ИБ", d.Asset.Name)

	sectionTitle(pdf, "4. Адресные рекомендации")
	if len(d.AttackPaths) == 0 {
		paragraph(pdf, "База угроз не оценена для актива; рекомендации не сформированы.")
	} else {
		recs := buildRecommendations(d.AttackPaths)
		if len(recs) == 0 {
			paragraph(pdf, "Все уязвимые звенья по применимым угрозам закрыты внедрёнными контролями. Дополнительных рекомендаций не сформировано.")
		} else {
			paragraph(pdf, fmt.Sprintf("Выявлено %d уязвимых звеньев, не закрытых ни одним внедрённым средством защиты. Для каждой группы предлагается набор контролей (по матрице vl_category_controls), внедрение которых полностью или частично закроет соответствующие угрозы.", len(recs)))
			pdf.Ln(2)
			bulletHeader(pdf, []string{"№", "Уязвимое звено (VL)", "Затрагивает угроз", "Рекомендуется внедрить"}, []float64{10, 60, 35, 90})
			for i, r := range recs {
				bulletRow(pdf,
					[]string{
						fmt.Sprintf("%d", i+1),
						r.vlName,
						fmt.Sprintf("%d", len(r.threats)),
						r.recommendation,
					},
					[]float64{10, 60, 35, 90},
				)
			}
		}
	}

	renderDocFooter(pdf, "Документ сформирован автоматически системой CyberRisk; рекомендации построены по матрице соответствия VL ↔ control из модели ПТСЗИ.")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ---- группировка контролей по «мероприятиям» из 8.png ----
//
// Каждый из 11 контролей назначен в одну группу по дипломной классификации:
//   АРМ            — защита автоматизированных рабочих мест
//   ЛВС            — защита локальной вычислительной сети
//   Документообор. — защита электронного документооборота
//   Конф. инф.     — защита конфиденциальной информации

type measureGroup struct {
	name         string
	description  string
	controlNames map[string]bool // имена контролей, входящих в группу
}

var measureGroupsCatalog = []measureGroup{
	{
		name:        "Мероприятия по защите автоматизированных рабочих мест (АРМ)",
		description: "Средства защиты конечной точки: антивирус, контроль ПО и доступа, локальная аутентификация.",
		controlNames: map[string]bool{
			"Антивирус":                 true,
			"Программная защита от НСД": true,
			"Системы администрирования": true,
		},
	},
	{
		name:        "Мероприятия по защите локальной вычислительной сети (ЛВС)",
		description: "Средства защиты периметра и внутренней сегментации сети, обнаружение вторжений.",
		controlNames: map[string]bool{
			"Межсетевой экран":              true,
			"Демилитаризованная зона":       true,
			"Honeypot":                      true,
			"Система обнаружения вторжений": true,
			"DDoS-фильтры":                  true,
		},
	},
	{
		name:        "Мероприятия по защите электронного документооборота",
		description: "Подтверждение целостности и подлинности электронных документов, защита каналов передачи.",
		controlNames: map[string]bool{
			"Шифрование трафика": true,
			"Цифровая подпись":   true,
		},
	},
	{
		name:        "Мероприятия по защите конфиденциальной информации",
		description: "Резервное копирование, контроль доступа к информационным активам, защита от утечек.",
		controlNames: map[string]bool{
			"Резервное копирование":     true,
			"Программная защита от НСД": true,
			"Системы администрирования": true,
		},
	},
}

type measureSection struct {
	name        string
	description string
	deployed    []string // внедрённые контроли в этой группе
	missing     []string // ещё не внедрённые
}

func groupControlsByMeasure(deployed []domain.Control) []measureSection {
	deployedNames := map[string]bool{}
	for _, c := range deployed {
		deployedNames[c.Name] = true
	}
	out := make([]measureSection, 0, len(measureGroupsCatalog))
	for _, g := range measureGroupsCatalog {
		ms := measureSection{name: g.name, description: g.description}
		for n := range g.controlNames {
			if deployedNames[n] {
				ms.deployed = append(ms.deployed, n)
			} else {
				ms.missing = append(ms.missing, n)
			}
		}
		out = append(out, ms)
	}
	return out
}

func renderMeasureBlock(pdf gofpdfShim, m measureSection) {
	pdf.SetFont("NotoSans", "B", 11)
	pdf.SetTextColor(40, 40, 40)
	pdf.MultiCell(0, 6, m.name, "", "L", false)
	pdf.SetFont("NotoSans", "", 9)
	pdf.SetTextColor(110, 110, 110)
	pdf.MultiCell(0, 4.5, m.description, "", "L", false)
	pdf.Ln(1)

	if len(m.deployed) == 0 {
		pdf.SetFont("NotoSans", "", 9)
		pdf.SetTextColor(180, 60, 60)
		pdf.MultiCell(0, 4.5, "  ✗ Ни одно средство этой группы не внедрено.", "", "L", false)
	} else {
		for _, n := range m.deployed {
			pdf.SetFont("NotoSans", "", 9)
			pdf.SetTextColor(40, 130, 60)
			pdf.MultiCell(0, 4.5, "  ✓ "+n, "", "L", false)
		}
	}
	for _, n := range m.missing {
		pdf.SetFont("NotoSans", "", 9)
		pdf.SetTextColor(150, 110, 40)
		pdf.MultiCell(0, 4.5, "  + рекомендуется внедрить: "+n, "", "L", false)
	}
	pdf.Ln(3)
}

// ---- адресные рекомендации (по непокрытым VL) ----

type recommendation struct {
	vlName         string
	threats        map[string]struct{}
	controls       map[string]struct{}
	recommendation string
}

func buildRecommendations(paths []domain.AttackPath) []recommendation {
	uncovered := map[int16]*recommendation{}
	for _, p := range paths {
		for _, vl := range p.VulnerableLinks {
			if !vl.Uncovered && len(vl.CoverageControls) > 0 {
				continue
			}
			r, ok := uncovered[vl.CategoryID]
			if !ok {
				r = &recommendation{
					vlName:   coalesce(vl.Code+" — "+vl.Name, vl.Name),
					threats:  map[string]struct{}{},
					controls: map[string]struct{}{},
				}
				uncovered[vl.CategoryID] = r
			}
			r.threats[p.Threat.Name] = struct{}{}
			// все контроли, известные как покрывающие эту VL — даже если ни один не внедрён
			for _, c := range vl.CoverageControls {
				r.controls[c.Name] = struct{}{}
			}
		}
	}

	out := make([]recommendation, 0, len(uncovered))
	for _, r := range uncovered {
		names := make([]string, 0, len(r.controls))
		for n := range r.controls {
			names = append(names, n)
		}
		if len(names) == 0 {
			r.recommendation = "Уточнить набор СЗИ по матрице VL ↔ control"
		} else {
			r.recommendation = joinControls(names)
		}
		out = append(out, *r)
	}
	return out
}

func joinControls(names []string) string {
	if len(names) == 0 {
		return "—"
	}
	if len(names) == 1 {
		return names[0]
	}
	out := names[0]
	for i := 1; i < len(names); i++ {
		out += ", " + names[i]
	}
	return out
}

func countUncoveredVL(paths []domain.AttackPath) int {
	seen := map[int16]bool{}
	for _, p := range paths {
		for _, vl := range p.VulnerableLinks {
			if vl.Uncovered || len(vl.CoverageControls) == 0 {
				seen[vl.CategoryID] = true
			}
		}
	}
	return len(seen)
}
