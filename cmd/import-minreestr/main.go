// import-minreestr — populates `software_catalog` from the unofficial
// каталогпо.рф API (≈26 094 products) which mirrors the official
// "Реестр российского ПО" (Минцифры).
//
// The API is paginated: GET /api/products?page=N&limit=100 returns
// {items: [{id, name, provider, description, websiteUrl, registryyear,
//           subcategories[]}, …], total: 26094}.
//
// We map:
//   provider     → software_catalog.vendor
//   name         → software_catalog.name
//   description  → software_catalog.description
//   websiteUrl   → software_catalog.website
//   registryyear → registry_date (parsed YYYY-MM-DD)
//   id           → registry_number (string form)
//   subcategories[].name → category_id (best-effort match against
//                          software_categories; default: "other")
//
// Federal flags (is_russian, fstec_*, fsb_*) are unknown from this source,
// so:
//   is_russian     = TRUE   (every entry in каталогпо.рф is by definition a
//                            registered Russian product)
//   fstec_certified = FALSE (cannot derive from this catalogue)
//   fsb_certified  = FALSE  (same)
//
// FSTEC/FSB-specific flags can be enriched later from a separate source if
// needed; out of scope for P4.
//
// Idempotency: per-row UPSERT on `registry_number`. Re-runs only update.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// httpClient = stdlib defaults. A custom Transport with disabled keep-alive
// or pinned HTTP/1.1 reproducibly hangs against каталогпо.рф; the default
// HTTP/2 + ALPN path works fine.
var httpClient = &http.Client{Timeout: 60 * time.Second}

const (
	defaultBaseURL  = "https://xn--80aajzhsbhw.xn--p1ai" // каталогпо.рф (punycode)
	defaultPageSize = 100
	userAgent       = "cyberrisk-import-minreestr/0.1 (+https://github.com/velvetway/CyberRisk)"
)

func main() {
	var (
		baseURL  = flag.String("base", defaultBaseURL, "Base URL of the каталогпо.рф API")
		dsn      = flag.String("dsn", os.Getenv("DB_DSN"), "Postgres DSN (or set DB_DSN env)")
		pageSize = flag.Int("page-size", defaultPageSize, "Page size (max 100 per API)")
		maxPages = flag.Int("max-pages", 0, "Stop after N pages (0 = no limit)")
		dryRun   = flag.Bool("dry-run", false, "Fetch but don't write to DB")
	)
	flag.Parse()

	if *dsn == "" && !*dryRun {
		log.Fatal("--dsn is required (or set DB_DSN env), or pass --dry-run")
	}
	if *pageSize < 1 || *pageSize > 100 {
		log.Fatal("--page-size must be 1..100")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var pool *pgxpool.Pool
	if !*dryRun {
		var err error
		pool, err = pgxpool.New(ctx, *dsn)
		if err != nil {
			log.Fatalf("connect db: %v", err)
		}
		defer pool.Close()
		if err := pool.Ping(ctx); err != nil {
			log.Fatalf("ping db: %v", err)
		}
	}

	// Cache the software_categories name → id map once.
	categoryByName := map[string]int16{}
	if !*dryRun {
		rows, err := pool.Query(ctx, `SELECT id, name, code FROM software_categories`)
		if err != nil {
			log.Fatalf("load software_categories: %v", err)
		}
		for rows.Next() {
			var id int16
			var name, code string
			if err := rows.Scan(&id, &name, &code); err != nil {
				log.Fatalf("scan category: %v", err)
			}
			categoryByName[strings.ToLower(name)] = id
			categoryByName[strings.ToLower(code)] = id
		}
		rows.Close()
	}

	stats := importStats{}
	page := 1
	for {
		items, total, err := fetchPage(ctx, *baseURL, page, *pageSize)
		if err != nil {
			log.Fatalf("fetch page %d: %v", page, err)
		}
		if len(items) == 0 {
			break
		}

		log.Printf("page %d: %d items (total reported %d)", page, len(items), total)

		if !*dryRun {
			if err := upsertBatch(ctx, pool, items, categoryByName, &stats); err != nil {
				log.Fatalf("upsert page %d: %v", page, err)
			}
		}

		if (*maxPages > 0 && page >= *maxPages) || page*(*pageSize) >= total {
			break
		}
		page++
	}

	fmt.Printf(`
Import complete:
  rows upserted:           %d
  rows skipped (dup):      %d
  rows with category:      %d
`, stats.upserted, stats.skipped, stats.withCategory)
}

// =====================================================================
// HTTP fetch
// =====================================================================

type apiProduct struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Provider      string `json:"provider"`
	WebsiteURL    string `json:"websiteUrl"`
	RegistryYear  string `json:"registryyear"`
	Replaces      string `json:"replaces"`
	Subcategories []struct {
		Name string `json:"name"`
		Code string `json:"subcategoryid"`
	} `json:"subcategories"`
}

