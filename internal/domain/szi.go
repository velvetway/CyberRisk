package domain

// SZICertificate — позиция Государственного реестра сертифицированных средств
// защиты информации ФСТЭК России.
type SZICertificate struct {
	ID                int64    `json:"id"`
	CertificateNumber string   `json:"certificate_number"`
	Name              string   `json:"name"`
	Applicant         *string  `json:"applicant,omitempty"`
	Requirements      *string  `json:"requirements,omitempty"`
	ToolType          *string  `json:"tool_type,omitempty"`
	ToolTypeName      *string  `json:"tool_type_name,omitempty"`
	ProtectionClass   *int16   `json:"protection_class,omitempty"`
	NDVLevel          *int16   `json:"ndv_level,omitempty"`
	RegisteredAt      *string  `json:"registered_at,omitempty"`
	ValidUntil        *string  `json:"valid_until,omitempty"`
	SupportUntil      *string  `json:"support_until,omitempty"`
	ValidityKind      string   `json:"validity_kind"`
	IsActive          bool     `json:"is_active"`
	Controls          []string `json:"controls,omitempty"`
}

// SZIControlCoverage — сколько сертифицированных средств доступно для метода
// противодействия ПТСЗИ. Нужно, чтобы видеть, где выбирать не из чего.
type SZIControlCoverage struct {
	ControlCode  string `json:"control_code"`
	Certificates int    `json:"certificates"`
	Vendors      int    `json:"vendors"`
}
