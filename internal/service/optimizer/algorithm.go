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

// applyCandidate отмечает как внедрённые все методы, которые закрывает
// средство, а не только тот, под который оно выбрано.
//
// Иначе покупка недооценивается: Dionis DPS, взятый ради демилитаризованной
// зоны, попутно закрывает межсетевой экран и обнаружение вторжений, и не
// учитывать это — значит предлагать докупать уже закрытое.
func applyCandidate(added map[string]float64, c Candidate) {
	codes := c.CoveredControls
	if len(codes) == 0 {
		codes = []string{c.ControlCode}
	}
	for _, code := range codes {
		if e, ok := added[code]; !ok || c.Effectiveness > e {
			added[code] = c.Effectiveness
		}
	}
}

// pathSet — набор сценариев актива. Отдельный тип нужен планировщику
// дорожной карты: он считает W многократно, для каждого месяца горизонта.
type pathSet []domain.PTSZIAttackPath

func (p pathSet) totalW(added map[string]float64) float64 {
	return totalW(p, added)
}

// activeControls — какие методы работают на активе: уже внедрённые плюс
// добавляемые планом. Нужны, чтобы правила совместимости знали полную
// картину, а не только новые закупки.
func activeControls(paths pathSet, added map[string]float64) map[string]bool {
	active := make(map[string]bool, len(added)+4)
	for code := range added {
		active[code] = true
	}
	for i := range paths {
		for _, vl := range paths[i].VulnerableLinks {
			for _, c := range vl.Controls {
				if c.Implemented {
					active[c.Control.Code] = true
				}
			}
		}
	}
	return active
}

// totalW — суммарный вес всех применимых угроз актива при заданном наборе
// дополнительно внедряемых мер.
//
// Суммой, а не средним: добавление меры не должно «размывать» эффект по числу
// угроз, а общая подверженность актива складывается из всех сценариев.
func totalW(paths []domain.PTSZIAttackPath, added map[string]float64) float64 {
	// Эффективность новых мер зависит от того, что уже работает рядом:
	// обнаружение вторжений без администрирования слабее, межсетевой экран
	// вместе с сегментацией — сильнее.
	if len(added) > 0 {
		added = applyCompatibility(added, activeControls(paths, added))
	}

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
			cost := c.TotalCost
			if spent+cost > budget {
				continue
			}

			trial := make(map[string]float64, len(added)+len(c.CoveredControls)+1)
			for k, v := range added {
				trial[k] = v
			}
			applyCandidate(trial, c)
			w := totalW(paths, trial)

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
		applyCandidate(added, c)
		spent += c.TotalCost

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
			cost += candidates[i].TotalCost
			if cost > budget {
				break
			}
			// Средство закрывает все свои методы; если один метод закрывают
			// несколько средств, засчитывается лучшее из них.
			applyCandidate(added, candidates[i])
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
		if candidates[i].TotalCost != candidates[j].TotalCost {
			return candidates[i].TotalCost < candidates[j].TotalCost
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
