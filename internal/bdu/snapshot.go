// Package bdu wraps the read-only БДУ ФСТЭК SQLite snapshot, downloaded by
// `cmd/import-bdu-snapshot`. The snapshot lives outside Postgres because it
// is large (~470 MB) and entirely static between syncs — we never write to it.
//
// All methods on Snapshot are safe for concurrent use; the underlying *sql.DB
// pool handles its own connection lifecycle.
package bdu

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// Vulnerability is the canonical projection of one row from `vulnerabilities`
// in the snapshot.
type Vulnerability struct {
	ID            string  // "BDU:2014-00001"
	Name          string
	Description   string
	SoftwareNames string  // space-joined products from БДУ
	Vendors       string  // space-joined vendors
	CVEs          string  // space-joined CVE ids
	Severity      string
	SeverityLevel int     // 1..4
	CVSSScore     float64
	IdentifyYear  int
	HasExploit    bool
	HasFix        bool
	CWEs          []string // joined from `cwes` table
}

// Snapshot is a thin facade over the БДУ SQLite file.
type Snapshot struct {
	db *sql.DB
}

// Open opens the snapshot in read-only mode. The path should point at the
// unpacked SQLite file (NOT the .gz).
func Open(path string) (*Snapshot, error) {
	if path == "" {
		return nil, errors.New("bdu snapshot: empty path")
	}
	// `immutable=1` is the right flag for a static snapshot: SQLite skips
	// journal files and assumes the file never changes — fastest path for
	// pure-read workloads.
	dsn := fmt.Sprintf("file:%s?mode=ro&immutable=1", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open bdu snapshot %q: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping bdu snapshot %q: %w", path, err)
	}
	// Tiny pool — most queries are single-shot; keep memory tight.
	db.SetMaxOpenConns(4)
	return &Snapshot{db: db}, nil
}

// Close releases the underlying pool.
func (s *Snapshot) Close() error { return s.db.Close() }

// Stats reports how many rows are in the snapshot. Useful as a health-check.
func (s *Snapshot) Stats(ctx context.Context) (vulnCount, softwareCount, cweCount int, err error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM vulnerabilities),
			(SELECT COUNT(*) FROM software),
			(SELECT COUNT(*) FROM cwes)`)
	err = row.Scan(&vulnCount, &softwareCount, &cweCount)
	return
}

// Get returns one vulnerability by its БДУ id ("BDU:2014-00001").
func (s *Snapshot) Get(ctx context.Context, bduID string) (*Vulnerability, error) {
	const q = `
SELECT id, name, description, software_names, vendors, cves_joined,
       severity, severity_level, cvss_score, identify_year, has_exploit, has_fix
FROM vulnerabilities WHERE id = ?`
	v := &Vulnerability{}
	row := s.db.QueryRowContext(ctx, q, bduID)
	var hasExploit, hasFix int
	err := row.Scan(&v.ID, &v.Name, &v.Description, &v.SoftwareNames, &v.Vendors, &v.CVEs,
		&v.Severity, &v.SeverityLevel, &v.CVSSScore, &v.IdentifyYear, &hasExploit, &hasFix)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	v.HasExploit = hasExploit != 0
	v.HasFix = hasFix != 0

	if cwes, err := s.cwesForBDU(ctx, bduID); err == nil {
		v.CWEs = cwes
	}
	return v, nil
}

// SoftwareLookup returns every БДУ vulnerability whose `software` row matches
// the given vendor + name + version.
//
// Used by the asset-vulnerability auto-detection in P6: when the operator
// adds, say, "Astra Linux" 1.7 to an asset, we materialize every relevant
// БДУ record into asset_vulnerabilities.
//
// Версия (assetVersion) — точная строка с актива; пустое значение значит
// «оператор не указал», тогда возвращаются все версии этого ПО. Фильтр
// проводится через VersionMatches — поддерживает exact, "-" (catch-all)
// и русские range-формы «от X до Y включительно».
//
// IMPORTANT: SQLite's stock `LOWER()` is ASCII-only — для кириллицы это
// no-op, поэтому vendor matching выполняется case-sensitive (вход — обычно
// дословная копия из того же реестра, что и `software.vendor`).
//
// `limit` каpит выходной результат (default 200). Внутри мы загружаем
// до 5×limit строк из БДУ, потому что после version-фильтра и дедупа по
// bdu_id остатки могут сильно сократиться.
func (s *Snapshot) SoftwareLookup(ctx context.Context, vendor, name, assetVersion string, limit int) ([]Vulnerability, error) {
	if limit <= 0 {
		limit = 200
	}
	vendorLike := "%" + strings.TrimSpace(vendor) + "%"
	nameLike := "%" + strings.ToLower(strings.TrimSpace(name)) + "%"
	assetVer := strings.TrimSpace(assetVersion)

	// Версионная стратегия:
	// • Если оператор не указал версию — берём всё (старое поведение, дедуп по bdu_id).
	// • Если указал — фильтруем в SQL по кандидатам: точное совпадение, catch-all
	//   ('-'/''), либо любые range-формы ('от %', 'до %'). Дальше в Go доводим
	//   фильтрацию до точного match через VersionMatches (для range-форм).
	//
	// Это даёт O(N_match) выборку даже для редких версий типа "4.2.6" вместо
	// зачистки топ-1000 по CVSS, в которые редкая версия может не попасть.
	var (
		q    string
		args []any
	)
	if assetVer == "" {
		q = `
