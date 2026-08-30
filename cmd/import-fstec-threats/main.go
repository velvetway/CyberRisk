// import-fstec-threats — populates the `threats` catalogue from the official
// ФСТЭК threat list (`thrlist.xlsx` from the bdu-fstec-mirror repository).
//
// The importer is idempotent: it UPSERTs each row by `bdu_id`. Old сидерные
// rows that have been wiped by migration 030 won't come back; new runs will
// only add or update rows.
//
// Usage:
//
//	import-fstec-threats \
//	    --source https://github.com/velvetway/bdu-fstec-mirror/raw/main/data/thrlist.xlsx \
//	    --dsn    postgres://app:app@localhost:5432/cyber_risk?sslmode=disable
//
// `--source` may also be a local file path.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"
)

const defaultSource = "https://github.com/velvetway/bdu-fstec-mirror/raw/main/data/thrlist.xlsx"

func main() {
	var (
		source = flag.String("source", defaultSource, "URL or local path to thrlist.xlsx")
		dsn    = flag.String("dsn", os.Getenv("DB_DSN"), "Postgres DSN (or set DB_DSN env)")
	)
	flag.Parse()

	if *dsn == "" {
		log.Fatal("--dsn is required (or set DB_DSN env)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	xlsxBytes, err := loadSource(ctx, *source)
	if err != nil {
		log.Fatalf("load source: %v", err)
	}
	log.Printf("downloaded %d bytes", len(xlsxBytes))

	rows, err := parseThreats(xlsxBytes)
	if err != nil {
		log.Fatalf("parse xlsx: %v", err)
	}
	log.Printf("parsed %d threats", len(rows))

	pool, err := pgxpool.New(ctx, *dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	stats, err := upsertAll(ctx, pool, rows)
	if err != nil {
		log.Fatalf("upsert: %v", err)
	}

	fmt.Printf(`
Import complete:
  threats inserted/updated:   %d
  source links written:       %d
  VL-category links written:  %d
  rows with derived types:    %d
  rows without type matches:  %d
`, stats.threatsUpserted, stats.sourceLinks, stats.vlLinks, stats.withTypes, stats.withoutTypes)
}

func loadSource(ctx context.Context, source string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("status %d", resp.StatusCode)
		}
		return io.ReadAll(resp.Body)
	}
	return os.ReadFile(source)
}

// =====================================================================
// XLSX → in-memory rows
// =====================================================================

// threatRow mirrors one row of the FSTEC list, normalized for DB upsert.
type threatRow struct {
	BDUID       string // "УБИ.001"
	Name        string
	Description string
	SourceText  string
	TargetText  string
	ImpactC     bool
	ImpactI     bool
	ImpactA     bool
	UpdatedAt   time.Time
	Status      string

	// derived
	Source       sourceClassification
	AssetTypes   []string // names from asset_types
	CategoryName string   // name from threat_categories
	VLCodes      []string // VL1..VL6 codes from vl_categories
}

func parseThreats(xlsxBytes []byte) ([]threatRow, error) {
	f, err := excelize.OpenReader(strings.NewReader(string(xlsxBytes)))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheet := f.GetSheetList()[0]
	rawRows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	// Headers occupy rows 1-2 (merged cells), data starts at row 3 (index 2).
	if len(rawRows) < 3 {
		return nil, fmt.Errorf("xlsx too short: %d rows", len(rawRows))
	}

	out := make([]threatRow, 0, len(rawRows)-2)
	for i, r := range rawRows[2:] {
		if len(r) < 11 || strings.TrimSpace(safeCell(r, 0)) == "" {
			continue // empty or partial row
		}
		row := threatRow{
			BDUID:       fmt.Sprintf("УБИ.%03s", strings.TrimSpace(safeCell(r, 0))),
			Name:        strings.TrimSpace(safeCell(r, 1)),
			Description: cleanCell(safeCell(r, 2)),
			SourceText:  strings.TrimSpace(safeCell(r, 3)),
			TargetText:  strings.TrimSpace(safeCell(r, 4)),
			ImpactC:     parseBoolish(safeCell(r, 5)),
			ImpactI:     parseBoolish(safeCell(r, 6)),
			ImpactA:     parseBoolish(safeCell(r, 7)),
			Status:      strings.TrimSpace(safeCell(r, 10)),
		}
		// updated_at = column 9 ("Дата последнего изменения данных")
		if t, ok := parseExcelDate(safeCell(r, 9)); ok {
			row.UpdatedAt = t
		}

		row.Source = classifySource(row.SourceText)
		row.AssetTypes = deriveAssetTypeNames(row.TargetText)
		row.CategoryName = matchCategoryName(row.Name, row.Description)
		row.VLCodes = deriveVLCategoryCodes(row.Name, row.Description, row.TargetText)

		// fix the BDUID padding for IDs > 99 (we want УБИ.100, not УБИ.10 0)
		row.BDUID = normalizeBDUID(strings.TrimSpace(safeCell(r, 0)))

		_ = i
		out = append(out, row)
	}
	return out, nil
}

func normalizeBDUID(idStr string) string {
	if idStr == "" {
		return ""
	}
	// Strip non-digits to be safe.
	digits := ""
	for _, c := range idStr {
		if c >= '0' && c <= '9' {
			digits += string(c)
		}
	}
	if digits == "" {
		return ""
	}
	return fmt.Sprintf("УБИ.%03s", digits)
}

func safeCell(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

func cleanCell(s string) string {
	// XLSX stores soft line breaks as `_x000D_`.
	s = strings.ReplaceAll(s, "_x000D_", "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimSpace(s)
}

// parseBoolish accepts "1"/"0"/"true"/"false"/"да"/"нет".
func parseBoolish(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "да", "yes", "y":
		return true
	default:
		return false
	}
}

// parseExcelDate accepts the typical XLSX date strings produced by GetRows().
func parseExcelDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{
		"01-02-06",         // m-d-yy
		"2006-01-02",       // ISO
		"02.01.2006",       // RU
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
