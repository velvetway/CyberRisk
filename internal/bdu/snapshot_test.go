package bdu

import (
	"context"
	"os"
	"testing"
)

// TestSnapshot_Smoke runs against the real БДУ snapshot if BDU_SNAPSHOT_PATH
// is set; otherwise it's skipped. We don't ship the snapshot in the repo,
// so CI will skip; local development with the snapshot present (via
// `cmd/import-bdu-snapshot`) will run it.
func TestSnapshot_Smoke(t *testing.T) {
	path := os.Getenv("BDU_SNAPSHOT_PATH")
	if path == "" {
		t.Skip("BDU_SNAPSHOT_PATH not set; skipping snapshot smoke test")
	}

	snap, err := Open(path)
	if err != nil {
		t.Fatalf("open snapshot: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	vc, sc, cc, err := snap.Stats(ctx)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	t.Logf("snapshot: %d vulnerabilities, %d software rows, %d CWE links", vc, sc, cc)
	if vc < 50000 {
		t.Errorf("vulnerability count looks suspicious: %d", vc)
	}

	v, err := snap.Get(ctx, "BDU:2014-00001")
	if err != nil {
		t.Fatalf("get BDU:2014-00001: %v", err)
	}
	if v == nil {
		t.Fatal("BDU:2014-00001 not found in snapshot")
	}
	if v.Name == "" {
		t.Errorf("empty name for BDU:2014-00001")
	}
	t.Logf("BDU:2014-00001: %s (CVSS %.1f, CWEs %v)", v.Name, v.CVSSScore, v.CWEs)

	// Vendor lookup smoke — D-Link is well-represented in the snapshot.
	matches, err := snap.SoftwareLookup(ctx, "D-Link", "DSR-500", "", 5)
	if err != nil {
		t.Fatalf("software lookup: %v", err)
	}
	if len(matches) == 0 {
		t.Errorf("expected at least one D-Link DSR-500 vulnerability")
	}
	t.Logf("D-Link DSR-500 lookup: %d vulnerabilities (showing top %d)", len(matches), min(3, len(matches)))
	for i, m := range matches {
		if i >= 3 {
			break
		}
		t.Logf("  - %s [CVSS %.1f] %s", m.ID, m.CVSSScore, m.Name)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
