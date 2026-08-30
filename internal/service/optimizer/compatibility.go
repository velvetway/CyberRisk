package optimizer

import "sort"

// Совместимость средств защиты.
//
// Важное наблюдение о природе задачи: между одиннадцатью методами ПТСЗИ
// прямых конфликтов почти нет — они по построению дополняют друг друга.
// Конфликты возникают между конкретными продуктами (два антивируса на одной
// машине, два средства защиты от НСД с несовместимыми драйверами), но модель
// работает на уровне методов и берёт по одному средству на метод, поэтому
// такие столкновения исключены самой постановкой.
//
// Что действительно есть между методами — это зависимости и взаимное
// усиление. Их и моделируем.
//
// Все коэффициенты ниже экспертные. Они намеренно скромные: задача не
// подкрутить результат, а отразить качественный факт «эти меры работают
// лучше вместе». Устойчивость плана к их изменению проверяется анализом
// чувствительности.

// Dependency — метод, который без опоры работает вполсилы.
type Dependency struct {
	// Control не даёт полного эффекта без RequiresControl.
	Control         string
	RequiresControl string
	// Penalty — во сколько раз падает эффективность при отсутствии опоры.
	Penalty float64
	Reason  string
}

// Synergy — пара методов, усиливающих друг друга.
type Synergy struct {
	A, B string
	// Bonus — во сколько раз растёт эффективность обоих при совместном
	// применении.
	Bonus  float64
	Reason string
}

// dependencies — методы, чья польза зависит от наличия другого.
var dependencies = []Dependency{
	{
		Control:         "IDS",
		RequiresControl: "AD",
		Penalty:         0.6,
		Reason: "система обнаружения вторжений производит события, " +
			"но без централизованного администрирования на них некому реагировать: " +
			"атака будет зафиксирована и не остановлена",
	},
	{
		Control:         "HP",
		RequiresControl: "IDS",
		Penalty:         0.5,
		Reason: "ловушка ценна тем, что выдаёт факт проникновения, " +
			"а заметить срабатывание без системы обнаружения практически невозможно",
	},
	{
		Control:         "TE",
		RequiresControl: "DS",
		Penalty:         0.7,
		Reason: "шифрование канала защищает от прослушивания, но без контроля " +
			"подлинности сторон остаётся уязвимым к подмене участника обмена",
	},
}

// synergies — пары, дающие больше в сочетании, чем по отдельности.
var synergies = []Synergy{
	{
		A: "FW", B: "DZ", Bonus: 1.15,
		Reason: "межсетевой экран задаёт правила, демилитаризованная зона — " +
			"структуру сети, на которой этим правилам есть что разграничивать",
	},
	{
		A: "R", B: "DS", Bonus: 1.10,
		Reason: "резервная копия ценна ровно настолько, насколько можно " +
			"доказать её неизменность: подпись превращает копию в доверенную",
	},
	{
		A: "A", B: "L", Bonus: 1.10,
		Reason: "антивирус ловит известное вредоносное ПО, замкнутая программная " +
			"среда не даёт запуститься неизвестному — вместе они закрывают оба случая",
	},
	{
		A: "IDS", B: "AD", Bonus: 1.20,
		Reason: "обнаружение вторжений вместе с централизованным управлением " +
			"даёт не только сигнал, но и возможность немедленно применить меры",
	},
}

// applyCompatibility корректирует эффективность мер с учётом того, какие
// методы работают вместе.
//
// active — коды методов, действующих на активе (уже внедрённых плюс
// запланированных). Возвращает поправленные эффективности.
func applyCompatibility(effectiveness map[string]float64, active map[string]bool) map[string]float64 {
	out := make(map[string]float64, len(effectiveness))
	for code, e := range effectiveness {
		out[code] = e
	}

	// Сначала штрафы: мера без своей опоры работает хуже.
	for _, d := range dependencies {
		if _, planned := out[d.Control]; !planned {
			continue
		}
		if !active[d.RequiresControl] {
			out[d.Control] = clamp01(out[d.Control] * d.Penalty)
		}
	}

	// Затем бонусы: пара, работающая вместе, усиливает обе стороны.
	for _, s := range synergies {
		if !active[s.A] || !active[s.B] {
			continue
		}
		if _, ok := out[s.A]; ok {
			out[s.A] = clamp01(out[s.A] * s.Bonus)
		}
		if _, ok := out[s.B]; ok {
			out[s.B] = clamp01(out[s.B] * s.Bonus)
		}
	}

	return out
}

// CompatibilityNote — пояснение, почему эффективность меры изменилась.
type CompatibilityNote struct {
	Kind    string  `json:"kind"` // dependency | synergy
	Control string  `json:"control"`
	Related string  `json:"related"`
	Factor  float64 `json:"factor"`
	Reason  string  `json:"reason"`
}

// compatibilityNotes перечисляет сработавшие правила для набора методов.
//
// Нужны, чтобы план не выглядел результатом непрозрачной подкрутки: рядом с
// каждой поправкой видно, из какого содержательного соображения она следует.
func compatibilityNotes(planned map[string]bool, active map[string]bool) []CompatibilityNote {
	notes := make([]CompatibilityNote, 0)

	for _, d := range dependencies {
		if planned[d.Control] && !active[d.RequiresControl] {
			notes = append(notes, CompatibilityNote{
				Kind:    "dependency",
				Control: d.Control,
				Related: d.RequiresControl,
				Factor:  d.Penalty,
				Reason:  d.Reason,
			})
		}
	}

	for _, s := range synergies {
		if active[s.A] && active[s.B] && (planned[s.A] || planned[s.B]) {
			notes = append(notes, CompatibilityNote{
				Kind:    "synergy",
				Control: s.A,
				Related: s.B,
				Factor:  s.Bonus,
				Reason:  s.Reason,
			})
		}
	}

	sort.Slice(notes, func(i, j int) bool {
		if notes[i].Kind != notes[j].Kind {
			return notes[i].Kind < notes[j].Kind
		}
		return notes[i].Control < notes[j].Control
	})
	return notes
}
