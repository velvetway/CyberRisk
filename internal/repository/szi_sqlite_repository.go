package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"Diplom/internal/domain"

	_ "modernc.org/sqlite"
)

// SZISearchFilter — отбор по каталогу сертифицированных СЗИ.
type SZISearchFilter struct {
	Query string
	// ControlCode — метод противодействия ПТСЗИ (A, FW, IDS, ...).
	ControlCode string
	// MaxProtectionClass — класс защиты не хуже указанного. Меньше значение —
	// строже требования, поэтому фильтр идёт по `<=`.
	MaxProtectionClass *int16
	// ActiveOnly — только сертификаты, действующие на дату сборки снапшота.
	ActiveOnly bool
	Limit      int
}

type SZIRepository interface {
	IsAvailable() bool
	Search(ctx context.Context, f SZISearchFilter) ([]domain.SZICertificate, error)
	GetByID(ctx context.Context, id int64) (*domain.SZICertificate, error)
	ControlCoverage(ctx context.Context) ([]domain.SZIControlCoverage, error)
}

type sziSQLiteRepository struct {
	db *sql.DB
}

func NewSZISQLiteRepository(path string) (SZIRepository, error) {
	if path == "" {
		return nil, nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("open szi sqlite: %w", err)
	}
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro&cache=shared")
	if err != nil {
		return nil, fmt.Errorf("open szi sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping szi sqlite: %w", err)
	}
	return &sziSQLiteRepository{db: db}, nil
}

func (r *sziSQLiteRepository) IsAvailable() bool {
	return r != nil && r.db != nil
}

const sziColumns = `
    c.rowid, c.certificate_number, c.name, c.applicant, c.requirements,
    c.tool_type, c.tool_type_name, c.protection_class, c.ndv_level,
    c.registered_at, c.valid_until, c.support_until, c.validity_kind, c.is_active`

func (r *sziSQLiteRepository) Search(ctx context.Context, f SZISearchFilter) ([]domain.SZICertificate, error) {
	if !r.IsAvailable() {
		return nil, fmt.Errorf("szi sqlite is not configured")
	}

	limit := f.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	filters := []string{"1=1"}
	args := []any{}

	if q := strings.TrimSpace(f.Query); q != "" {
		filters = append(filters, "(c.name LIKE ? OR c.applicant LIKE ?)")
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	if f.ActiveOnly {
		filters = append(filters, "c.is_active = 1")
	}
	if f.MaxProtectionClass != nil {
		filters = append(filters, "c.protection_class IS NOT NULL AND c.protection_class <= ?")
		args = append(args, *f.MaxProtectionClass)
	}

	from := "certificates c"
	if code := strings.TrimSpace(f.ControlCode); code != "" {
		from = "certificates c JOIN certificate_controls cc ON cc.certificate_rowid = c.rowid"
		filters = append(filters, "cc.control_code = ?")
		args = append(args, code)
	}

	// Сортировка ставит вперёд средства со строгим классом защиты: у них выше
	// требования, а NULL-класс уводим в конец, чтобы нераспознанное не мешало.
	query := fmt.Sprintf(`
SELECT DISTINCT %s
FROM %s
WHERE %s
ORDER BY (c.protection_class IS NULL), c.protection_class, c.name
LIMIT ?`, sziColumns, from, strings.Join(filters, " AND "))
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search szi certificates: %w", err)
	}
	defer rows.Close()

	items, err := scanSZICertificates(rows)
	if err != nil {
		return nil, err
	}
	items, err = r.attachControls(ctx, items)
	if err != nil {
		return nil, err
	}
	return r.attachPrices(ctx, items)
}

