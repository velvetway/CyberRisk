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
	defaultBaseURL = "https://xn--80aajzhsbhw.xn--p1ai" // каталогпо.рф (punycode)
	// Шлюз перед API режет HTTP/2-стримы, если Content-Length превышает
	// ~115 KB: первые пакеты приходят, потом стрим закрывается с
	// INTERNAL_ERROR, JSON оказывается truncated. Лимит=30 даёт ~98 KB
	// и стабильно проходит; 35 ещё ок, 40 уже падает.
	defaultPageSize = 30
	maxPageSize     = 35
	maxRetries      = 5
	retryBaseDelay  = time.Second
	userAgent       = "cyberrisk-import-minreestr/0.2 (+https://github.com/velvetway/CyberRisk)"
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
	if *pageSize > maxPageSize {
		log.Printf("warn: --page-size=%d > %d (gateway truncates HTTP/2 streams ~115KB); clamping to %d",
			*pageSize, maxPageSize, maxPageSize)
		*pageSize = maxPageSize
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
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
	totalReported := 0
	for {
		items, total, err := fetchPageWithRetry(ctx, *baseURL, page, *pageSize)
		if err != nil {
			// Не валим процесс — продолжаем со следующей страницы. Без
			// этого один сетевой сбой стоил бы потери всего хвоста (как
			// в первом запуске, где из 26 094 импортнулось 3 001).
			stats.failedPages++
			log.Printf("page %d: skipped after retries: %v", page, err)
			if (*maxPages > 0 && page >= *maxPages) ||
				(totalReported > 0 && page*(*pageSize) >= totalReported) {
				break
			}
			page++
			continue
		}
		if total > 0 {
			totalReported = total
		}
		if len(items) == 0 {
			break
		}

		log.Printf("page %d: %d items (total reported %d)", page, len(items), total)

		if !*dryRun {
			if err := upsertBatch(ctx, pool, items, categoryByName, &stats); err != nil {
				// upsert-ошибки = проблема в БД, тут уже надо падать
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
  pages failed (skipped):  %d
`, stats.upserted, stats.skipped, stats.withCategory, stats.failedPages)
	if stats.failedPages > 0 {
		os.Exit(2)
	}
}

// =====================================================================
// HTTP fetch
// =====================================================================

type apiSubcategory struct {
	Name string `json:"name"`
	Code string `json:"subcategoryid"`
}

type apiProduct struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Provider      string           `json:"provider"`
	WebsiteURL    string           `json:"websiteUrl"`
	RegistryYear  string           `json:"registryyear"`
	Replaces      string           `json:"replaces"`
	Subcategories []apiSubcategory `json:"subcategories"`
}

type apiResponse struct {
	Items []apiProduct `json:"items"`
	Total int          `json:"total"`
}

// fetchPageWithRetry оборачивает fetchPage экспоненциальным backoff'ом.
// Шлюз каталогпо.рф периодически рвёт HTTP/2-стримы (INTERNAL_ERROR) даже
// при limit=30; повторный запрос обычно проходит. Между попытками — 1s, 2s,
// 4s, 8s, 16s; ctx.Done прерывает цикл досрочно.
func fetchPageWithRetry(ctx context.Context, base string, page, limit int) ([]apiProduct, int, error) {
	var lastErr error
	delay := retryBaseDelay
	for attempt := 1; attempt <= maxRetries; attempt++ {
		items, total, err := fetchPage(ctx, base, page, limit)
		if err == nil {
			return items, total, nil
		}
		lastErr = err
		if attempt == maxRetries {
			break
		}
		log.Printf("page %d attempt %d/%d failed: %v; retrying in %s",
			page, attempt, maxRetries, err, delay)
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}
	return nil, 0, fmt.Errorf("after %d attempts: %w", maxRetries, lastErr)
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
	failedPages  int
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
		categoryID := resolveCategory(p.Subcategories, catByName)
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

// =====================================================================
// Категоризация
// =====================================================================
//
// Минцифры использует иерархический классификатор «XX.YY» (например, 02.07
// = «Средства управления базами данных»). Их таксономия в разы детальнее
// нашей (14 общих категорий: os/dbms/erp/crm/office/antivirus/backup/…),
// поэтому маппим в три прохода:
//   1) точный subcategoryid → конкретная категория (надёжнее всего);
//   2) префикс XX.* → категория (для секций, где почти всё одинаково,
//      например 03.* = ИБ, 04.* = разработка);
//   3) keyword по subcategory.name (страховка от незнакомых кодов);
//   4) fallback в «other», чтобы не возвращать NULL.
//
// При множественных подкатегориях у продукта берём ПЕРВУЮ совпавшую —
// результаты приоритизируются проходами, не порядком в массиве.

// subcatCodeMap — точные коды Минцифры → код нашей категории.
var subcatCodeMap = map[string]string{
	// 01.* — встроенные системные
	"01.02": "os",
	// 02.* — серверная инфра
	"02.04": "virtualization", "02.12": "virtualization",
	"02.05": "backup",
	"02.06": "web",
	"02.07": "dbms",
	"02.08": "monitoring", "02.11": "monitoring",
	// 04.03 — отдельный кейс: «Офисные приложения (№621)» внутри секции разработки
	"04.03": "office",
	// 06.* — офис/коммуникации
	"06.02": "network", "06.13": "network",
	"06.03": "office", "06.05": "office",
	"06.08": "office", "06.09": "office",
	"06.10": "office", "06.11": "office", "06.12": "office",
	"06.04": "mail",
	"06.07": "web",
	// 09.* — корпоративное управление
	"09.07": "erp",
	"09.09": "crm",
	"09.10": "monitoring",
}

// subcatPrefixMap — двузначный префикс → категория. Применяется только
// если точный код не дал ответа.
var subcatPrefixMap = map[string]string{
	"03": "antivirus",   // вся секция ИБ → «Антивирус/СЗИ»
	"04": "development", // вся секция разработки
	"07": "development", // парсеры/анализаторы
}

// nameKeywordMap — корни слов в subcategory.name для случая, когда коды
// продукта не покрыты. Порядок важен: более специфичные раньше.
var nameKeywordMap = []struct{ kw, code string }{
	{"антивирус", "antivirus"}, {"межсетев", "antivirus"}, {"шифрован", "antivirus"},
	{"информационной безопасности", "antivirus"},
	{"операционн", "os"},
	{"субд", "dbms"}, {"базами данных", "dbms"}, {"база данных", "dbms"},
	{"erp", "erp"}, {"crm", "crm"},
	{"почтов", "mail"},
	{"виртуализ", "virtualization"}, {"контейнер", "virtualization"},
	{"резервн", "backup"}, {"бэкап", "backup"},
	{"мониторинг", "monitoring"},
	{"коммуникацион", "network"}, {"сетев", "network"},
	{"разработк", "development"}, {"программирован", "development"},
	{"веб", "web"}, {"браузер", "web"}, {"серверное", "web"},
	{"офис", "office"}, {"редактор", "office"}, {"документообор", "office"},
}

// resolveCategory возвращает id нашей категории для одного продукта или nil.
func resolveCategory(subs []apiSubcategory, catByCode map[string]int16) *int16 {
	pick := func(code string) *int16 {
		if id, ok := catByCode[code]; ok {
			return &id
		}
		return nil
	}

	// 1) точный код подкатегории
	for _, s := range subs {
		if c, ok := subcatCodeMap[s.Code]; ok {
			if id := pick(c); id != nil {
				return id
			}
		}
	}
	// 2) префикс XX.*
	for _, s := range subs {
		if len(s.Code) >= 2 {
			if c, ok := subcatPrefixMap[s.Code[:2]]; ok {
				if id := pick(c); id != nil {
					return id
				}
			}
		}
	}
	// 3) keyword по имени подкатегории
	for _, s := range subs {
		n := strings.ToLower(s.Name)
		for _, kv := range nameKeywordMap {
			if strings.Contains(n, kv.kw) {
				if id := pick(kv.code); id != nil {
					return id
				}
			}
		}
	}
	// 4) fallback на «other»
	return pick("other")
}