type apiResponse struct {
	Items []apiProduct `json:"items"`
	Total int          `json:"total"`
}

func fetchPage(ctx context.Context, base string, page, limit int) ([]apiProduct, int, error) {
	u, err := url.Parse(base + "/api/products")
	if err != nil {
		return nil, 0, err
	}
	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("status %d: %s", resp.StatusCode, string(body[:min(200, len(body))]))
	}
	var ar apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, 0, err
	}
	return ar.Items, ar.Total, nil
}

// =====================================================================
// DB upsert
// =====================================================================

type importStats struct {
	upserted     int
	skipped      int
	withCategory int
}

const upsertSQL = `
INSERT INTO software_catalog (
    name, vendor, version, category_id,
    is_russian, registry_number, registry_date, registry_url,
    fstec_certified, fsb_certified,
    description, website
) VALUES ($1,$2,NULL,$3, TRUE, $4, $5, $6, FALSE, FALSE, $7, $8)
ON CONFLICT (registry_number) DO UPDATE SET
    name         = EXCLUDED.name,
    vendor       = EXCLUDED.vendor,
    category_id  = EXCLUDED.category_id,
    registry_date = EXCLUDED.registry_date,
    registry_url = EXCLUDED.registry_url,
    description  = EXCLUDED.description,
    website      = EXCLUDED.website,
    updated_at   = now()
`

func upsertBatch(ctx context.Context, pool *pgxpool.Pool, items []apiProduct, catByName map[string]int16, stats *importStats) error {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Ensure registry_number is unique so ON CONFLICT works. A plain
	// UNIQUE index is required (Postgres rejects partial indexes for
	// ON CONFLICT inference); we just leave NULL slots out by filling
	// registry_number for every imported row.
	if _, err := tx.Exec(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS uniq_software_registry_number ON software_catalog(registry_number)`); err != nil {
		return fmt.Errorf("ensure unique registry_number index: %w", err)
	}

	for _, p := range items {
		if p.ID == "" || p.Name == "" {
			stats.skipped++
			continue
		}
		var categoryID *int16
		for _, sub := range p.Subcategories {
			if id, ok := catByName[strings.ToLower(sub.Name)]; ok {
				categoryID = &id
				break
			}
		}
		if categoryID != nil {
			stats.withCategory++
		}

		var regDate *time.Time
		if t, ok := parseAPIDate(p.RegistryYear); ok {
			regDate = &t
		}

		_, err := tx.Exec(ctx, upsertSQL,
			p.Name,
			p.Provider,
			categoryID,
			p.ID,
			regDate,
			fmt.Sprintf("https://xn--80aajzhsbhw.xn--p1ai/product/%s", p.ID),
			p.Description,
			p.WebsiteURL,
		)
		if err != nil {
			return fmt.Errorf("upsert id=%s: %w", p.ID, err)
		}
		stats.upserted++
	}
	return tx.Commit(ctx)
}

func parseAPIDate(s string) (time.Time, bool) {
	for _, layout := range []string{
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
