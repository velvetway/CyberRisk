// Package risk holds shared classification helpers used by both the import
// pipelines and the runtime risk service. The current contents (CWE → VL
// category mapping) is consumed by the asset_vulnerability service when it
// auto-detects CVEs from installed software, and by the risk service when
// it materializes the S → ST → VL → DA chain.
package risk

import "strings"

// CWEToVLCode maps a CWE identifier (e.g. "CWE-94") to the VL category
// from the diploma it most naturally belongs to. The mapping is deliberately
// coarse: the diploma uses 6 VL buckets, CWE has thousands of entries — we
// only enumerate CWE classes that show up in БДУ ФСТЭК with non-trivial
// frequency, and route everything else to the catch-all VL2 ("устаревшие
// версии ПО или версии, имеющие уязвимости").
//
// Returns "" only when the input is empty or obviously malformed; callers
// should fall back to "VL2" themselves rather than special-casing "".
func CWEToVLCode(cwe string) string {
	c := strings.ToUpper(strings.TrimSpace(cwe))
	if c == "" {
		return ""
	}
	if !strings.HasPrefix(c, "CWE-") {
		return "VL2"
	}

	switch c {
	// VL1 — нештатное дополнительное ПО / неконтролируемое исполнение кода
	case "CWE-94",   // Improper Control of Generation of Code
		"CWE-95",   // Eval-like injection
		"CWE-77",   // Command Injection
		"CWE-78",   // OS Command Injection
		"CWE-114",  // Process Control
		"CWE-829",  // Inclusion of Functionality from Untrusted Control Sphere
		"CWE-918":  // SSRF
		return "VL1"

	// VL3 — недекларируемое ПО / закладки / шпионское поведение
	case "CWE-506", // Embedded Malicious Code
		"CWE-507",  // Trojan Horse
		"CWE-508",  // Non-Replicating Malicious Code
		"CWE-509",  // Replicating Malicious Code (Virus or Worm)
		"CWE-510",  // Trapdoor
		"CWE-511",  // Logic/Time Bomb
		"CWE-912":  // Hidden Functionality
		return "VL3"

	// VL4 — обход политик / повышение привилегий / misconfiguration
	case "CWE-264", // Permissions, Privileges, and Access Controls
		"CWE-269",  // Improper Privilege Management
		"CWE-285",  // Improper Authorization
		"CWE-732",  // Incorrect Permission Assignment for Critical Resource
		"CWE-862",  // Missing Authorization
		"CWE-863",  // Incorrect Authorization
		"CWE-287",  // Improper Authentication
		"CWE-306",  // Missing Authentication for Critical Function
		"CWE-798",  // Use of Hard-coded Credentials
		"CWE-799":  // Improper Control of Interaction Frequency
		return "VL4"

	// VL5 — носители информации (физический USB / removable / hardware)
	case "CWE-1240", // Use of a Cryptographic Primitive with a Risky Implementation
		"CWE-1390",  // Weak Authentication
		"CWE-1338":  // Improper Protections Against Hardware Overheating
		return "VL5"

	// VL6 — открытые ОС / отсутствие защиты ЛВС / сетевые / утечки
	case "CWE-200", // Information Exposure
		"CWE-22",   // Path Traversal
		"CWE-79",   // Cross-site Scripting
		"CWE-89",   // SQL Injection
		"CWE-352",  // CSRF
		"CWE-601",  // Open Redirect
		"CWE-611",  // XML External Entities
		"CWE-918-NET", // SSRF в сетевых компонентах
		"CWE-693",  // Protection Mechanism Failure
		"CWE-444",  // HTTP Request Smuggling
		"CWE-400":  // Uncontrolled Resource Consumption (DoS)
		return "VL6"
	}

	// CWE-119, CWE-787, CWE-125, CWE-416, CWE-476 и прочие memory-bugs —
	// это классические «уязвимости в версии ПО», т.е. VL2.
	return "VL2"
}

// CWEToVLCodeWithFallback пробегается по списку CWE и возвращает первый
// найденный код VL по таблице выше. Если все CWE дали "VL2" — отдаём
// "VL2". Если CWE-список пустой — тоже "VL2" (consumed by the auto-detect
// pipeline, where a CVE is by definition «версия ПО с уязвимостью»).
func CWEToVLCodeWithFallback(cwes []string) string {
	preferred := ""
	for _, c := range cwes {
		if code := CWEToVLCode(c); code != "" && code != "VL2" {
			return code
		} else if code == "VL2" {
			preferred = "VL2"
		}
	}
	if preferred == "" {
		return "VL2"
	}
	return preferred
}
