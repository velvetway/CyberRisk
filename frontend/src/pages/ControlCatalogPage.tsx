// ControlCatalogPage — «Каталог средств защиты информации (БД ПАСЗИ)».
// Закрывает блок «БД ПАСЗИ» (программно-аппаратные средства защиты)
// из 7.png диплома.
//
// Показывает все 11 контролей из ПТСЗИ (A/FW/HP/DZ/IDS/AD/R/L/TE/DS/DD)
// с описанием, какие VL-категории каждый закрывает, на скольких активах
// внедрён, и из каких compliance-требований ФСТЭК-17 / ISO он покрывает.
import React, { useEffect, useMemo, useState } from "react";
import { api, authFetch, Control } from "../api/client";
import { Card, Chip, Icon } from "../components/design";

// VL-категории — из диплома, фиксированные 6 шт.
const VL_NAMES: Record<number, string> = {
  1: "VL₁ Нештатное доп. ПО",
  2: "VL₂ Устаревшие версии ПО",
  3: "VL₃ Недекларируемое ПО",
  4: "VL₄ Обход правил администратором",
  5: "VL₅ Носители (USB/диски)",
  6: "VL₆ Открытые ОС / нет защиты ЛВС",
};

// Матрица VL → контроли (из диплома 4.png; коды контролей).
const CONTROL_TO_VLS: Record<string, number[]> = {
  "Антивирус": [1, 2, 3, 5],
  "Программная защита от НСД": [1, 2, 3, 4, 5],
  "Системы администрирования": [1, 2, 3, 4, 5],
  "Межсетевой экран": [6],
  "Демилитаризованная зона": [6],
  "Honeypot": [6],
  "Система обнаружения вторжений": [6],
  "Шифрование трафика": [6],
  "DDoS-фильтры": [6],
  "Резервное копирование": [],   // в дипломной модели входит, но не закрывает VL напрямую
  "Цифровая подпись": [],
};

// Соответствие control → стандарты (по нашей миграции 038).
// Для каждого контроля — список (стандарт.код, requirement.код) что он закрывает.
const CONTROL_STDS: Record<string, { fstec: string[]; iso: string[] }> = {
  "Антивирус":                     { fstec: ["АВЗ.1", "АВЗ.2", "ОЦЛ.1", "ОЦЛ.4"], iso: ["A.8.7", "A.8.8"] },
  "Программная защита от НСД":     { fstec: ["ИАФ.1", "ИАФ.5", "УПД.2", "УПД.6", "УПД.13", "ОЦЛ.1", "РСБ.3"], iso: ["A.5.10", "A.5.15", "A.8.9"] },
  "Системы администрирования":     { fstec: ["ИАФ.1", "ИАФ.3", "ИАФ.5", "УПД.1", "УПД.2", "УПД.6", "УПД.13", "РСБ.1", "РСБ.2", "РСБ.3"], iso: ["A.5.15", "A.5.16", "A.5.17", "A.8.5", "A.8.8", "A.8.9"] },
  "Межсетевой экран":              { fstec: ["ЗИС.1", "ЗИС.5", "ЗИС.18"], iso: ["A.8.20", "A.8.22", "A.8.23", "A.8.26"] },
  "Демилитаризованная зона":       { fstec: ["ЗИС.1", "ЗИС.5"], iso: ["A.8.22"] },
  "Honeypot":                      { fstec: ["СОВ.1"], iso: ["A.7.4", "A.8.16"] },
  "Система обнаружения вторжений": { fstec: ["СОВ.1", "СОВ.2", "ЗИС.5", "РСБ.1", "РСБ.2", "ОЦЛ.4"], iso: ["A.7.4", "A.8.16", "A.8.20", "A.8.23", "A.8.26"] },
  "Шифрование трафика":            { fstec: ["УПД.13", "ЗИС.10", "ЗИС.20", "ЗИС.27"], iso: ["A.5.17", "A.5.23", "A.8.5", "A.8.24"] },
  "Резервное копирование":         { fstec: ["ОДТ.1"], iso: ["A.8.13"] },
  "Цифровая подпись":              { fstec: ["ОЦЛ.1", "ЗИС.10"], iso: ["A.8.24"] },
  "DDoS-фильтры":                  { fstec: ["ОДТ.4", "ЗИС.18"], iso: ["A.8.26"] },
};

// Группировка контролей по «мероприятиям» из 8.png диплома.
const MEASURE_GROUP: Record<string, string> = {
  "Антивирус":                     "АРМ",
  "Программная защита от НСД":     "АРМ / Конф.инф.",
  "Системы администрирования":     "АРМ / Конф.инф.",
  "Межсетевой экран":              "ЛВС",
  "Демилитаризованная зона":       "ЛВС",
  "Honeypot":                      "ЛВС",
  "Система обнаружения вторжений": "ЛВС",
  "DDoS-фильтры":                  "ЛВС",
  "Шифрование трафика":            "Документооборот",
  "Цифровая подпись":              "Документооборот",
  "Резервное копирование":         "Конф.инф.",
};

