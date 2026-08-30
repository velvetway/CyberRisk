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
