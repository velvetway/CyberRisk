package risk

import "testing"

func TestCWEToVLCode_KnownClasses(t *testing.T) {
	cases := map[string]string{
		"CWE-94":   "VL1",
		"cwe-78":   "VL1",
		"CWE-506":  "VL3",
		"CWE-269":  "VL4",
		"CWE-200":  "VL6",
		"CWE-79":   "VL6",
		"CWE-22":   "VL6",
		"CWE-119":  "VL2", // memory bug → fallback
		"CWE-9999": "VL2", // unknown → fallback
		"":         "",
		"junk":     "VL2",
	}
	for in, want := range cases {
		if got := CWEToVLCode(in); got != want {
			t.Errorf("CWEToVLCode(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestCWEToVLCodeWithFallback_PrefersSpecific(t *testing.T) {
	// VL3 (specific) wins over VL2 (default).
	got := CWEToVLCodeWithFallback([]string{"CWE-119", "CWE-506", "CWE-787"})
	if got != "VL3" {
		t.Errorf("expected VL3 (specific wins), got %q", got)
	}
}

func TestCWEToVLCodeWithFallback_AllDefault(t *testing.T) {
	got := CWEToVLCodeWithFallback([]string{"CWE-119", "CWE-787", "CWE-9999"})
	if got != "VL2" {
		t.Errorf("expected VL2, got %q", got)
	}
}

func TestCWEToVLCodeWithFallback_Empty(t *testing.T) {
	if got := CWEToVLCodeWithFallback(nil); got != "VL2" {
		t.Errorf("expected VL2 for empty list, got %q", got)
	}
}
