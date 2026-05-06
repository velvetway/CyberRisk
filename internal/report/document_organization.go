package report

import (
	"bytes"
	"fmt"
	"sort"

	"Diplom/internal/domain"
)

// GenerateOrganizationReportPDF — сводный отчёт по всей организации.
// 4 раздела:
//   1. Обзорные метрики (количество активов, распределение, max W)
//   2. Сводное соответствие по стандартам (avg/min/max %)
//   3. Таблица активов с показателями (W_max, контроли, compliance)
//   4. Топ критических рисков (asset × threat с наибольшим W)
func GenerateOrganizationReportPDF(
	overview *domain.OrganizationOverview,
	matrix []domain.AssetMatrixRow,
	critical []domain.CriticalRisk,
) ([]byte, error) {
	if overview == nil {
		return nil, fmt.Errorf("nil overview")
	}
	pdf := newPDF()
	pdf.AddPage()
	renderDocHeader(pdf, "Сводный отчёт по организации", "все активы")

	// ----- Раздел 1: Метрики -----
	sectionTitle(pdf, "1. Обзорные метрики")
	row(pdf, "Всего активов", fmt.Sprintf("%d", overview.TotalAssets))
	row(pdf, "Изолированных активов (Z = 0.5)", fmt.Sprintf("%d", overview.IsolatedAssets))
	row(pdf, "Внедрено СЗИ (всего пар актив-control)", fmt.Sprintf("%d", overview.TotalControls))
	row(pdf, "Непокрытых VL по угрозам", fmt.Sprintf("%d", overview.UncoveredVLs))
	row(pdf, "Максимальный риск W", fmt.Sprintf("%.2f", overview.WMax))
	if overview.WMaxAsset != "" {
		row(pdf, "    актив-источник", overview.WMaxAsset)
	}
	if overview.WMaxThreat != "" {
		row(pdf, "    угроза-источник", overview.WMaxThreat)
	}
	row(pdf, "Средний W_max по активам", fmt.Sprintf("%.2f", overview.AvgWPerAsset))
	pdf.Ln(2)

	// Распределение по уровням риска
	if total := overview.RiskDistribution["critical"] + overview.RiskDistribution["high"] +
		overview.RiskDistribution["medium"] + overview.RiskDistribution["low"]; total > 0 {
		sectionTitle(pdf, "Распределение активов по уровню риска")
		bulletHeader(pdf, []string{"Уровень", "Активов", "Доля"}, []float64{60, 40, 40})
		for _, lvl := range []string{"critical", "high", "medium", "low"} {
			n := overview.RiskDistribution[lvl]
			pct := 0.0
			if total > 0 {
				pct = 100.0 * float64(n) / float64(total)
			}
			bulletRow(pdf, []string{
				levelLabel(lvl),
				fmt.Sprintf("%d", n),
				fmt.Sprintf("%.1f%%", pct),
			}, []float64{60, 40, 40})
		}
		pdf.Ln(2)
	}

	// Распределение по типам активов
	if len(overview.AssetsByType) > 0 {
		sectionTitle(pdf, "Распределение по типам активов")
		bulletHeader(pdf, []string{"Тип", "Активов"}, []float64{120, 30})
		for _, b := range overview.AssetsByType {
			bulletRow(pdf, []string{b.TypeName, fmt.Sprintf("%d", b.Count)}, []float64{120, 30})
		}
		pdf.Ln(2)
	}

	// ----- Раздел 2: Compliance -----
	sectionTitle(pdf, "2. Соответствие стандартам ИБ")
	if len(overview.ComplianceByStd) == 0 {
		paragraph(pdf, "Стандарты не настроены или нет активов для оценки.")
	} else {
		bulletHeader(pdf, []string{"Стандарт", "Среднее %", "Минимум %", "Максимум %", "Активов"}, []float64{70, 30, 30, 30, 25})
		for _, s := range overview.ComplianceByStd {
			bulletRow(pdf, []string{
				s.Standard.Name,
				fmt.Sprintf("%.0f%%", s.AvgScore*100),
				fmt.Sprintf("%.0f%%", s.MinScore*100),
				fmt.Sprintf("%.0f%%", s.MaxScore*100),
				fmt.Sprintf("%d", s.AssetsCount),
			}, []float64{70, 30, 30, 30, 25})
		}
	}

	// ----- Раздел 3: Таблица активов -----
	pdf.AddPage()
	renderDocHeader(pdf, "Таблица активов", "сводный отчёт по организации")

	sectionTitle(pdf, "3. Активы и их показатели")
	bulletHeader(pdf,
		[]string{"Актив", "Тип", "Среда", "W_max", "Уровень", "Угроз", "СЗИ"},
		[]float64{55, 30, 18, 18, 22, 17, 15},
	)
	for _, m := range matrix {
		env := m.Environment
		if env == "" {
			env = "—"
		}
		bulletRow(pdf,
			[]string{
				m.Name,
				coalesce(m.TypeName, "—"),
				env,
				fmt.Sprintf("%.2f", m.WMax),
				levelLabel(m.Level),
				fmt.Sprintf("%d", m.ThreatCount),
				fmt.Sprintf("%d", m.ControlCount),
			},
			[]float64{55, 30, 18, 18, 22, 17, 15},
		)
	}

	// ----- Раздел 4: Критические риски -----
	if len(critical) > 0 {
		pdf.AddPage()
		renderDocHeader(pdf, "Критические риски", "топ по величине W")

		sectionTitle(pdf, "4. Критические риски")
		paragraph(pdf, fmt.Sprintf("Топ-%d пар (актив × угроза) с наибольшим расчётным риском W. Это конкретные комбинации, которые требуют первоочередного внимания: либо за счёт внедрения дополнительных контролей, либо за счёт переоценки применимости угрозы.",
			len(critical)))
		pdf.Ln(2)
		bulletHeader(pdf,
			[]string{"#", "Актив", "Угроза", "БДУ", "W", "Уровень"},
			[]float64{8, 55, 60, 22, 18, 22},
		)

		// (CriticalRisks уже отсортирован service'ом; на всякий случай — ещё раз)
		sort.SliceStable(critical, func(i, j int) bool { return critical[i].W > critical[j].W })

		for i, r := range critical {
			bulletRow(pdf,
				[]string{
					fmt.Sprintf("%d", i+1),
					r.AssetName,
					r.ThreatName,
					r.BDUID,
					fmt.Sprintf("%.2f", r.W),
					levelLabel(r.Level),
				},
				[]float64{8, 55, 60, 22, 18, 22},
			)
		}
	}

	renderDocFooter(pdf, "Документ сформирован автоматически системой CyberRisk на основе данных всех активов организации.")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
