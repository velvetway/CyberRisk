import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { motion, AnimatePresence } from "framer-motion";
import {
  api,
  AssetSoftwareLink,
  AssetVulnerability,
  Control,
} from "../api/client";
import { Asset, Software } from "../types";
import { Btn, Card, Chip, Icon, IconBtn, RiskBadge } from "../components/design";

// 6 категорий VL из диплома; индекс соответствует id (после миграции 034
// SMALLSERIAL стартует с 1).
const VL_LABELS: Record<number, { code: string; name: string }> = {
  1: { code: "VL1", name: "Нештатное доп. ПО" },
  2: { code: "VL2", name: "Устаревшее ПО / уязвимости" },
  3: { code: "VL3", name: "Недекларируемое ПО" },
  4: { code: "VL4", name: "Обход админом" },
  5: { code: "VL5", name: "Носители информации" },
  6: { code: "VL6", name: "Открытые ОС / отсутствие защиты ЛВС" },
};

interface VLBucket {
  vlID: number;
  code: string;
  name: string;
  items: AssetVulnerability[];
}

function bucketByVL(items: AssetVulnerability[]): VLBucket[] {
  const buckets = new Map<number, VLBucket>();
  for (const av of items) {
    const id = av.vl_category_id ?? 0;
    if (!buckets.has(id)) {
      const meta = VL_LABELS[id] ?? { code: id ? `VL${id}` : "—", name: "Без категории" };
      buckets.set(id, { vlID: id, code: meta.code, name: meta.name, items: [] });
    }
    buckets.get(id)!.items.push(av);
  }
  return Array.from(buckets.values()).sort((a, b) => a.vlID - b.vlID);
}

// Возвращает Chip-tone — должен совпадать с ChipTone из Primitives.tsx
// (neutral | accent | success | warn | danger | ghost). Если вернуть кастомный
// — Chip упадёт на toneStyles[undefined].bg.
function levelTone(score: number | undefined): "danger" | "warn" | "neutral" {
  if (score == null) return "neutral";
  if (score >= 7) return "danger";   // critical/high CVSS
  if (score >= 4) return "warn";     // medium CVSS
  return "neutral";
}

