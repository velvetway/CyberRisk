package asset

import (
	"strings"

	"Diplom/internal/domain"
)

// CalculateCIA автоматически рассчитывает CIA триаду на основе типа актива,
// среды и бизнес-критичности.
func CalculateCIA(assetType, environment string, businessCriticality int16) (confidentiality, integrity, availability int16) {
	// Базовые значения в зависимости от типа актива
	var baseC, baseI, baseA float64

	// Нормализуем тип актива к нижнему регистру для сопоставления
	normalizedType := domain.AssetType(strings.ToLower(assetType))

	switch normalizedType {
	case domain.AssetTypeDatabase:
		// Базы данных: высокая конфиденциальность и целостность
		baseC, baseI, baseA = 4.0, 5.0, 3.0
	case domain.AssetTypeServer:
		// Серверы: высокая доступность и целостность
		baseC, baseI, baseA = 3.0, 4.0, 5.0
	case domain.AssetTypeApplication:
		// Приложения: баланс всех параметров
		baseC, baseI, baseA = 3.0, 3.0, 4.0
	case domain.AssetTypeNetwork:
		// Сеть: критична доступность
		baseC, baseI, baseA = 2.0, 3.0, 5.0
	case domain.AssetTypeWorkstation:
		// Рабочие станции: средние значения
		baseC, baseI, baseA = 2.0, 2.0, 3.0
	case domain.AssetTypeMobile:
		// Мобильные устройства: риск утечки данных
		baseC, baseI, baseA = 4.0, 2.0, 3.0
	case domain.AssetTypeIoT:
		// IoT: низкая конфиденциальность, важна доступность
		baseC, baseI, baseA = 2.0, 2.0, 4.0
	case domain.AssetTypeCloud:
		// Облако: высокие требования ко всем параметрам
		baseC, baseI, baseA = 4.0, 4.0, 4.0
	default:
		// По умолчанию: средние значения
		baseC, baseI, baseA = 3.0, 3.0, 3.0
	}

	// Корректировка по среде
	var envModifier float64
	switch environment {
	case "prod", "production":
		envModifier = 1.0 // Увеличение на 1
	case "stage", "staging":
		envModifier = 0.5 // Небольшое увеличение
	case "test":
		envModifier = -0.5 // Небольшое снижение
	case "dev", "development":
		envModifier = -1.0 // Снижение на 1
	default:
		envModifier = 0.0
	}

	// Корректировка по бизнес-критичности (отклонение от среднего значения 3)
	criticalityModifier := float64(businessCriticality-3) * 0.3

	// Применяем модификаторы
	confidentiality = clamp(int16(baseC + envModifier + criticalityModifier))
	integrity = clamp(int16(baseI + envModifier + criticalityModifier))
	availability = clamp(int16(baseA + envModifier + criticalityModifier))

	return confidentiality, integrity, availability
}

// clamp ограничивает значение диапазоном 1-5
func clamp(value int16) int16 {
	if value < 1 {
		return 1
	}
	if value > 5 {
		return 5
	}
	return value
}

// CalculateCriticality автоматически рассчитывает бизнес-критичность актива
// на основе регуляторных полей: категории данных, уровня защищённости,
// категории КИИ, объёма ПДн, среды и сетевой доступности.
func CalculateCriticality(
	dataCategory *string,
	protectionLevel *string,
	kiiCategory *string,
	hasPersonalData bool,
	personalDataVolume *string,
	hasInternetAccess bool,
	isIsolated bool,
	environment string,
) int16 {
	var score float64 = 2.0 // Базовый уровень

	// 1. Категория данных (наивысший приоритет)
	if dataCategory != nil {
		switch *dataCategory {
		case "state_secret":
			score = 5.0 // Гостайна - максимальная критичность
		case "special_pd", "personal_data_special":
			score = 4.5 // Специальные ПДн (здоровье, биометрия)
		case "biometric_pd", "personal_data_biometric":
			score = 4.5 // Биометрические ПДн
		case "kii":
			score = 5.0 // КИИ - максимальная критичность
		case "common_pd", "personal_data":
			score = 3.5 // Обычные ПДн
		case "public_pd":
			score = 2.5 // Общедоступные ПДн
		case "confidential", "banking_secret", "medical_secret", "commercial_secret":
			score = 3.5 // Конфиденциальная информация
		case "internal":
			score = 2.5 // Внутренняя информация
		case "public":
			score = 1.5 // Публичная информация
		}
	}

	// 2. Категория КИИ (187-ФЗ) - может повысить критичность
	if kiiCategory != nil {
		switch *kiiCategory {
		case "1", "cat1": // Первая категория - максимальная значимость
			score = max(score, 5.0)
		case "2", "cat2": // Вторая категория
			score = max(score, 4.0)
		case "3", "cat3": // Третья категория
			score = max(score, 3.5)
		}
	}

	// 3. Уровень защищённости ПДн (152-ФЗ) - может повысить критичность
	if protectionLevel != nil {
		switch *protectionLevel {
		case "УЗ-1", "uz1": // Максимальный уровень защищённости
			score = max(score, 4.5)
		case "УЗ-2", "uz2":
			score = max(score, 4.0)
		case "УЗ-3", "uz3":
			score = max(score, 3.0)
		case "УЗ-4", "uz4":
			score = max(score, 2.5)
		}
	}

	// 4. Объём ПДн - модификатор
	if hasPersonalData && personalDataVolume != nil {
		switch *personalDataVolume {
		case ">100000":
			score += 0.5
		case "1000-100000":
			score += 0.25
		}
	}

	// 5. Среда - модификатор
	switch environment {
	case "prod", "production":
		score += 0.5
	case "stage", "staging":
		// без изменений
	case "test":
		score -= 0.5
	case "dev", "development":
		score -= 1.0
	}

	// 6. Сетевая доступность - модификаторы
	if hasInternetAccess {
		score += 0.5 // Доступ из интернета повышает риск
	}
	if isIsolated {
		score -= 0.5 // Изоляция снижает риск
	}

	// Ограничиваем результат диапазоном 1-5
	if score < 1.0 {
		score = 1.0
	}
	if score > 5.0 {
		score = 5.0
	}

	return int16(score + 0.5) // Округление до ближайшего целого
}

// max возвращает максимальное из двух float64
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ApplyCIA применяет автоматически рассчитанные значения CIA к активу
func ApplyCIA(asset *domain.Asset) {
	var assetType string
	if asset.Type != nil {
		assetType = *asset.Type
	}
	c, i, a := CalculateCIA(assetType, string(asset.Environment), asset.BusinessCriticality)
	asset.Confidentiality = c
	asset.Integrity = i
	asset.Availability = a
}

// ApplyCriticality автоматически рассчитывает и применяет критичность к активу
func ApplyCriticality(asset *domain.Asset) {
	// Конвертируем типы domain в строки для CalculateCriticality
	var dataCat, protLevel, kiiCat *string
	if asset.DataCategory != nil {
		s := string(*asset.DataCategory)
		dataCat = &s
	}
	if asset.ProtectionLevel != nil {
		s := string(*asset.ProtectionLevel)
		protLevel = &s
	}
	if asset.KIICategory != nil {
		s := string(*asset.KIICategory)
		kiiCat = &s
	}

	asset.BusinessCriticality = CalculateCriticality(
		dataCat,
		protLevel,
		kiiCat,
		asset.HasPersonalData,
		asset.PersonalDataVolume,
		asset.HasInternetAccess,
		asset.IsIsolated,
		string(asset.Environment),
	)
}