SELECT v.id, v.name, v.description, v.software_names, v.vendors, v.cves_joined,
       v.severity, v.severity_level, v.cvss_score, v.identify_year, v.has_exploit, v.has_fix,
       s.version
FROM vulnerabilities v
JOIN software s ON s.bdu_id = v.id
WHERE s.vendor LIKE ? AND LOWER(s.name) LIKE ?
ORDER BY v.cvss_score DESC NULLS LAST
LIMIT ?`
		args = []any{vendorLike, nameLike, limit * 5}
	} else {
		q = `
SELECT v.id, v.name, v.description, v.software_names, v.vendors, v.cves_joined,
       v.severity, v.severity_level, v.cvss_score, v.identify_year, v.has_exploit, v.has_fix,
       s.version
FROM vulnerabilities v
JOIN software s ON s.bdu_id = v.id
WHERE s.vendor LIKE ? AND LOWER(s.name) LIKE ?
  AND (s.version = ? OR s.version = '-' OR s.version = ''
       OR s.version LIKE 'от %' OR s.version LIKE 'до %')
ORDER BY v.cvss_score DESC NULLS LAST
LIMIT ?`
		args = []any{vendorLike, nameLike, assetVer, limit * 5}
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Vulnerability, 0, limit)
	seen := make(map[string]bool)
	for rows.Next() {
		v := Vulnerability{}
		var hasExploit, hasFix int
		var bduVersion string
		if err := rows.Scan(&v.ID, &v.Name, &v.Description, &v.SoftwareNames, &v.Vendors, &v.CVEs,
			&v.Severity, &v.SeverityLevel, &v.CVSSScore, &v.IdentifyYear, &hasExploit, &hasFix,
			&bduVersion); err != nil {
			return nil, err
		}
		// Дедуп: одна и та же уязвимость может быть для нескольких версий
		// одного и того же ПО — если хотя бы одна version row подошла под
		// фильтр, уязвимость попадает в результат один раз.
		if seen[v.ID] {
			continue
		}
		if !VersionMatches(bduVersion, assetVer) {
			continue
		}
		seen[v.ID] = true
		v.HasExploit = hasExploit != 0
		v.HasFix = hasFix != 0
		out = append(out, v)
		if len(out) >= limit {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Bulk-fetch CWEs in one query for the matched IDs.
	if err := s.attachCWEs(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

// cwesForBDU returns CWE codes for one vulnerability.
func (s *Snapshot) cwesForBDU(ctx context.Context, bduID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT cwe_id FROM cwes WHERE bdu_id = ? ORDER BY cwe_id`, bduID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// attachCWEs adds CWE arrays to every Vulnerability in `vs`, in a single round-trip.
func (s *Snapshot) attachCWEs(ctx context.Context, vs []Vulnerability) error {
	if len(vs) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(vs))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(vs))
	idx := map[string]int{}
	for i, v := range vs {
		args[i] = v.ID
		idx[v.ID] = i
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT bdu_id, cwe_id FROM cwes WHERE bdu_id IN (`+placeholders+`) ORDER BY bdu_id, cwe_id`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var bduID, cwe string
		if err := rows.Scan(&bduID, &cwe); err != nil {
			return err
		}
		if i, ok := idx[bduID]; ok {
			vs[i].CWEs = append(vs[i].CWEs, cwe)
		}
	}
	return rows.Err()
}
