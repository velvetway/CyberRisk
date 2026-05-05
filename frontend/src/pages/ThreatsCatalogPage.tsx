import React, { useEffect, useMemo, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api, AssetTypeRef, ThreatFull, ThreatUpdatePayload } from "../api/client";
import { Btn, Card, Chip, Icon, IconBtn } from "../components/design";

type SourceFilter = "all" | "external" | "internal" | "third_party";

const sourceLabel = (s: string): string => ({
  external: "внешний",
  internal: "внутренний",
  third_party: "третья сторона",
}[s] ?? s);

export const ThreatsCatalogPage: React.FC = () => {
  const [threats, setThreats] = useState<ThreatFull[]>([]);
  const [assetTypes, setAssetTypes] = useState<AssetTypeRef[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);

  // фильтры
  const [search, setSearch] = useState("");
  const [filterAT, setFilterAT] = useState<number | "all">("all");
  const [filterSource, setFilterSource] = useState<SourceFilter>("all");
  const [filterImpact, setFilterImpact] = useState<"" | "C" | "I" | "A">("");

  // редактор
  const [editing, setEditing] = useState<ThreatFull | null>(null);

  const reload = async () => {
    setErr(null);
    try {
      const [t, at] = await Promise.all([
        api.getThreatsAll(500, 0),
        api.getAssetTypes(),
      ]);
      setThreats(Array.isArray(t) ? t : []);
      setAssetTypes(Array.isArray(at) ? at : []);
    } catch (e: any) {
      setErr(e.message ?? String(e));
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => { reload(); }, []);

  const atByID = useMemo(() => {
    const m = new Map<number, AssetTypeRef>();
    for (const a of assetTypes) m.set(a.id, a);
    return m;
  }, [assetTypes]);

  const filtered = useMemo(() => {
    const s = search.trim().toLowerCase();
    return threats.filter(t => {
      if (filterSource !== "all" && t.source_type !== filterSource) return false;
      if (filterAT !== "all") {
        if (!t.applies_to_asset_types || !t.applies_to_asset_types.includes(filterAT)) {
          return false;
        }
      }
      if (filterImpact === "C" && !t.impact_c) return false;
      if (filterImpact === "I" && !t.impact_i) return false;
      if (filterImpact === "A" && !t.impact_a) return false;
      if (s) {
        const haystack = (t.bdu_id || "") + " " + t.name + " " + (t.description || "");
        if (!haystack.toLowerCase().includes(s)) return false;
      }
      return true;
    });
  }, [threats, search, filterAT, filterSource, filterImpact]);

  if (loading) return <div style={{ padding: 24, color: "var(--fg-muted)" }}>Загрузка каталога…</div>;
  if (err) return <div style={{ padding: 24, color: "var(--risk-critical)" }}>Ошибка: {err}</div>;

  return (
    <div style={{ padding: "20px 24px", maxWidth: 1400, margin: "0 auto" }}>
      <div style={{ display: "flex", alignItems: "baseline", gap: 12, marginBottom: 16 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700, color: "var(--fg)", margin: 0 }}>Каталог угроз ФСТЭК</h1>
        <span style={{ color: "var(--fg-muted)", fontSize: 13 }}>{filtered.length} из {threats.length}</span>
      </div>

      <Card style={{ marginBottom: 16, display: "flex", gap: 8, flexWrap: "wrap" }} pad={12}>
        <input
          value={search}
          onChange={e => setSearch(e.target.value)}
          placeholder="Поиск: УБИ.001, шифрование, перехват…"
          style={{
            flex: "1 1 280px", padding: "8px 12px", borderRadius: 8,
            border: "1px solid var(--border)", background: "var(--bg-elev-1)",
            color: "var(--fg)", fontSize: 14,
          }}
        />
        <select
          value={filterSource}
          onChange={e => setFilterSource(e.target.value as SourceFilter)}
          style={selectStyle}
        >
          <option value="all">Любой источник</option>
          <option value="external">внешний</option>
          <option value="internal">внутренний</option>
          <option value="third_party">третья сторона</option>
        </select>
        <select
          value={filterAT === "all" ? "all" : String(filterAT)}
          onChange={e => setFilterAT(e.target.value === "all" ? "all" : Number(e.target.value))}
          style={selectStyle}
        >
          <option value="all">Любой тип актива</option>
          {assetTypes.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}
        </select>
        <select
          value={filterImpact}
          onChange={e => setFilterImpact(e.target.value as any)}
          style={selectStyle}
        >
          <option value="">Любое воздействие</option>
          <option value="C">нарушает C</option>
          <option value="I">нарушает I</option>
          <option value="A">нарушает A</option>
        </select>
      </Card>

      <Card pad={0}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
            <thead>
              <tr style={{ background: "var(--bg-elev-2)", textAlign: "left" }}>
                <th style={th}>УБИ</th>
                <th style={th}>Название</th>
                <th style={th}>Источник</th>
                <th style={th}>Q^threat</th>
                <th style={th}>q^severity</th>
                <th style={th}>CIA</th>
                <th style={th}>Применима к</th>
                <th style={th}></th>
              </tr>
            </thead>
            <tbody>
              {filtered.slice(0, 200).map(t => (
                <tr key={t.id} style={{ borderTop: "1px solid var(--border)" }}>
                  <td style={td}><code style={{ fontSize: 12 }}>{t.bdu_id || `#${t.id}`}</code></td>
                  <td style={{ ...td, maxWidth: 480 }}>
                    <div style={{ fontWeight: 600 }}>{t.name}</div>
                    {t.applies_to_targets && (
                      <div style={{ color: "var(--fg-muted)", fontSize: 11, marginTop: 2 }}>
                        {t.applies_to_targets.length > 80 ? t.applies_to_targets.slice(0, 80) + "…" : t.applies_to_targets}
                      </div>
                    )}
                  </td>
                  <td style={td}><Chip tone="ghost">{sourceLabel(t.source_type)}</Chip></td>
                  <td style={td}><strong>{t.q_threat.toFixed(2)}</strong></td>
                  <td style={td}><strong>{t.q_severity.toFixed(2)}</strong></td>
                  <td style={td}>
                    <span style={{ display: "inline-flex", gap: 4 }}>
                      {t.impact_c && <Chip tone="warn">C</Chip>}
                      {t.impact_i && <Chip tone="warn">I</Chip>}
                      {t.impact_a && <Chip tone="warn">A</Chip>}
                    </span>
                  </td>
                  <td style={td}>
                    {t.applies_to_asset_types && t.applies_to_asset_types.length > 0 ? (
                      <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
                        {t.applies_to_asset_types.map(id => (
                          <Chip key={id} tone="neutral">{atByID.get(id)?.name ?? `#${id}`}</Chip>
                        ))}
                      </div>
                    ) : (
                      <span style={{ color: "var(--fg-muted)", fontSize: 12 }}>универсальная</span>
                    )}
                  </td>
                  <td style={td}>
                    <IconBtn title="Редактировать applicability" onClick={() => setEditing(t)}>
                      <Icon name="edit" />
                    </IconBtn>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {filtered.length > 200 && (
            <div style={{ padding: 12, color: "var(--fg-muted)", textAlign: "center", fontSize: 12 }}>
              Показаны первые 200 из {filtered.length}. Уточните фильтры.
            </div>
          )}
        </div>
      </Card>

      <AnimatePresence>
        {editing && (
          <ThreatEditor
            threat={editing}
            assetTypes={assetTypes}
            onClose={() => setEditing(null)}
            onSaved={async () => {
              setEditing(null);
              await reload();
            }}
          />
        )}
      </AnimatePresence>
    </div>
  );
};

const ThreatEditor: React.FC<{
  threat: ThreatFull;
  assetTypes: AssetTypeRef[];
  onClose: () => void;
  onSaved: () => void | Promise<void>;
}> = ({ threat, assetTypes, onClose, onSaved }) => {
  const [selected, setSelected] = useState<Set<number>>(
    new Set(threat.applies_to_asset_types ?? []),
  );
  const [appliesTo, setAppliesTo] = useState(threat.applies_to_targets ?? "");
  const [qThreat, setQThreat] = useState(threat.q_threat);
  const [qSeverity, setQSeverity] = useState(threat.q_severity);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const toggle = (id: number) => {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id); else next.add(id);
    setSelected(next);
  };

  const save = async () => {
    setBusy(true);
    setErr(null);
    try {
      const payload: ThreatUpdatePayload = {
        name: threat.name,
        threat_category_id: threat.threat_category_id ?? null,
        source_type: threat.source_type,
        description: threat.description ?? null,
        q_threat: qThreat,
        q_severity: qSeverity,
        bdu_id: threat.bdu_id ?? null,
        applies_to_targets: appliesTo || null,
        applies_to_asset_types: Array.from(selected),
        impact_c: threat.impact_c,
        impact_i: threat.impact_i,
        impact_a: threat.impact_a,
      };
      await api.updateThreat(threat.id, payload);
      await onSaved();
    } catch (e: any) {
      setErr(e.message ?? String(e));
    } finally {
      setBusy(false);
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
      style={{
        position: "fixed", inset: 0, background: "rgba(0,0,0,.45)", zIndex: 80,
        display: "flex", alignItems: "flex-start", justifyContent: "center", paddingTop: 60,
      }}
      onClick={onClose}
    >
      <div onClick={e => e.stopPropagation()} style={{
        width: 640, background: "var(--bg-elev-2)", borderRadius: 12,
        padding: 16, border: "1px solid var(--border)", maxHeight: "80vh", overflowY: "auto",
      }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
          <div>
            <div style={{ fontSize: 11, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.5 }}>
              {threat.bdu_id || `#${threat.id}`}
            </div>
            <strong style={{ fontSize: 16 }}>{threat.name}</strong>
          </div>
          <IconBtn onClick={onClose}><Icon name="x" /></IconBtn>
        </div>

        {threat.description && (
          <div style={{ color: "var(--fg-muted)", fontSize: 13, marginBottom: 12 }}>
            {threat.description}
          </div>
        )}

        <div style={{ marginBottom: 12, display: "flex", gap: 12 }}>
          <NumberField label="Q^threat" value={qThreat} onChange={setQThreat} />
          <NumberField label="q^severity" value={qSeverity} onChange={setQSeverity} />
        </div>

        <div style={{ marginBottom: 12 }}>
          <div style={{ fontSize: 12, color: "var(--fg-faint)", marginBottom: 6, textTransform: "uppercase", letterSpacing: 0.5 }}>
            Применима к типам активов
          </div>
          <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
            {assetTypes.map(a => (
              <Chip
                key={a.id}
                tone={selected.has(a.id) ? "success" : "ghost"}
                onClick={() => toggle(a.id)}
              >
                {selected.has(a.id) ? "✓ " : "+ "}{a.name}
              </Chip>
            ))}
          </div>
          <div style={{ marginTop: 6, fontSize: 11, color: "var(--fg-faint)" }}>
            Пусто = угроза универсальная (применима ко всем).
          </div>
        </div>

        <div style={{ marginBottom: 12 }}>
          <div style={{ fontSize: 12, color: "var(--fg-faint)", marginBottom: 6, textTransform: "uppercase", letterSpacing: 0.5 }}>
            Объект воздействия (raw text из ФСТЭК)
          </div>
          <textarea
            value={appliesTo}
            onChange={e => setAppliesTo(e.target.value)}
            rows={3}
            style={{
              width: "100%", padding: "8px 10px", borderRadius: 8,
              border: "1px solid var(--border)", background: "var(--bg-elev-1)",
              color: "var(--fg)", fontSize: 13, resize: "vertical",
            }}
          />
        </div>

        {err && <div style={{ color: "var(--risk-critical)", marginBottom: 8, fontSize: 12 }}>{err}</div>}

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <Btn variant="ghost" onClick={onClose}>Отмена</Btn>
          <Btn variant="primary" onClick={save} disabled={busy}>{busy ? "Сохраняю…" : "Сохранить"}</Btn>
        </div>
      </div>
    </motion.div>
  );
};

const NumberField: React.FC<{ label: string; value: number; onChange: (v: number) => void }> = ({ label, value, onChange }) => (
  <div style={{ flex: 1 }}>
    <div style={{ fontSize: 11, color: "var(--fg-faint)", marginBottom: 4 }}>{label}</div>
    <input
      type="number" min={0} max={1} step={0.05} value={value}
      onChange={e => onChange(Number(e.target.value))}
      style={{
        width: "100%", padding: "6px 10px", borderRadius: 6,
        border: "1px solid var(--border)", background: "var(--bg-elev-1)",
        color: "var(--fg)", fontSize: 14,
      }}
    />
  </div>
);

const selectStyle: React.CSSProperties = {
  padding: "8px 12px", borderRadius: 8,
  border: "1px solid var(--border)", background: "var(--bg-elev-1)",
  color: "var(--fg)", fontSize: 14, minWidth: 160,
};

const th: React.CSSProperties = { padding: "10px 12px", fontSize: 11, fontWeight: 600, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.5 };
const td: React.CSSProperties = { padding: "10px 12px", verticalAlign: "top" };

export default ThreatsCatalogPage;
