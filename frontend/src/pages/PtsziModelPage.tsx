import React, { useEffect, useMemo, useState } from "react";
import { useLocation } from "react-router-dom";
import { api } from "../api/client";
import { Asset } from "../types";
import { PtsziAssetProfile, PtsziAttackPath, PtsziControl, PtsziVulnerableLink } from "../types/ptszi";
import { Btn, Card, Chip, Icon, RiskBadge } from "../components/design";

export const PtsziModelPage: React.FC = () => {
  const location = useLocation();
  const [assets, setAssets] = useState<Asset[]>([]);
  const [assetId, setAssetId] = useState<number | null>(null);
  const [profile, setProfile] = useState<PtsziAssetProfile | null>(null);
  const [allVL, setAllVL] = useState<PtsziVulnerableLink[]>([]);
  const [allControls, setAllControls] = useState<PtsziControl[]>([]);
  const [selectedThreat, setSelectedThreat] = useState<PtsziAttackPath | null>(null);
  const [vlIDs, setVlIDs] = useState<number[]>([]);
  const [controlIDs, setControlIDs] = useState<number[]>([]);
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([api.getAssets(), api.getPtsziVulnerableLinks(), api.getPtsziControls()])
      .then(([assetList, links, controls]) => {
        setAssets(assetList);
        setAllVL(links);
        setAllControls(controls);
        const requestedID = Number(new URLSearchParams(location.search).get("asset"));
        const requestedAsset = assetList.find(a => a.id === requestedID);
        if (requestedAsset) {
          setAssetId(requestedAsset.id);
        } else if (assetList.length) {
          setAssetId(assetList[0].id);
        }
      })
      .catch(e => setErr(e.message));
  }, [location.search]);

  useEffect(() => {
    if (!assetId) return;
    setErr(null);
    api.getAssetPtsziProfile(assetId)
      .then(p => {
        setProfile(p);
        setVlIDs(p.vulnerable_links.map(v => v.vulnerable_link.id));
        setControlIDs(p.controls.map(c => c.control.id));
        setSelectedThreat(p.applicable_threats[0] || null);
      })
      .catch(e => setErr(e.message));
  }, [assetId]);

  const implementedControls = useMemo(() => new Set(controlIDs), [controlIDs]);

  const saveProfile = async () => {
    if (!assetId) return;
    setSaving(true);
    try {
      await api.updateAssetPtsziVulnerableLinks(assetId, vlIDs);
      await api.updateAssetPtsziControls(assetId, controlIDs.map(id => ({ control_id: id, effectiveness: 1 })));
      const next = await api.getAssetPtsziProfile(assetId);
      setProfile(next);
      setSelectedThreat(next.applicable_threats[0] || null);
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setSaving(false);
    }
  };

  const toggle = (id: number, values: number[], setValues: (v: number[]) => void) => {
    setValues(values.includes(id) ? values.filter(x => x !== id) : [...values, id]);
  };

  return (
    <div style={{ padding: 20, display: "flex", flexDirection: "column", gap: 16 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-end", gap: 16 }}>
        <div>
          <div style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)", fontFamily: "var(--font-mono)", textTransform: "uppercase" }}>ПТСЗИ · каноническая модель</div>
          <h1 style={{ margin: 0, fontSize: "var(--text-2xl)", fontWeight: 600 }}>{"S -> ST -> VL -> методы -> DA -> W"}</h1>
          <div style={{ marginTop: 4, color: "var(--fg-muted)", fontSize: "var(--text-sm)" }}>
            УБИ ФСТЭК используются как справочник источника, объекта воздействия и последствий C/I/A.
          </div>
        </div>
        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          <select
            value={assetId ?? ""}
            onChange={e => setAssetId(Number(e.target.value))}
            style={{ height: 32, background: "var(--bg-elev-2)", color: "var(--fg)", border: "1px solid var(--border)", borderRadius: "var(--r-sm)", padding: "0 10px" }}
          >
            {assets.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}
          </select>
          <Btn onClick={saveProfile} disabled={!assetId || saving} icon={<Icon name="refresh" size={13} />}>{saving ? "Сохранение" : "Пересчитать"}</Btn>
        </div>
      </div>

      {err && <Card dense><div style={{ color: "var(--risk-critical)" }}>{err}</div></Card>}

      <div style={{ display: "grid", gridTemplateColumns: "340px 1fr 360px", gap: 16 }}>
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <Card title="Уязвимые звенья VL" subtitle="Что реально присутствует на активе" dense>
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              {allVL.map(v => (
                <label key={v.id} style={{ display: "grid", gridTemplateColumns: "18px 44px 1fr", gap: 8, alignItems: "start", fontSize: "var(--text-sm)" }}>
                  <input type="checkbox" checked={vlIDs.includes(v.id)} onChange={() => toggle(v.id, vlIDs, setVlIDs)} />
                  <span className="mono" style={{ color: "var(--accent)" }}>{v.code}</span>
                  <span>{v.name}</span>
                </label>
              ))}
            </div>
          </Card>

          <Card title="Методы противодействия" subtitle="Внедренные A/FW/IDS/..." dense>
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8 }}>
              {allControls.map(c => (
                <label key={c.id} style={{ display: "flex", gap: 7, alignItems: "center", fontSize: "var(--text-sm)" }}>
                  <input type="checkbox" checked={controlIDs.includes(c.id)} onChange={() => toggle(c.id, controlIDs, setControlIDs)} />
                  <span className="mono" style={{ color: "var(--accent)" }}>{c.code}</span>
                  <span>{c.name}</span>
                </label>
              ))}
            </div>
          </Card>
        </div>

        <Card title="Применимые угрозы ST" subtitle={profile ? `${profile.applicable_threats.length} сценариев для контура ${profile.security_contour}` : "Загрузка"} dense>
          <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
            {(profile?.applicable_threats || []).map(t => (
              <button
                key={t.threat.id}
                onClick={() => setSelectedThreat(t)}
                style={{
                  textAlign: "left",
                  background: selectedThreat?.threat.id === t.threat.id ? "var(--bg-active)" : "var(--bg-elev-2)",
                  border: "1px solid var(--border)",
                  borderRadius: "var(--r-md)",
                  padding: 12,
                  color: "var(--fg)",
                  cursor: "pointer",
                }}
              >
                <div style={{ display: "flex", justifyContent: "space-between", gap: 10, alignItems: "center" }}>
                  <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
                    <Chip mono tone="accent">{t.threat.code}</Chip>
                    <strong>{t.threat.name}</strong>
                  </div>
                  <RiskBadge level={t.level} score={t.w} compact />
                </div>
                <div style={{ display: "flex", gap: 6, flexWrap: "wrap", marginTop: 8 }}>
                  <Chip mono>VL={t.vulnerable_links.length}</Chip>
                  <Chip mono>примеры УБИ={t.ubi.length}</Chip>
                  <Chip mono tone={t.missing_controls.length > 0 ? "warn" : "success"}>меры: {t.missing_controls.length > 0 ? `нужно ${t.missing_controls.length}` : "закрыто"}</Chip>
                </div>
              </button>
            ))}
          </div>
        </Card>

        <Card title="Сценарий" subtitle={selectedThreat ? selectedThreat.threat.name : "Выберите ST"} dense>
          {selectedThreat ? (
            <div style={{ display: "flex", flexDirection: "column", gap: 14 }}>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8 }}>
                <SummaryBox label="Итоговый риск">
                  <RiskBadge level={selectedThreat.level} score={selectedThreat.w} compact />
                </SummaryBox>
                <SummaryBox label="Состояние защиты">
                  <Chip mono tone={selectedThreat.missing_controls.length > 0 ? "warn" : "success"}>
                    {selectedThreat.missing_controls.length > 0 ? `не хватает ${selectedThreat.missing_controls.length}` : "закрыто"}
                  </Chip>
                </SummaryBox>
              </div>

              <Section title="Источники S">
                {selectedThreat.sources.map(s => <Chip key={s.code} mono tone="accent">{s.code} {s.name}</Chip>)}
              </Section>

              <Section title="Актуальные VL">
                {selectedThreat.vulnerable_links.map(v => (
                  <div key={v.vulnerable_link.id} style={{ padding: 8, border: "1px solid var(--border)", borderRadius: "var(--r-sm)" }}>
                    <div style={{ display: "flex", justifyContent: "space-between", gap: 8 }}>
                      <strong className="mono">{v.vulnerable_link.code}</strong>
                      <Chip mono tone={v.coverage > 0 ? "success" : "danger"}>{v.coverage > 0 ? "есть защита" : "нет защиты"}</Chip>
                    </div>
                    <div style={{ fontSize: "var(--text-xs)", color: "var(--fg-muted)", marginTop: 3 }}>{v.vulnerable_link.name}</div>
                    <div style={{ display: "flex", gap: 5, flexWrap: "wrap", marginTop: 6 }}>
                      {v.controls.map(c => (
                        <Chip key={c.control.id} mono tone={implementedControls.has(c.control.id) ? "success" : "ghost"}>
                          {c.control.code}
                        </Chip>
                      ))}
                    </div>
                  </div>
                ))}
              </Section>

              <Section title="УБИ ФСТЭК">
                {selectedThreat.ubi.slice(0, 5).map(u => (
                  <div key={u.ubi_code} style={{ fontSize: "var(--text-xs)", borderBottom: "1px solid var(--border)", paddingBottom: 7 }}>
                    <div style={{ display: "flex", gap: 6, alignItems: "center", marginBottom: 3 }}>
                      <Chip mono tone="warn">{u.ubi_code}</Chip>
                      <span>{u.name}</span>
                    </div>
                    <div style={{ color: "var(--fg-dim)" }}>Источник: {u.source_raw || "не указан"} · Объект: {u.impact_object || "не указан"}</div>
                  </div>
                ))}
              </Section>

              <Section title="Недостающие методы">
                {(selectedThreat.missing_controls || []).map(c => <Chip key={c.id} mono tone="danger">{c.code} {c.name}</Chip>)}
              </Section>
            </div>
          ) : (
            <div style={{ color: "var(--fg-muted)" }}>Нет применимых сценариев для выбранного актива.</div>
          )}
        </Card>
      </div>
    </div>
  );
};

const SummaryBox: React.FC<{ label: string; children: React.ReactNode }> = ({ label, children }) => (
  <div style={{ border: "1px solid var(--border)", borderRadius: "var(--r-sm)", padding: 9, background: "var(--bg-elev-2)" }}>
    <div style={{ fontSize: "var(--text-2xs)", color: "var(--fg-dim)", fontFamily: "var(--font-mono)", textTransform: "uppercase" }}>{label}</div>
    <div style={{ marginTop: 6 }}>{children}</div>
  </div>
);

const Section: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <div>
    <div style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)", textTransform: "uppercase", fontFamily: "var(--font-mono)", marginBottom: 7 }}>{title}</div>
    <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>{children}</div>
  </div>
);
