// Подбор комплекса средств защиты: типы ответов оптимизатора.

/** Размер актива. На него умножается цена за единицу лицензирования. */
export interface AssetScale {
    workstations: number;
    servers: number;
    appliances: number;
}

/** Средство, предлагаемое для закрытия метода противодействия. */
export interface OptimizerCandidate {
    control_code: string;
    control_name: string;
    certificate_id: number;
    product_name: string;
    vendor?: string;
    protection_class?: number;
    /** Цена за одну единицу лицензирования. */
    cost_min: number;
    cost_max: number;
    license_model: string;
    /** За что берётся цена: node | server | appliance | bundle. */
    pricing_unit: string;
    /** Сколько единиц нужно при масштабе актива. */
    units: number;
    /** cost_max × units — во столько обойдётся закрытие метода. */
    total_cost: number;
    source_url?: string;
    source_type: string;
    effectiveness: number;
    /** Дата окончания сертификата ФСТЭК; отсутствует у бессрочных. */
    valid_until?: string;
    validity_kind?: string;
}

export interface OptimizerStep {
    candidate: OptimizerCandidate;
    w_before: number;
    w_after: number;
    delta_w: number;
    /** Снижение риска на миллион рублей. */
    efficiency: number;
    cumulative_cost: number;
}

export interface SkippedCandidate {
    candidate: OptimizerCandidate;
    reason: string;
}

export interface OptimizerPlan {
    asset_id: number;
    budget: number;
    scale: AssetScale;
    baseline_w: number;
    resulting_w: number;
    total_delta: number;
    total_cost: number;
    steps: OptimizerStep[];
    skipped?: SkippedCandidate[];
    method: string;
    /** Сверялся ли результат с полным перебором. */
    exhaustive_checked: boolean;
    /** Совпал ли жадный план с точным оптимумом. */
    greedy_is_optimal: boolean;
    exhaustive_delta?: number;
    warnings?: string[];
}

export interface RoadmapPurchase {
    candidate: OptimizerCandidate;
    /** Месяц горизонта, с которого средство начинает снижать риск. */
    active_from_month: number;
    deploy_months: number;
    cost: number;
    /** Месяц, когда истекает сертификат; отсутствует у бессрочных. */
    expires_at_month?: number;
}

export interface RoadmapPeriod {
    year: number;
    purchases: RoadmapPurchase[];
    spent: number;
    w_at_start: number;
    w_at_end: number;
    /** Вклад года в площадь под кривой риска, в единицах «W · год». */
    risk_area: number;
}

export interface Roadmap {
    asset_id: number;
    horizon_years: number;
    budget_per_year: number;
    scale: AssetScale;
    periods: RoadmapPeriod[];
    baseline_w: number;
    final_w: number;
    total_cost: number;
    /** Площадь, если не делать ничего. */
    baseline_area: number;
    risk_area: number;
    area_reduction: number;
    skipped?: SkippedCandidate[];
    warnings?: string[];
}

/** Устойчивость отдельного метода к неточности коэффициентов. */
export interface ControlStability {
    control_code: string;
    product_name?: string;
    /** Доля прогонов, где метод вошёл в план, от 0 до 1. */
    frequency: number;
    runs: number;
}

/** Отчёт о зависимости плана от точности экспертных коэффициентов. */
export interface SensitivityReport {
    asset_id: number;
    budget: number;
    runs: number;
    /** Коридор возмущения долей: 0.2 означает ±20%. */
    variation: number;
    base_delta: number;
    mean_delta: number;
    min_delta: number;
    max_delta: number;
    std_dev: number;
    /** Доля прогонов, где состав плана совпал с исходным. */
    composition_stability: number;
    controls: ControlStability[];
    verdict: string;
}
