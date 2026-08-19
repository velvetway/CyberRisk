package domain

// SZICertificate — позиция Государственного реестра сертифицированных средств
// защиты информации ФСТЭК России.
type SZICertificate struct {
	ID                int64      `json:"id"`
	CertificateNumber string     `json:"certificate_number"`
	Name              string     `json:"name"`
	Applicant         *string    `json:"applicant,omitempty"`
	Requirements      *string    `json:"requirements,omitempty"`
	ToolType          *string    `json:"tool_type,omitempty"`
	ToolTypeName      *string    `json:"tool_type_name,omitempty"`
	ProtectionClass   *int16     `json:"protection_class,omitempty"`
	NDVLevel          *int16     `json:"ndv_level,omitempty"`
	RegisteredAt      *string    `json:"registered_at,omitempty"`
	ValidUntil        *string    `json:"valid_until,omitempty"`
	SupportUntil      *string    `json:"support_until,omitempty"`
	ValidityKind      string     `json:"validity_kind"`
	IsActive          bool       `json:"is_active"`
	Controls          []string   `json:"controls,omitempty"`
	Prices            []SZIPrice `json:"prices,omitempty"`
}

// SZIPrice — курируемая цена средства защиты.
//
// Реестр ФСТЭК цен не содержит: это перечень сертификатов, а не прайс-лист.
// Цены собираются отдельно из открытых источников, поэтому каждая обязана
// нести ссылку на источник и дату сбора — без них цифру нечем подтвердить.
//
// Значения даются диапазоном: стоимость зависит от объёма закупки и состава
// пакета, и точное число создавало бы ложную точность.
type SZIPrice struct {
	ProductName  string   `json:"product_name"`
	Vendor       *string  `json:"vendor,omitempty"`
	PriceMin     *float64 `json:"price_min,omitempty"`
	PriceMax     *float64 `json:"price_max,omitempty"`
	Currency     string   `json:"currency"`
	LicenseModel string   `json:"license_model"`
	SourceURL    *string  `json:"source_url,omitempty"`
	SourceType   string   `json:"source_type"`
	CollectedAt  string   `json:"collected_at"`
	Note         *string  `json:"note,omitempty"`
}

// SZIControlCoverage — сколько сертифицированных средств доступно для метода
// противодействия ПТСЗИ. Нужно, чтобы видеть, где выбирать не из чего.
type SZIControlCoverage struct {
	ControlCode  string `json:"control_code"`
	Certificates int    `json:"certificates"`
	Vendors      int    `json:"vendors"`
	// WithPrice — сколько средств метода имеют собранную цену. Подбор комплекса
	// возможен только там, где есть из чего выбирать по стоимости.
	WithPrice int `json:"with_price"`
}
