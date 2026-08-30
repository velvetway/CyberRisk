package optimizer

import "math"

// Приведение затрат и старение защиты — две поправки, которые становятся
// заметны только на длинном горизонте. Обе по умолчанию выключены: молча
// менять числа, которые оператор уже видел, нельзя, а осмысленного значения
// «по умолчанию» ни у ставки дисконтирования, ни у скорости деградации нет.

const (
	// maxDiscountRate ограничивает ставку разумным пределом: всё, что выше,
	// означает ошибку ввода, а не экономическое допущение.
	maxDiscountRate = 0.5
	// maxDegradationRate — то же для скорости старения защиты.
	maxDegradationRate = 0.5
)

// discountFactor — во сколько раз рубль года year дешевле сегодняшнего.
//
// Годы считаются от нуля: затраты первого года не дисконтируются, потому
// что тратятся сейчас.
func discountFactor(rate float64, year int) float64 {
	if rate <= 0 || year <= 0 {
		return 1
	}
	return 1 / math.Pow(1+rate, float64(year))
}

// presentValue — приведённая стоимость закупок по годам.
//
// Нужна, чтобы сравнивать планы с разным распределением трат во времени:
// миллион, потраченный на третий год, обходится дешевле миллиона сегодня,
// и при выборе между «купить всё сразу» и «растянуть» это имеет значение.
func presentValue(periods []Period, rate float64) float64 {
	total := 0.0
	for i, p := range periods {
		total += p.Spent * discountFactor(rate, i)
	}
	return total
}

// degradedEffectiveness — эффективность средства через monthsInService
// месяцев работы.
//
// Защита стареет: сигнатуры отстают от новых образцов вредоносного ПО,
// правила фильтрации — от новых техник обхода. В отличие от истечения
// сертификата это не документальный факт, а экспертное допущение, поэтому
// параметр задаётся явно и по умолчанию равен нулю.
func degradedEffectiveness(base, yearlyRate float64, monthsInService int) float64 {
	if yearlyRate <= 0 || monthsInService <= 0 {
		return base
	}
	years := float64(monthsInService) / 12.0
	return clamp01(base * math.Pow(1-yearlyRate, years))
}

func normalizeRate(rate, max float64) float64 {
	if rate < 0 || rate > max {
		return 0
	}
	return rate
}
