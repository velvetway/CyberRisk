package bdu

import "testing"

func TestVersionMatches_EmptyOrDash(t *testing.T) {
	cases := []string{"", "-"}
	for _, b := range cases {
		if !VersionMatches(b, "1.7") {
			t.Errorf("BDU %q must match any asset version", b)
		}
		if !VersionMatches(b, "") {
			t.Errorf("BDU %q + empty asset must match", b)
		}
	}
}

func TestVersionMatches_EmptyAssetVersion(t *testing.T) {
	// Оператор не указал версию — берём всё.
	if !VersionMatches("8.4.10", "") {
		t.Error("empty asset version must accept any BDU version")
	}
}

func TestVersionMatches_Exact(t *testing.T) {
	if !VersionMatches("1.7", "1.7") {
		t.Error("exact equal must match")
	}
	if !VersionMatches("1.6 «Смоленск»", "1.6 «Смоленск»") {
		t.Error("exact named version must match")
	}
	if VersionMatches("1.7", "1.6") {
		t.Error("different exact versions must not match")
	}
	if VersionMatches("1.6 «Смоленск»", "1.6") {
		t.Error("named version is not equal to bare semver — must NOT match")
	}
}

func TestVersionMatches_RangeFromTo_Inclusive(t *testing.T) {
	bdu := "от 8.4.0 до 8.4.16 включительно"
	matches := []string{"8.4.0", "8.4.10", "8.4.16"}
	misses := []string{"8.3.99", "8.4.17", "9.0.0"}
	for _, m := range matches {
		if !VersionMatches(bdu, m) {
			t.Errorf("%q must match %q", m, bdu)
		}
	}
	for _, m := range misses {
		if VersionMatches(bdu, m) {
			t.Errorf("%q must NOT match %q", m, bdu)
		}
	}
}

func TestVersionMatches_RangeFromTo_Exclusive(t *testing.T) {
	// "от X до Y" — half-open (без "включительно" → < Y)
	bdu := "от 9.0 до 9.5"
	if VersionMatches(bdu, "9.5") {
		t.Errorf("9.5 must NOT match %q (no inclusive)", bdu)
	}
	if !VersionMatches(bdu, "9.4") {
		t.Errorf("9.4 must match %q", bdu)
	}
}

func TestVersionMatches_RangeTo(t *testing.T) {
	if !VersionMatches("до 8.4.20 включительно", "8.4.10") {
		t.Error("8.4.10 ≤ 8.4.20 must match")
	}
	if VersionMatches("до 8.4.20 включительно", "8.4.21") {
		t.Error("8.4.21 > 8.4.20 must not match")
	}
	if !VersionMatches("до 4 включительно", "3.9") {
		t.Error("3.9 ≤ 4 must match")
	}
	if VersionMatches("до 4 включительно", "5") {
		t.Error("5 > 4 must not match")
	}
}

func TestVersionMatches_RangeFrom(t *testing.T) {
	if !VersionMatches("от 1.5", "1.6") {
		t.Error("1.6 > 1.5 must match")
	}
	if VersionMatches("от 1.5", "1.5") {
		t.Error("1.5 == 1.5 without inclusive must not match")
	}
	if !VersionMatches("от 1.5 включительно", "1.5") {
		t.Error("1.5 == 1.5 with inclusive must match")
	}
}

func TestCompareVer(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"8.4.10", "8.4.16", -1},
		{"8.4.16", "8.4.10", 1},
		{"8.4", "8.4.0", 0},
		{"8.4.10", "8.4.10", 0},
		{"9", "10", -1},
		{"10", "9", 1},
		{"1.7", "1.6 «Смоленск»", 1}, // числово 7 > 6
		{"7.1(5b)su5", "7.1(5b)su5", 0},
	}
	for _, c := range cases {
		got := compareVer(c.a, c.b)
		// нормализуем -1/+1
		if (got < 0 && c.want >= 0) || (got > 0 && c.want <= 0) || (got == 0 && c.want != 0) {
			t.Errorf("compareVer(%q, %q) = %d; want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestVersionMatches_NoMatchExoticBDU(t *testing.T) {
	// "до 16.01.2023" — date form, не parsable как range, но reTo matches
	// (т.к. \S+ ловит всё). Дальше compareVer сравнит "16.01.2023" с asset
	// как строки — обычно не даст осмысленный результат, но уж точно не
	// «всё подряд». Проверим, что не падаем.
	_ = VersionMatches("до 16.01.2023 включительно", "1.7")
	// Ensure no panic.
}