export const AssetDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const assetID = Number(id);

  const [asset, setAsset] = useState<Asset | null>(null);
  const [software, setSoftware] = useState<AssetSoftwareLink[]>([]);
  const [vulns, setVulns] = useState<AssetVulnerability[]>([]);
  const [controls, setControls] = useState<Control[]>([]);
  const [aggregate, setAggregate] = useState<{ w_max: number; level: string; threat_count: number; uncovered_count: number } | null>(null);

  const [loading, setLoading] = useState(true);
  const [recomputing, setRecomputing] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  const reload = useCallback(async () => {
    setErr(null);
    try {
      const [a, sw, vv, ctl, ap] = await Promise.all([
        api.getAsset(assetID),
        api.getAssetSoftware(assetID),
        api.getAssetVulnerabilities(assetID),
        api.getAssetControls(assetID),
        api.getAssetAttackPaths(assetID).catch(() => null),
      ]);
      setAsset(a);
      setSoftware(sw ?? []);
      setVulns(vv ?? []);
      setControls(ctl ?? []);
      setAggregate(ap?.aggregate ?? null);
    } catch (e: any) {
      setErr(e.message ?? String(e));
    } finally {
      setLoading(false);
    }
  }, [assetID]);

  useEffect(() => {
    if (!Number.isFinite(assetID) || assetID <= 0) {
      setErr("Некорректный id актива");
      setLoading(false);
      return;
    }
    reload();
  }, [assetID, reload]);

  const buckets = useMemo(() => bucketByVL(vulns), [vulns]);
  const cveCount = vulns.length;
  const autoCount = vulns.filter(v => v.source.startsWith("auto")).length;

  const recomputeRisk = async () => {
    setRecomputing(true);
    try {
      const ap = await api.getAssetAttackPaths(assetID);
      setAggregate(ap?.aggregate ?? null);
    } catch (e: any) {
      setErr(e.message ?? String(e));
    } finally {
      setRecomputing(false);
    }
  };

  if (loading) {
    return <div style={{ padding: 24, color: "var(--fg-muted)" }}>Загрузка карточки актива…</div>;
  }
  if (err) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <div style={{ color: "var(--risk-critical)", fontWeight: 600, marginBottom: 8 }}>Ошибка</div>
          <div style={{ color: "var(--fg-muted)" }}>{err}</div>
          <Btn variant="ghost" onClick={() => navigate("/assets")} style={{ marginTop: 12 }}>← К списку активов</Btn>
        </Card>
      </div>
    );
  }
  if (!asset) {
    return <div style={{ padding: 24 }}>Актив не найден</div>;
  }

  return (
    <div style={{ padding: "20px 24px", maxWidth: 1200, margin: "0 auto" }}>
      {/* ----- Header ----- */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16 }}>
        <div>
          <Btn variant="ghost" size="sm" onClick={() => navigate("/assets")} icon={<Icon name="arrowL" />}>
            К списку активов
          </Btn>
          <h1 style={{ fontSize: 24, fontWeight: 700, margin: "8px 0 4px", color: "var(--fg)" }}>{asset.name}</h1>
          <div style={{ display: "flex", gap: 8, color: "var(--fg-muted)", fontSize: 13 }}>
            <span>id #{asset.id}</span>
            <span>•</span>
            <span>{asset.environment || "—"}</span>
            <span>•</span>
            <span>{asset.is_isolated ? "Изолированный (Z=0.5)" : "Открытый (Z=1.0)"}</span>
          </div>
        </div>
        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          {aggregate && (
            <RiskBadge level={(aggregate.level as any) || "low"} score={aggregate.w_max} />
          )}
          <Btn variant="primary" onClick={recomputeRisk} disabled={recomputing} icon={<Icon name="refresh" />}>
            {recomputing ? "Считаю…" : "Пересчитать W"}
          </Btn>
        </div>
      </div>

      {/* ----- Aggregate strip ----- */}
      {aggregate && (
        <Card style={{ marginBottom: 16, display: "flex", gap: 24, alignItems: "center" }} pad={14}>
          <div>
            <div style={{ fontSize: 11, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.5 }}>W max</div>
            <div style={{ fontSize: 22, fontWeight: 700 }}>{aggregate.w_max.toFixed(2)}</div>
          </div>
          <div style={{ width: 1, height: 32, background: "var(--border)" }} />
          <div>
            <div style={{ fontSize: 11, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.5 }}>Угроз</div>
            <div style={{ fontSize: 22, fontWeight: 600 }}>{aggregate.threat_count}</div>
          </div>
          <div style={{ width: 1, height: 32, background: "var(--border)" }} />
          <div>
            <div style={{ fontSize: 11, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.5 }}>Непокрытых VL</div>
            <div style={{ fontSize: 22, fontWeight: 600, color: aggregate.uncovered_count > 0 ? "var(--risk-high)" : "var(--fg)" }}>{aggregate.uncovered_count}</div>
          </div>
          <div style={{ width: 1, height: 32, background: "var(--border)" }} />
          <div>
            <div style={{ fontSize: 11, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.5 }}>Свидетельств CVE</div>
            <div style={{ fontSize: 22, fontWeight: 600 }}>{cveCount}</div>
          </div>
          <div style={{ marginLeft: "auto" }}>
            <Btn variant="outline" size="sm" onClick={() => navigate(`/risk/graph/${assetID}`)} icon={<Icon name="flow" />}>
              Граф атак
            </Btn>
          </div>
        </Card>
      )}

      {/* ----- 3 секции ----- */}
      <div style={{ display: "grid", gridTemplateColumns: "1fr", gap: 16 }}>
        <SoftwareSection assetID={assetID} items={software} onChange={reload} />
        <VulnSection assetID={assetID} buckets={buckets} totalCves={cveCount} autoCount={autoCount} onChange={reload} />
        <ControlSection assetID={assetID} attached={controls} onChange={reload} />
      </div>
    </div>
  );
};

