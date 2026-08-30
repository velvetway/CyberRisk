package optimizer

// AssetScale — размер защищаемого актива.
//
// Без него стоимость комплекса невозможно посчитать честно: цены собраны за
// разные единицы, и лицензия на рабочую станцию несопоставима с ценой шасси
// межсетевого экрана. Масштаб переводит их в общие деньги.
type AssetScale struct {
	// Workstations — рабочие станции и пользователи: к ним привязано
	// большинство программных лицензий.
	Workstations int `json:"workstations"`
	// Servers — серверы, процессоры, ядра. Серверные лицензии считаются
	// по ним, хотя единица у разных вендоров разная: vGate берёт за
	// физический процессор, Postgres Pro — за ядро.
	Servers int `json:"servers"`
	// Appliances — точки, где ставится аппаратный комплекс: шлюзы,
	// узлы периметра.
	Appliances int `json:"appliances"`
}

// DefaultScale — масштаб по умолчанию: один узел каждого вида.
//
// При нём стоимость совпадает со списочной ценой за единицу, то есть
// поведение остаётся прежним, если масштаб не задан.
func DefaultScale() AssetScale {
	return AssetScale{Workstations: 1, Servers: 1, Appliances: 1}
}

func (s AssetScale) normalized() AssetScale {
	if s.Workstations < 1 {
		s.Workstations = 1
	}
	if s.Servers < 1 {
		s.Servers = 1
	}
	if s.Appliances < 1 {
		s.Appliances = 1
	}
	return s
}

// pricingUnit определяет, за что берётся цена.
//
// Поле license_model в собранных данных смешивает две разные вещи: единицу
// измерения (`per_node`, `per_server`, `appliance`, `bundle`) и срок действия
// (`perpetual`, `yearly`). Разделить их в источниках нечем — прайсы пишут
// «бессрочная лицензия на рабочее место» одной строкой, — поэтому срочные
// модели трактуются по тому, как они устроены в собранном наборе: и у
// Dallas Lock (`perpetual`), и у Secret Net Studio (`yearly`) цена стоит за
// рабочее место.
func pricingUnit(licenseModel string) string {
	switch licenseModel {
	case "per_server":
		return "server"
	case "appliance":
		return "appliance"
	case "bundle":
		return "bundle"
	case "per_node", "yearly", "perpetual":
		return "node"
	default:
		// Неизвестную модель считаем ценой за комплект: умножать на масштаб
		// то, про что ничего не известно, опаснее, чем не умножать.
		return "bundle"
	}
}

// unitCount — сколько единиц придётся купить при заданном масштабе.
func unitCount(licenseModel string, scale AssetScale) int {
	switch pricingUnit(licenseModel) {
	case "node":
		return scale.Workstations
	case "server":
		return scale.Servers
	case "appliance":
		return scale.Appliances
	default:
		return 1
	}
}

// scaledCost — стоимость закрытия метода при заданном масштабе.
func scaledCost(c Candidate, scale AssetScale) float64 {
	return c.CostMax * float64(unitCount(c.LicenseModel, scale))
}
