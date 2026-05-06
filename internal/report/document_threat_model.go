package report

import (
	"bytes"
	"fmt"
	"strings"

	"Diplom/internal/domain"
)

// GenerateThreatModelPDF — «Модель угроз информационной безопасности» актива.
// Соответствует блоку «Формирование модели угроз ИБ» из 7.png диплома и
// результатам в 8.png. Для каждой применимой угрозы расписывается:
//   • источник (S), способ реализации (ST), уязвимое звено (VL),
//     деструктивные действия (DA);
//   • степень реализации, опасность, реакция, контурный коэффициент Z;
//   • расчётный риск W и его уровень.
func GenerateThreatModelPDF(asset *domain.Asset, resp *domain.AssetAttackPathsResponse) ([]byte, error) {
	if asset == nil {
		return nil, fmt.Errorf("nil asset")
	}
	pdf := newPDF()
	pdf.AddPage()

	renderDocHeader(pdf, "Модель угроз информационной безопасности", asset.Name)

	// 0. Преамбула
	sectionTitle(pdf, "1. Назначение документа")
	paragraph(pdf, "Настоящий документ описывает модель угроз информационной безопасности для указанного объекта защиты, сформированную на основании справочника угроз ФСТЭК России (БДУ) и графа атак ПТСЗИ (S → ST → VL → DA). Для каждой применимой угрозы рассчитан риск W в соответствии с методологией, описанной в разделе «Модель риска ПТСЗИ» проекта.")

	sectionTitle(pdf, "2. Расчётная формула")
	paragraph(pdf, "W = ( Q^угроза + q^серьёзность + ( 1 − Q^реакция ) ) / 3 · Z,   где W ∈ [0, 1].")
	paragraph(pdf, "Q^угроза — степень реализации угрозы (из справочника БДУ). q^серьёзность — степень опасности (из БДУ). Q^реакция — доля уязвимых звеньев, закрытых внедрёнными на активе средствами защиты. Z = 0.5 (изолированный контур) либо 1.0 (общий контур).")

	if resp == nil || len(resp.Paths) == 0 {
		sectionTitle(pdf, "3. Применимые угрозы")
		paragraph(pdf, "Для данного актива применимых угроз не выявлено (либо база угроз ещё не сопоставлена с типом актива).")
		renderDocFooter(pdf, "Документ сформирован автоматически системой CyberRisk на основе данных об активе и базы угроз ФСТЭК.")
		var buf bytes.Buffer
		if err := pdf.Output(&buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	// 3. Сводный показатель
	sectionTitle(pdf, "3. Агрегированный риск")
	row(pdf, "W max", fmt.Sprintf("%.2f", resp.Aggregate.WMax))
	row(pdf, "Уровень риска", levelLabel(resp.Aggregate.Level))
	row(pdf, "Учтено угроз", fmt.Sprintf("%d", resp.Aggregate.ThreatCount))
	row(pdf, "Непокрытых VL", fmt.Sprintf("%d", resp.Aggregate.UncoveredCount))
	pdf.Ln(2)

	// 4. Перечень угроз
	sectionTitle(pdf, "4. Перечень применимых угроз")
	bulletHeader(pdf,
		[]string{"№", "БДУ", "Угроза", "W", "Уровень"},
		[]float64{10, 25, 110, 18, 25},
	)
	for i, p := range resp.Paths {
		bulletRow(pdf,
			[]string{
				fmt.Sprintf("%d", i+1),
				p.Threat.BDUID,
				p.Threat.Name,
				fmt.Sprintf("%.2f", p.W),
				levelLabel(p.Level),
			},
			[]float64{10, 25, 110, 18, 25},
		)
	}

	// 5. Детализация по каждой угрозе (топ-N с W)
	const detailLimit = 5
	limit := detailLimit
	if len(resp.Paths) < limit {
		limit = len(resp.Paths)
	}
	for i := 0; i < limit; i++ {
		p := resp.Paths[i]
		pdf.AddPage()
		renderDocHeader(pdf, fmt.Sprintf("Угроза %d. %s", i+1, p.Threat.Name), asset.Name)

		sectionTitle(pdf, "Идентификация")
		row(pdf, "Код БДУ", coalesce(p.Threat.BDUID, "—"))
		row(pdf, "Расчётный риск W", fmt.Sprintf("%.2f (%s)", p.W, levelLabel(p.Level)))
		pdf.Ln(2)

		sectionTitle(pdf, "Параметры формулы W")
		row(pdf, "Q^угроза (реализация)", fmt.Sprintf("%.2f", p.QThreat))
		row(pdf, "q^серьёзность (опасность)", fmt.Sprintf("%.2f", p.QSeverity))
		row(pdf, "Q^реакция (предотвращение)", fmt.Sprintf("%.2f", p.QReaction))
		row(pdf, "Z (контур)", fmt.Sprintf("%.2f", p.Z))
		pdf.Ln(2)

		sectionTitle(pdf, "Источники угрозы (S)")
		if len(p.Sources) == 0 {
			paragraph(pdf, "Источники не указаны.")
		} else {
			parts := make([]string, 0, len(p.Sources))
			for _, s := range p.Sources {
				parts = append(parts, s.Name)
			}
			paragraph(pdf, strings.Join(parts, "; "))
		}

		sectionTitle(pdf, "Уязвимые звенья (VL)")
		if len(p.VulnerableLinks) == 0 {
			paragraph(pdf, "VL-связи отсутствуют.")
		} else {
			bulletHeader(pdf, []string{"Код", "Уязвимое звено", "Покрыто контролями"}, []float64{18, 80, 90})
			for _, vl := range p.VulnerableLinks {
				covers := []string{}
				for _, c := range vl.CoverageControls {
					covers = append(covers, c.Name)
				}
				covStr := strings.Join(covers, ", ")
				if covStr == "" {
					covStr = "— (не закрыто)"
				}
				bulletRow(pdf, []string{vl.Code, vl.Name, covStr}, []float64{18, 80, 90})
			}
		}
		pdf.Ln(2)

		sectionTitle(pdf, "Деструктивные действия (DA)")
		if len(p.DestructiveActions) == 0 {
			paragraph(pdf, "Деструктивные действия не определены.")
		} else {
			parts := make([]string, 0, len(p.DestructiveActions))
			for _, da := range p.DestructiveActions {
				parts = append(parts, da.Name)
			}
			paragraph(pdf, strings.Join(parts, "; "))
		}
	}

	if len(resp.Paths) > detailLimit {
		pdf.Ln(4)
		paragraph(pdf, fmt.Sprintf("Подробное описание представлено для первых %d угроз с наибольшим W. Остальные %d угроз приведены в сводном перечне (раздел 4).", detailLimit, len(resp.Paths)-detailLimit))
	}

	renderDocFooter(pdf, "Документ сформирован автоматически системой CyberRisk на основе модели угроз ПТСЗИ и БДУ ФСТЭК.")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