// =====================================================================
// Section 1: Установленное ПО
// =====================================================================

const SoftwareSection: React.FC<{
  assetID: number;
  items: AssetSoftwareLink[];
  onChange: () => Promise<void> | void;
}> = ({ assetID, items, onChange }) => {
  const [picker, setPicker] = useState(false);
  const [busyID, setBusyID] = useState<number | null>(null);
  const [status, setStatus] = useState<string | null>(null);

  const detach = async (softwareID: number) => {
    setBusyID(softwareID);
    setStatus(null);
    try {
      await api.detachSoftware(assetID, softwareID);
      await onChange();
    } catch (e: any) {
      setStatus(e.message);
    } finally {
      setBusyID(null);
    }
  };

  return (
    <Card
      title={<span><Icon name="package" /> Установленное ПО ({items.length})</span>}
      action={<Btn size="sm" variant="primary" icon={<Icon name="plus" />} onClick={() => setPicker(true)}>Добавить ПО</Btn>}
    >
      {items.length === 0 && (
        <div style={{ color: "var(--fg-muted)", padding: "12px 0" }}>
          Никакого ПО не привязано. Кнопка «Добавить ПО» включит автодетекцию CVE из БДУ ФСТЭК.
        </div>
      )}
      {items.map(it => (
        <div key={it.link.id} style={{
          display: "flex", justifyContent: "space-between", alignItems: "center",
          padding: "10px 0", borderTop: "1px solid var(--border)",
        }}>
          <div>
            <div style={{ fontWeight: 600 }}>{it.software.name}</div>
            <div style={{ color: "var(--fg-muted)", fontSize: 12 }}>
              {it.software.vendor} {it.link.version ? ` • ${it.link.version}` : ""}
              {it.software.is_russian && <Chip tone="success">РФ</Chip>}
              {it.software.fstec_certified && <Chip tone="accent">ФСТЭК</Chip>}
            </div>
          </div>
          <IconBtn title="Открепить" onClick={() => detach(it.software.id)} >
            {busyID === it.software.id ? "…" : <Icon name="trash" />}
          </IconBtn>
        </div>
      ))}
      {status && <div style={{ marginTop: 8, color: "var(--risk-critical)", fontSize: 12 }}>{status}</div>}

      <AnimatePresence>
        {picker && (
          <SoftwarePicker
            onClose={() => setPicker(false)}
            onPick={async (sw, version) => {
              try {
                const r = await api.attachSoftware(assetID, sw.id, version || undefined);
                const verHint = version ? ` (версия ${version})` : "";
                setStatus(`✓ ${sw.name}${verHint}: найдено ${r.detected_vulnerabilities} БДУ-уязвимостей`);
                setPicker(false);
                await onChange();
              } catch (e: any) {
                setStatus(`✗ ${e.message}`);
              }
            }}
          />
        )}
      </AnimatePresence>
    </Card>
  );
};

