package risk

import "Diplom/internal/domain"

// IsApplicable отвечает на вопрос «имеет ли смысл считать W для пары
// (актив, угроза)?». Семантически это первый этаж двух-слойной модели
// applicability + presence: applicability — статика по типу актива,
// presence — динамика по asset_vulnerabilities (вычисляется отдельно).
//
// Правила P7:
//   1. Если threat.AppliesToAssetTypes пуст — угроза универсальна.
//      Применима ко всем активам (включая активы без типа).
//   2. Если массив непустой и у актива нет типа (asset_type_id = NULL)
//      — мы не можем подтвердить применимость, но и опровергнуть тоже:
//      в этом случае оставляем пару «применимой» (соблюдаем
//      консервативный принцип «лучше показать лишнее, чем спрятать
//      реальный риск»).
//   3. Если у актива есть тип и он не входит в AppliesToAssetTypes —
//      пара не применима.
func IsApplicable(asset domain.Asset, threat domain.Threat) bool {
	if len(threat.AppliesToAssetTypes) == 0 {
		return true
	}
	if asset.AssetTypeID == nil {
		return true
	}
	target := *asset.AssetTypeID
	for _, t := range threat.AppliesToAssetTypes {
		if t == target {
			return true
		}
	}
	return false
}