export const ControlCatalogPage: React.FC = () => {
  const [controls, setControls] = useState<Control[]>([]);
  const [usage, setUsage] = useState<Map<number, number>>(new Map());
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      api.getControls(),
      authFetch("/api/organization/asset-matrix").then(r => r.ok ? r.json() : []),
    ])
      .then(([cs, matrix]) => {
        setControls(cs);
        // Подсчёт «на скольких активах внедрён» через подсчёт control_count в asset-matrix
        // (быстрый, но неточный — посчитаем на отдельном /api/assets/:id/controls)
        // Пока — заглушка: итог через /api/assets потом уточняем.
        Promise.all(
          (Array.isArray(matrix) ? matrix : []).map((row: any) =>
            authFetch(`/api/assets/${row.asset_id}/controls`).then(r => r.ok ? r.json() : [])
          )
        ).then((perAsset: any[]) => {
          const map = new Map<number, number>();
          for (const list of perAsset) {
            if (Array.isArray(list)) {
              for (const c of list) {
                map.set(c.id, (map.get(c.id) || 0) + 1);
              }
            }
          }
          setUsage(map);
        });
      })
      .catch((e: Error) => setErr(e.message))
      .finally(() => setLoading(false));
  }, []);

  const controlsByGroup = useMemo(() => {
    const groups = new Map<string, Control[]>();
    for (const c of controls) {
      const g = MEASURE_GROUP[c.name] || "Прочие";
      if (!groups.has(g)) groups.set(g, []);
      groups.get(g)!.push(c);
    }
    return Array.from(groups.entries()).sort();
  }, [controls]);

  if (loading) return <div style={{ padding: 24 }}>Загружаем каталог СЗИ…</div>;
  if (err) return <div style={{ padding: 24, color: "var(--risk-critical)" }}>{err}</div>;

  return (
    <div style={{ padding: "20px 24px", maxWidth: 1400, margin: "0 auto" }}>
      <h1 style={{ fontSize: 24, fontWeight: 700, margin: 0, color: "var(--fg)" }}>
        Каталог средств защиты информации
      </h1>
      <div style={{ color: "var(--fg-muted)", fontSize: 13, marginTop: 4, marginBottom: 16 }}>
        База программно-аппаратных СЗИ (БД ПАСЗИ из ПТСЗИ): 11 контролей по
        диплому, что закрывают, на скольких активах внедрены, какие требования
        ФСТЭК-17 / ISO 27001 покрывают.
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        {controlsByGroup.map(([groupName, items]) => (
          <Card
            key={groupName}
            title={
              <span>
                <Icon name="shield" /> Мероприятия — {groupName}
              </span>
            }
          >
            <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(360px, 1fr))", gap: 12 }}>
              {items.map((c) => {
                const usageCount = usage.get(c.id) || 0;
                const vls = CONTROL_TO_VLS[c.name] || [];
                const stds = CONTROL_STDS[c.name] || { fstec: [], iso: [] };
                return (
                  <div
                    key={c.id}
                    style={{
                      border: "1px solid var(--border)",
                      borderRadius: "var(--r-md)",
                      padding: 14,
                      background: "var(--bg-elev-1)",
                      display: "flex",
                      flexDirection: "column",
                      gap: 10,
                    }}
                  >
                    <div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between" }}>
                      <div style={{ fontWeight: 700, fontSize: 14 }}>{c.name}</div>
                      <Chip tone={usageCount > 0 ? "success" : "neutral"} mono>
                        внедрено: {usageCount}
                      </Chip>
                    </div>

                    {c.description && (
                      <div style={{ fontSize: 12, color: "var(--fg-muted)" }}>{c.description}</div>
                    )}

                    {vls.length > 0 && (
                      <div>
                        <div style={{ fontSize: 11, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.4, marginBottom: 4 }}>
                          Закрывает уязвимые звенья
                        </div>
                        <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
                          {vls.map((vlID) => (
                            <Chip key={vlID} tone="accent">
                              {VL_NAMES[vlID]}
                            </Chip>
                          ))}
                        </div>
                      </div>
                    )}

                    {(stds.fstec.length > 0 || stds.iso.length > 0) && (
                      <div>
                        <div style={{ fontSize: 11, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.4, marginBottom: 4 }}>
                          Покрывает требования
                        </div>
                        {stds.fstec.length > 0 && (
                          <div style={{ display: "flex", flexWrap: "wrap", gap: 3, marginBottom: 4 }}>
                            <span style={{ fontSize: 11, color: "var(--fg-muted)", marginRight: 4 }}>ФСТЭК-17:</span>
                            {stds.fstec.map((code) => (
                              <span key={code} style={{ fontSize: 11, fontFamily: "var(--font-mono)", color: "var(--fg-muted)" }}>
                                {code}
                              </span>
                            ))}
                          </div>
                        )}
                        {stds.iso.length > 0 && (
                          <div style={{ display: "flex", flexWrap: "wrap", gap: 3 }}>
                            <span style={{ fontSize: 11, color: "var(--fg-muted)", marginRight: 4 }}>ISO 27001:</span>
                            {stds.iso.map((code) => (
                              <span key={code} style={{ fontSize: 11, fontFamily: "var(--font-mono)", color: "var(--fg-muted)" }}>
                                {code}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </Card>
        ))}
      </div>

      <div style={{ marginTop: 16, padding: 14, background: "var(--bg-elev-2)", borderRadius: "var(--r-md)", fontSize: 12, color: "var(--fg-muted)" }}>
        Чтобы внедрить контроль на актив, перейдите в карточку актива →
        секция «Внедрённые контроли». Список загружен из таблицы <code>controls</code>;
        соответствие требованиям — из таблицы <code>requirement_controls</code>.
      </div>
    </div>
  );
};

export default ControlCatalogPage;