const SoftwarePicker: React.FC<{
  onClose: () => void;
  onPick: (sw: Software, version: string) => Promise<void> | void;
}> = ({ onClose, onPick }) => {
  const [q, setQ] = useState("");
  const [results, setResults] = useState<Software[]>([]);
  const [searching, setSearching] = useState(false);
  const [picked, setPicked] = useState<Software | null>(null);
  const [version, setVersion] = useState("");

  useEffect(() => {
    if (q.trim().length < 2) {
      setResults([]);
      return;
    }
    const t = setTimeout(async () => {
      setSearching(true);
      try {
        const res = await fetch(`/api/software?search=${encodeURIComponent(q.trim())}&limit=20`, {
          headers: { Authorization: `Bearer ${localStorage.getItem("token") ?? ""}` },
        });
        const data = await res.json();
        setResults(Array.isArray(data) ? data : []);
      } finally {
        setSearching(false);
      }
    }, 250);
    return () => clearTimeout(t);
  }, [q]);

  return (
    <motion.div
      initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}
      style={{
        position: "fixed", inset: 0, background: "rgba(0,0,0,.45)",
        zIndex: 80, display: "flex", alignItems: "flex-start", justifyContent: "center",
        paddingTop: 80,
      }}
      onClick={onClose}
    >
      <div onClick={e => e.stopPropagation()} style={{
        width: 560, background: "var(--bg-elev-2)", borderRadius: 12,
        padding: 16, border: "1px solid var(--border)",
      }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
          <strong>{picked ? "Версия установленного ПО" : "Поиск ПО в каталоге"}</strong>
          <IconBtn onClick={onClose}><Icon name="x" /></IconBtn>
        </div>

        {picked ? (
          <div>
            <div style={{
              padding: 12, borderRadius: 8, background: "var(--bg-elev-1)",
              border: "1px solid var(--border)", marginBottom: 12,
            }}>
              <div style={{ fontWeight: 600 }}>{picked.name}</div>
              <div style={{ color: "var(--fg-muted)", fontSize: 12 }}>{picked.vendor}</div>
            </div>
            <label style={{ fontSize: 12, color: "var(--fg-faint)", marginBottom: 4, display: "block", textTransform: "uppercase", letterSpacing: 0.5 }}>
              Версия (опционально)
            </label>
            <input
              autoFocus
              value={version}
              onChange={e => setVersion(e.target.value)}
              placeholder="например: 1.7, 8.0.21, 20.04 LTS"
              style={{
                width: "100%", padding: "10px 12px", borderRadius: 8,
                border: "1px solid var(--border)", background: "var(--bg-elev-1)",
                color: "var(--fg)", fontSize: 14, marginBottom: 8,
              }}
              onKeyDown={e => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  onPick(picked, version.trim());
                }
              }}
            />
            <div style={{ color: "var(--fg-muted)", fontSize: 11, marginBottom: 12 }}>
              Версия фильтрует БДУ-уязвимости (точное совпадение или диапазон «от X до Y»).
              Пусто → подтянуть всё для этого продукта.
            </div>
            <div style={{ display: "flex", justifyContent: "space-between", gap: 8 }}>
              <Btn variant="ghost" size="sm" onClick={() => { setPicked(null); setVersion(""); }}>← Назад к поиску</Btn>
              <Btn variant="primary" size="sm" onClick={() => onPick(picked, version.trim())}>
                Привязать
              </Btn>
            </div>
          </div>
        ) : (
          <>
            <input
              autoFocus
              value={q}
              onChange={e => setQ(e.target.value)}
              placeholder="Astra Linux, 1С, КриптоПро…"
              style={{
                width: "100%", padding: "10px 12px", borderRadius: 8,
                border: "1px solid var(--border)", background: "var(--bg-elev-1)",
                color: "var(--fg)", marginBottom: 8, fontSize: 14,
              }}
            />
            {searching && <div style={{ color: "var(--fg-muted)", fontSize: 12 }}>Поиск…</div>}
            <div style={{ maxHeight: 380, overflowY: "auto" }}>
              {results.map(sw => (
                <div key={sw.id} style={{
                  padding: 10, borderBottom: "1px solid var(--border)",
                  cursor: "pointer", borderRadius: 6,
                }}
                  onClick={() => setPicked(sw)}
                >
                  <div style={{ fontWeight: 600 }}>{sw.name}</div>
                  <div style={{ color: "var(--fg-muted)", fontSize: 12 }}>{sw.vendor}</div>
                </div>
              ))}
              {!searching && q.trim().length >= 2 && results.length === 0 && (
                <div style={{ color: "var(--fg-muted)", padding: 16, textAlign: "center" }}>Ничего не найдено</div>
              )}
            </div>
          </>
        )}
      </div>
    </motion.div>
  );
};

// =====================================================================
// Section 2: Уязвимости и VL-категории
// =====================================================================

