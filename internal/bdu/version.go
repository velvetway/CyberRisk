package bdu

import (
	"regexp"
	"strconv"
	"strings"
)

// VersionMatches проверяет, попадает ли установленная на активе версия ПО
// под версию-критерий из БДУ ФСТЭК.
//
// БДУ хранит version в свободном текстовом виде. Поддерживаемые формы:
//
//	"-" / ""                                 — версия не указана, БДУ-запись
//	                                            релевантна всем версиям этого ПО.
//	"1.7", "8.0.21", "20.04 LTS"             — точная версия (case-insensitive).
//	"1.6 «Смоленск»", "15 SP6"               — точная именованная версия.
//	"от X до Y включительно"                 — closed range [X, Y].
//	"от X до Y"                              — half-open [X, Y).
//	"до Y включительно", "до Y"              — (-∞, Y] или (-∞, Y).
//	"от X включительно", "от X"              — [X, +∞) или (X, +∞).
//
// Всё, что не подошло под exact или range, → false (no match).
//
// Семантика для пустого assetVer: оператор не указал версию на активе →
// возвращаем true (conservative — лучше вытащить лишнее, чем спрятать).
func VersionMatches(bduVer, assetVer string) bool {
	bduVer = strings.TrimSpace(bduVer)
	assetVer = strings.TrimSpace(assetVer)

	if bduVer == "" || bduVer == "-" {
		return true
	}
	if assetVer == "" {
		return true
	}
	if strings.EqualFold(bduVer, assetVer) {
		return true
	}
	if r, ok := parseRussianRange(bduVer); ok {
		return r.contains(assetVer)
	}
	return false
}

// =====================================================================
// Range parsing
// =====================================================================

type versionRange struct {
	from         string // "" если без нижней границы
	fromInclusive bool
	to           string // "" если без верхней границы
	toInclusive   bool
}

var (
	reFromTo       = regexp.MustCompile(`^от\s+(\S+)\s+до\s+(\S+)(\s+включительно)?$`)
	reTo           = regexp.MustCompile(`^до\s+(\S+)(\s+включительно)?$`)
	reFrom         = regexp.MustCompile(`^от\s+(\S+)(\s+включительно)?$`)
)

func parseRussianRange(s string) (versionRange, bool) {
	if m := reFromTo.FindStringSubmatch(s); m != nil {
		return versionRange{
			from: m[1], fromInclusive: true,
			to: m[2], toInclusive: m[3] != "",
		}, true
	}
	if m := reTo.FindStringSubmatch(s); m != nil {
		return versionRange{
			to: m[1], toInclusive: m[2] != "",
		}, true
	}
	if m := reFrom.FindStringSubmatch(s); m != nil {
		return versionRange{
			from: m[1], fromInclusive: m[2] != "",
		}, true
	}
	return versionRange{}, false
}

func (r versionRange) contains(v string) bool {
	if r.from != "" {
		c := compareVer(v, r.from)
		if r.fromInclusive {
			if c < 0 {
				return false
			}
		} else {
			if c <= 0 {
				return false
			}
		}
	}
	if r.to != "" {
		c := compareVer(v, r.to)
		if r.toInclusive {
			if c > 0 {
				return false
			}
		} else {
			if c >= 0 {
				return false
			}
		}
	}
	return true
}

// =====================================================================
// Comparator — semver-like, по компонентам через '.'.
//
// Идея: "8.4.10" vs "8.4.16" → split, compare как int если можно, иначе
// как string. Если длины разные, недостающие компоненты считаем "0".
// =====================================================================

func compareVer(a, b string) int {
	if a == b {
		return 0
	}
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	n := len(ap)
	if len(bp) > n {
		n = len(bp)
	}
	for i := 0; i < n; i++ {
		var aTok, bTok string = "0", "0"
		if i < len(ap) {
			aTok = ap[i]
		}
		if i < len(bp) {
			bTok = bp[i]
		}
		ai, aErr := atoiPrefix(aTok)
		bi, bErr := atoiPrefix(bTok)
		if aErr == nil && bErr == nil {
			if ai != bi {
				if ai < bi {
					return -1
				}
				return 1
			}
			// числовые равны — сравним строки целиком (возможно "8a" vs "8b")
			if aTok != bTok {
				if aTok < bTok {
					return -1
				}
				return 1
			}
			continue
		}
		// один или оба нечисловые — побайтовое сравнение
		if aTok != bTok {
			if aTok < bTok {
				return -1
			}
			return 1
		}
	}
	return 0
}

// atoiPrefix парсит ведущие цифры компонента — "5", "10rc1" → 5, 10.
// Если префикс не цифра, возвращает ошибку.
func atoiPrefix(s string) (int, error) {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(s[:end])
}
