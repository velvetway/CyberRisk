package optimizer

import (
	"math"
	"sort"

	"Diplom/internal/domain"
	"Diplom/internal/service/ptszi"
)

// maxExhaustiveCandidates — предел, до которого считается точный оптимум.
// Перебор экспоненциален: при 18 кандидатах это 262144 комбинации, что ещё
// укладывается в доли секунды, дальше расти нельзя.
const maxExhaustiveCandidates = 18

// recalcW пересчитывает вес угрозы, считая внедрёнными меры из added
// (код метода → эффективность внедрения) в дополнение к уже имеющимся.
//
// Повторяет модель покрытия из ptszi: покрытие звена — вероятность, что
// сработала хотя бы одна мера, Q_reaction — среднее покрытие по звеньям.
// Формула W берётся напрямую из ptszi.CalculateW, чтобы оптимизатор и
// основной расчёт не разошлись.
func recalcW(path *domain.PTSZIAttackPath, added map[string]float64) float64 {
	if len(path.VulnerableLinks) == 0 {
		return ptszi.CalculateW(path.QThreat, path.QSeverity, 0, path.Z)
	}

	sum := 0.0
	for _, vl := range path.VulnerableLinks {
		product := 1.0
		for _, c := range vl.Controls {
			effectiveness := 0.0
			switch {
			case c.Implemented:
				effectiveness = c.Effectiveness
			default:
				if e, ok := added[c.Control.Code]; ok {
					effectiveness = e
				}
			}
			if effectiveness <= 0 {
				continue
			}
			product *= 1 - clamp01(c.Coverage*effectiveness)
		}
		sum += clamp01(1 - product)
	}

	qReaction := clamp01(sum / float64(len(path.VulnerableLinks)))
	return ptszi.CalculateW(path.QThreat, path.QSeverity, qReaction, path.Z)
}

// pathSet — набор сценариев актива. Отдельный тип нужен планировщику
// дорожной карты: он считает W многократно, для каждого месяца горизонта.
type pathSet []domain.PTSZIAttackPath

func (p pathSet) totalW(added map[string]float64) float64 {
	return totalW(p, added)
}

// totalW — суммарный вес всех применимых угроз актива при заданном наборе
// дополнительно внедряемых мер.
//
// Суммой, а не средним: добавление меры не должно «размывать» эффект по числу
// угроз, а общая подверженность актива складывается из всех сценариев.
func totalW(paths []domain.PTSZIAttackPath, added map[string]float64) float64 {
	sum := 0.0
	for i := range paths {
		sum += recalcW(&paths[i], added)
	}
	return sum
}

// greedy набирает комплекс, каждый раз добавляя меру с лучшим отношением
// «снижение риска к стоимости».
//
// Жадный выбор не гарантирует оптимума: мера с лучшей отдачей на рубль может
// съесть бюджет, оставив без денег пару мер, которые вместе дали бы больше.
// Зато результат объясним оператору — «следующий рубль лучше всего потратить
// сюда», — и именно это чаще всего требуется на практике.
func greedy(paths []domain.PTSZIAttackPath, candidates []Candidate, budget float64) ([]Step, float64) {
	added := map[string]float64{}
	used := map[int]bool{}
	steps := make([]Step, 0, len(candidates))

	current := totalW(paths, added)
	spent := 0.0

	for {
		bestIdx := -1
		bestEfficiency := 0.0
		bestW := current

		for i, c := range candidates {
			if used[i] {
				continue
			}
			cost := c.CostMax
			if spent+cost > budget {
				continue
			}

			added[c.ControlCode] = c.Effectiveness
			w := totalW(paths, added)
			delete(added, c.ControlCode)

			delta := current - w
			if delta <= 1e-9 {
				continue
			}
			// Отдача на миллион рублей — величина удобного порядка.
			efficiency := delta / (cost / 1_000_000)
			if efficiency > bestEfficiency {
				bestIdx, bestEfficiency, bestW = i, efficiency, w
			}
		}

		if bestIdx < 0 {
			break
		}

		c := candidates[bestIdx]
		used[bestIdx] = true
		added[c.ControlCode] = c.Effectiveness
		spent += c.CostMax

		steps = append(steps, Step{
			Candidate:      c,
			WBefore:        current,
			WAfter:         bestW,
			DeltaW:         current - bestW,
			Efficiency:     bestEfficiency,
			CumulativeCost: spent,
		})
		current = bestW
	}

	return steps, spent
}

// exhaustive перебирает все подмножества кандидатов и возвращает лучшее
// снижение W, укладывающееся в бюджет.
//
// Нужен как эталон: позволяет измерить, насколько жадный алгоритм отстаёт от
// точного оптимума. На реальных размерностях (методов всего 11) перебор
// выполним, поэтому проверка не теоретическая.
func exhaustive(paths []domain.PTSZIAttackPath, candidates []Candidate, budget float64) (float64, bool) {
	n := len(candidates)
	if n == 0 || n > maxExhaustiveCandidates {
		return 0, false
	}

	baseline := totalW(paths, nil)
	best := 0.0

	for mask := 1; mask < (1 << n); mask++ {
		cost := 0.0
		added := make(map[string]float64, n)
		for i := 0; i < n; i++ {
			if mask&(1<<i) == 0 {
				continue
			}
			cost += candidates[i].CostMax
			if cost > budget {
				break
			}
			// Один метод могут закрывать несколько средств: засчитываем лучшее,
			// иначе комбинация выглядела бы сильнее, чем есть.
			if e, ok := added[candidates[i].ControlCode]; !ok || candidates[i].Effectiveness > e {
				added[candidates[i].ControlCode] = candidates[i].Effectiveness
			}
		}
		if cost > budget {
			continue
		}
		if delta := baseline - totalW(paths, added); delta > best {
			best = delta
		}
	}

	return best, true
}

// sortCandidates даёт устойчивый порядок: сначала дешёвые, потом по коду.
// Без него результат перебора мог бы плавать между запусками при равных
// значениях целевой функции.
func sortCandidates(candidates []Candidate) {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].CostMax != candidates[j].CostMax {
			return candidates[i].CostMax < candidates[j].CostMax
		}
		return candidates[i].ControlCode < candidates[j].ControlCode
	})
}

func clamp01(v float64) float64 {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