// Только валидные ChipTone (см. Primitives.tsx):
// neutral | accent | success | warn | danger | ghost.
const VlChipColor: Record<string, "warn" | "danger" | "neutral" | "accent"> = {
  VL1: "warn",
  VL2: "danger",
  VL3: "danger",
  VL4: "warn",
  VL5: "neutral",
  VL6: "danger",
};

const VulnSection: React.FC<{
  assetID: number;
  buckets: VLBucket[];
  totalCves: number;
  autoCount: number;
  onChange: () => Promise<void> | void;
}> = ({ assetID, buckets, totalCves, autoCount, onChange }) => {
  const [expanded, setExpanded] = useState<Set<number>>(new Set());
  const [showManualForm, setShowManualForm] = useState(false);
  const [bduInput, setBduInput] = useState("");
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState<string | null>(null);

  const toggle = (id: number) => {
    const next = new Set(expanded);
    if (next.has(id)) next.delete(id); else next.add(id);
    setExpanded(next);
  };

  const addManual = async () => {
    const id = bduInput.trim();
    if (!id) return;
    setBusy(true);
    setStatus(null);
    try {
      await api.addManualVulnerability(assetID, id);
      setBduInput("");
      setShowManualForm(false);
      await onChange();
    } catch (e: any) {
      setStatus(e.message);
    } finally {
      setBusy(false);
    }
  };

  const removeOne = async (vulnID: number) => {
    try {
      await api.deleteAssetVulnerability(assetID, vulnID);
      await onChange();
    } catch (e: any) {
      setStatus(e.message);
    }
  };

  return (
    <Card
      title={<span><Icon name="alert" /> Уязвимости и VL-категории ({totalCves})</span>}
      subtitle={<span style={{ color: "var(--fg-muted)", fontSize: 12 }}>авто: {autoCount}, вручную: {totalCves - autoCount}</span>}
      action={<Btn size="sm" variant="ghost" icon={<Icon name="plus" />} onClick={() => setShowManualForm(v => !v)}>Добавить БДУ-id</Btn>}
    >
      {showManualForm && (
        <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
          <input
            value={bduInput}
            onChange={e => setBduInput(e.target.value)}
            placeholder="BDU:2024-00123"
            style={{
              flex: 1, padding: "8px 10px", borderRadius: 8,
              border: "1px solid var(--border)", background: "var(--bg-elev-1)", color: "var(--fg)",
            }}
          />
          <Btn size="sm" variant="primary" onClick={addManual} disabled={busy}>{busy ? "…" : "Добавить"}</Btn>
        </div>
      )}

      {buckets.length === 0 && (
        <div style={{ color: "var(--fg-muted)", padding: "12px 0" }}>
          Уязвимости не обнаружены. Привяжите ПО — автодетекция найдёт совпадения с БДУ ФСТЭК.
        </div>
      )}

      {buckets.map(b => {
        const isOpen = expanded.has(b.vlID);
        const tone = VlChipColor[b.code] ?? "neutral";
        return (
          <div key={b.vlID} style={{ borderTop: "1px solid var(--border)", padding: "12px 0" }}>
            <div
              onClick={() => toggle(b.vlID)}
              style={{ display: "flex", alignItems: "center", gap: 12, cursor: "pointer" }}
            >
              <Chip tone={tone as any}>{b.code}</Chip>
              <strong>{b.name}</strong>
              <Chip tone="neutral">{b.items.length} CVE</Chip>
              <span style={{ marginLeft: "auto", color: "var(--fg-muted)", fontSize: 12 }}>
                {isOpen ? "▼" : "▶"}
              </span>
            </div>
            {isOpen && (
              <div style={{ marginTop: 8 }}>
                {b.items.slice(0, 50).map(av => (
                  <div key={av.id} style={{
                    display: "flex", justifyContent: "space-between", alignItems: "center",
                    padding: "6px 0 6px 28px", borderTop: "1px dashed var(--border-faint, var(--border))", fontSize: 13,
                  }}>
                    <div style={{ display: "flex", gap: 8, alignItems: "center", flexWrap: "wrap" }}>
                      <code style={{ fontSize: 12, color: "var(--fg)" }}>{av.bdu_id}</code>
                      {av.cwe && <Chip tone="ghost">{av.cwe}</Chip>}
                      {av.cvss_score != null && (
                        <Chip tone={levelTone(av.cvss_score) as any}>CVSS {av.cvss_score.toFixed(1)}</Chip>
                      )}
                      {av.source.startsWith("auto") && <Chip tone="accent">авто</Chip>}
                      {av.title && <span style={{ color: "var(--fg-muted)" }}>{av.title.length > 60 ? av.title.slice(0, 60) + "…" : av.title}</span>}
                    </div>
                    <IconBtn title="Удалить запись" onClick={() => removeOne(av.id)}>
                      <Icon name="trash" />
                    </IconBtn>
                  </div>
                ))}
                {b.items.length > 50 && (
                  <div style={{ color: "var(--fg-muted)", padding: 8 }}>… и ещё {b.items.length - 50} записей</div>
                )}
              </div>
            )}
          </div>
        );
      })}

      {status && <div style={{ color: "var(--risk-critical)", marginTop: 8, fontSize: 12 }}>{status}</div>}
    </Card>
  );
};

