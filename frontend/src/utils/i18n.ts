export type RiskLevelKey = "low" | "medium" | "high" | "critical";

const riskLevelLabels: Record<RiskLevelKey, string> = {
    critical: "Критический",
    high: "Высокий",
    medium: "Средний",
    low: "Низкий",
};

const threatTypeLabels: Record<string, string> = {
    external: "Внешняя угроза",
    internal: "Внутренняя угроза",
    third_party: "Сторонняя угроза",
};

const recommendationPriorityLabels: Record<string, string> = {
    high: "Высокий",
    medium: "Средний",
    low: "Низкий",
};

const recommendationCategoryLabels: Record<string, string> = {
    general: "Общие меры",
    hardening: "Усиление защиты",
    access_control: "Управление доступом",
    application_security: "Безопасность приложений",
    data_protection: "Защита данных",
    network_security: "Сетевая безопасность",
    vulnerability_management: "Управление уязвимостями",
    backup: "Резервное копирование",
    regulatory_compliance: "Регуляторное соответствие",
    availability: "Обеспечение доступности",
};

export const normalizeRiskLevel = (level?: string): RiskLevelKey | null => {
    if (!level) return null;
    const normalized = level.toLowerCase() as RiskLevelKey;
    return normalized && normalized in riskLevelLabels ? normalized : null;
};

export const getRiskLevelLabel = (level?: string): string => {
    const normalized = normalizeRiskLevel(level);
    if (!normalized) {
        return level || "—";
    }
    return riskLevelLabels[normalized];
};

export const getThreatTypeLabel = (type?: string): string => {
    if (!type) {
        return "—";
    }
    const normalized = type.toLowerCase();
    return threatTypeLabels[normalized] || type;
};

export const getPriorityLabel = (priority?: string): string => {
    if (!priority) {
        return "—";
    }
    const normalized = priority.toLowerCase();
    return recommendationPriorityLabels[normalized] || priority;
};

export const getRecommendationCategoryLabel = (category?: string): string => {
    if (!category) {
        return "—";
    }
    const normalized = category.toLowerCase();
    return recommendationCategoryLabels[normalized] || category.replace("_", " ");
};