func (r *sziSQLiteRepository) GetByID(ctx context.Context, id int64) (*domain.SZICertificate, error) {
	if !r.IsAvailable() {
		return nil, fmt.Errorf("szi sqlite is not configured")
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT %s FROM certificates c WHERE c.rowid = ?`, sziColumns), id)
	if err != nil {
		return nil, fmt.Errorf("get szi certificate: %w", err)
	}
	defer rows.Close()

	items, err := scanSZICertificates(rows)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	withControls, err := r.attachControls(ctx, items)
	if err != nil {
		return nil, err
	}
	withPrices, err := r.attachPrices(ctx, withControls)
	if err != nil {
		return nil, err
	}
	return &withPrices[0], nil
}

// ControlCoverage показывает, сколько действующих сертификатов приходится на
// каждый метод ПТСЗИ. Методы без единого средства в ответе отсутствуют —
// это и есть места, где выбирать не из чего.
func (r *sziSQLiteRepository) ControlCoverage(ctx context.Context) ([]domain.SZIControlCoverage, error) {
	if !r.IsAvailable() {
		return nil, fmt.Errorf("szi sqlite is not configured")
	}
	// Цены живут в отдельных таблицах, которых может не быть: они создаются
	// отдельным шагом сборки зеркала.
	priceExpr := "0"
	priceJoin := ""
	if r.hasPrices(ctx) {
		priceExpr = "COUNT(DISTINCT CASE WHEN p.price_min IS NOT NULL THEN c.rowid END)"
		priceJoin = `
LEFT JOIN certificate_prices cp ON cp.certificate_rowid = c.rowid
LEFT JOIN product_prices p ON p.id = cp.price_id`
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT cc.control_code,
       COUNT(DISTINCT c.rowid),
       COUNT(DISTINCT c.applicant),
       %s
FROM certificates c
JOIN certificate_controls cc ON cc.certificate_rowid = c.rowid%s
WHERE c.is_active = 1
GROUP BY cc.control_code
ORDER BY 2 DESC`, priceExpr, priceJoin))
	if err != nil {
		return nil, fmt.Errorf("szi control coverage: %w", err)
	}
	defer rows.Close()

	out := make([]domain.SZIControlCoverage, 0, 11)
	for rows.Next() {
		var c domain.SZIControlCoverage
		if err := rows.Scan(&c.ControlCode, &c.Certificates, &c.Vendors, &c.WithPrice); err != nil {
			return nil, fmt.Errorf("scan szi coverage: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// attachControls дозаполняет методы противодействия одним запросом на выборку,
// чтобы не бегать в базу за каждым сертификатом.
func (r *sziSQLiteRepository) attachControls(ctx context.Context, items []domain.SZICertificate) ([]domain.SZICertificate, error) {
	if len(items) == 0 {
		return items, nil
	}

	placeholders := make([]string, len(items))
	args := make([]any, len(items))
	index := make(map[int64]int, len(items))
	for i := range items {
		placeholders[i] = "?"
		args[i] = items[i].ID
		index[items[i].ID] = i
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT certificate_rowid, control_code
FROM certificate_controls
WHERE certificate_rowid IN (%s)
ORDER BY control_code`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, fmt.Errorf("load szi controls: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, fmt.Errorf("scan szi control: %w", err)
		}
		if i, ok := index[id]; ok {
			items[i].Controls = append(items[i].Controls, code)
		}
	}
	return items, rows.Err()
}

// hasPrices сообщает, подмешаны ли в базу курируемые цены. Таблица создаётся
// отдельным шагом сборки зеркала, и её может не быть — тогда каталог работает
// без цен, а не падает.
func (r *sziSQLiteRepository) hasPrices(ctx context.Context) bool {
	var name string
	err := r.db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type='table' AND name='product_prices'").Scan(&name)
	return err == nil
}

// attachPrices дозаполняет цены одним запросом на всю выборку.
func (r *sziSQLiteRepository) attachPrices(ctx context.Context, items []domain.SZICertificate) ([]domain.SZICertificate, error) {
	if len(items) == 0 || !r.hasPrices(ctx) {
		return items, nil
	}

	placeholders := make([]string, len(items))
	args := make([]any, len(items))
	index := make(map[int64]int, len(items))
	for i := range items {
		placeholders[i] = "?"
		args[i] = items[i].ID
		index[items[i].ID] = i
	}

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
SELECT cp.certificate_rowid, p.product_name, p.vendor, p.price_min, p.price_max,
       p.currency, p.license_model, p.source_url, p.source_type, p.collected_at, p.note
FROM certificate_prices cp
JOIN product_prices p ON p.id = cp.price_id
WHERE cp.certificate_rowid IN (%s)
ORDER BY p.product_name`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, fmt.Errorf("load szi prices: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var certID int64
		var p domain.SZIPrice
		var min, max sql.NullFloat64

		if err := rows.Scan(
			&certID, &p.ProductName, &p.Vendor, &min, &max,
			&p.Currency, &p.LicenseModel, &p.SourceURL, &p.SourceType, &p.CollectedAt, &p.Note,
		); err != nil {
			return nil, fmt.Errorf("scan szi price: %w", err)
		}
		if min.Valid {
			v := min.Float64
			p.PriceMin = &v
		}
		if max.Valid {
			v := max.Float64
			p.PriceMax = &v
		}
		if i, ok := index[certID]; ok {
			items[i].Prices = append(items[i].Prices, p)
		}
	}
	return items, rows.Err()
}

func scanSZICertificates(rows *sql.Rows) ([]domain.SZICertificate, error) {
	out := make([]domain.SZICertificate, 0)
	for rows.Next() {
		var c domain.SZICertificate
		var protectionClass, ndvLevel sql.NullInt64
		var active int

		if err := rows.Scan(
			&c.ID, &c.CertificateNumber, &c.Name, &c.Applicant, &c.Requirements,
			&c.ToolType, &c.ToolTypeName, &protectionClass, &ndvLevel,
			&c.RegisteredAt, &c.ValidUntil, &c.SupportUntil, &c.ValidityKind, &active,
		); err != nil {
			return nil, fmt.Errorf("scan szi certificate: %w", err)
		}

		if protectionClass.Valid {
			v := int16(protectionClass.Int64)
			c.ProtectionClass = &v
		}
		if ndvLevel.Valid {
			v := int16(ndvLevel.Int64)
			c.NDVLevel = &v
		}
		c.IsActive = active == 1
		out = append(out, c)
	}
	return out, rows.Err()
}