// =====================================================================
// Section 3: Внедрённые контроли
// =====================================================================

const ControlSection: React.FC<{
  assetID: number;
  attached: Control[];
  onChange: () => Promise<void> | void;
}> = ({ assetID, attached, onChange }) => {
  const [catalog, setCatalog] = useState<Control[]>([]);
  const [busyID, setBusyID] = useState<number | null>(null);
  const [status, setStatus] = useState<string | null>(null);

  useEffect(() => {
    api.getControls().then(setCatalog).catch(e => setStatus(e.message));
  }, []);

  const attachedIDs = useMemo(() => new Set(attached.map(c => c.id)), [attached]);
  const available = catalog.filter(c => !attachedIDs.has(c.id));

  const attach = async (controlID: number) => {
    setBusyID(controlID);
    try {
      await api.attachControl(assetID, controlID);
      await onChange();
    } catch (e: any) {
      setStatus(e.message);
    } finally {
      setBusyID(null);
    }
  };

  const detach = async (controlID: number) => {
    setBusyID(controlID);
    try {
      await api.detachControl(assetID, controlID);
      await onChange();
    } catch (e: any) {
      setStatus(e.message);
    } finally {
      setBusyID(null);
    }
  };

  return (
    <Card title={<span><Icon name="shield" /> Внедрённые контроли ({attached.length})</span>}>
      <div style={{ marginBottom: 12 }}>
        <div style={{ fontSize: 12, color: "var(--fg-faint)", marginBottom: 6, textTransform: "uppercase", letterSpacing: 0.5 }}>Внедрено</div>
        {attached.length === 0 && (
          <div style={{ color: "var(--fg-muted)", padding: "8px 0" }}>
            Контролей нет. Без них Q^reaction = 0 и формула W даёт максимум.
          </div>
        )}
        <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
          {attached.map(c => (
            <Chip key={c.id} tone="success" onClick={() => detach(c.id)}>
              {c.name} <span style={{ marginLeft: 6, opacity: 0.7 }}>{busyID === c.id ? "…" : "✕"}</span>
            </Chip>
          ))}
        </div>
      </div>

      <div>
        <div style={{ fontSize: 12, color: "var(--fg-faint)", marginBottom: 6, textTransform: "uppercase", letterSpacing: 0.5 }}>Каталог</div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: 8 }}>
          {available.map(c => (
            <Chip key={c.id} tone="ghost" onClick={() => attach(c.id)}>
              + {c.name}
            </Chip>
          ))}
          {available.length === 0 && (
            <span style={{ color: "var(--fg-muted)", fontSize: 12 }}>Все контроли каталога уже внедрены.</span>
          )}
        </div>
      </div>

      {status && <div style={{ color: "var(--risk-critical)", marginTop: 8, fontSize: 12 }}>{status}</div>}
    </Card>
  );
};

export default AssetDetailPage;
