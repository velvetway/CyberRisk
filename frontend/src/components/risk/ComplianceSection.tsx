// ComplianceSection — секция «Соответствие стандартам ИБ» на странице
// карточки актива. Отвечает за блок ОСЗ (оценка состояния защищённости)
// из 7.png диплома: показывает % соответствия активa каждому стандарту
// (ФСТЭК-17, ISO 27001) и раскрывающийся список требований с пометкой,
// какие закрыты внедрёнными контролями, а какие ещё нет.
import React, { useEffect, useMemo, useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  api,
  authFetch,
  AssetComplianceOverview,
  AssetStandardCompliance,
  Control,
  RequirementStatus,
} from "../../api/client";
import { Btn, Card, Chip, Icon } from "../design";

interface Props {
  assetID: number;
}

export const ComplianceSection: React.FC<Props> = ({ assetID }) => {
  const [overview, setOverview] = useState<AssetComplianceOverview[] | null>(null);
  const [activeCode, setActiveCode] = useState<string | null>(null);
  const [detail, setDetail] = useState<AssetStandardCompliance | null>(null);
  const [loadingOverview, setLoadingOverview] = useState(true);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  // Загружаем сводку по всем стандартам один раз.
  useEffect(() => {
    let cancelled = false;
    setLoadingOverview(true);
    setErr(null);
    api
      .getAssetCompliance(assetID)
      .then((data) => {
        if (cancelled) return;
        setOverview(data);
        // Авто-выбор первого стандарта для drill-down
        if (data && data.length > 0 && !activeCode) {
          setActiveCode(data[0].standard.code);
        }
      })
      .catch((e: Error) => {
        if (!cancelled) setErr(e.message);
      })
      .finally(() => {
        if (!cancelled) setLoadingOverview(false);
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [assetID]);

  // Подгружаем детали по выбранному стандарту.
  useEffect(() => {
    if (!activeCode) return;
    let cancelled = false;
    setLoadingDetail(true);
    api
      .getAssetComplianceDetail(assetID, activeCode)
      .then((data) => {
        if (!cancelled) setDetail(data);
      })
      .catch((e: Error) => {
        if (!cancelled) setErr(e.message);
      })
      .finally(() => {
        if (!cancelled) setLoadingDetail(false);
      });
    return () => {
      cancelled = true;
    };
  }, [assetID, activeCode]);

  const downloadPDF = async () => {
    try {
      const res = await authFetch(`/api/compliance/asset/${assetID}/report.pdf`);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `compliance-asset-${assetID}.pdf`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (e) {
      setErr(`PDF: ${(e as Error).message}`);
    }
  };

  return (
    <Card
      title={
        <span>
          <Icon name="award" /> Соответствие стандартам ИБ
        </span>
      }
      subtitle="Оценка состояния защищённости (ОСЗ) по требованиям ФСТЭК и ISO 27001 на основе внедрённых на активе контролей."
      action={
        overview && overview.length > 0 ? (
          <Btn variant="outline" size="sm" onClick={downloadPDF} icon={<Icon name="file" />}>
            Скачать PDF
          </Btn>
        ) : undefined
      }
    >
      {err && (
        <div style={{ color: "var(--risk-critical)", padding: 8 }}>{err}</div>
      )}

      {loadingOverview && !overview ? (
        <div style={{ color: "var(--fg-faint)", padding: 16 }}>Загружаем…</div>
      ) : overview && overview.length === 0 ? (
        <div style={{ color: "var(--fg-faint)", padding: 16 }}>
          В системе нет настроенных стандартов соответствия.
        </div>
      ) : (
        overview && (
          <>
            {/* Карточки-плитки со сводкой */}
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fit, minmax(280px, 1fr))",
                gap: 12,
                marginBottom: 16,
              }}
            >
              {overview.map((o) => (
                <StandardTile
                  key={o.standard.code}
                  data={o}
                  active={activeCode === o.standard.code}
                  onClick={() => setActiveCode(o.standard.code)}
                />
              ))}
            </div>

            {/* Детальный разрез */}
            {activeCode && (
              <div
                style={{
                  borderTop: "1px solid var(--border)",
                  paddingTop: 14,
                }}
              >
                {loadingDetail || !detail || detail.standard.code !== activeCode ? (
                  <div style={{ color: "var(--fg-faint)", padding: 8 }}>Загружаем требования…</div>
                ) : (
                  <RequirementsList detail={detail} />
                )}
              </div>
            )}
          </>
        )
      )}
    </Card>
  );
};

// ----- Tile (одна плитка стандарта со сводкой) -----

const StandardTile: React.FC<{
  data: AssetComplianceOverview;
  active: boolean;
  onClick: () => void;
}> = ({ data, active, onClick }) => {
  const pct = data.overall_score * 100;
  const color = scoreColor(data.overall_score);

  return (
    <button
      onClick={onClick}
      style={{
        textAlign: "left",
        background: active ? "var(--bg-elev-2)" : "var(--bg-elev-1)",
        border: `1px solid ${active ? color.fg : "var(--border)"}`,
        borderRadius: "var(--r-md)",
        padding: 14,
        cursor: "pointer",
        transition: "var(--transition)",
        display: "flex",
        flexDirection: "column",
        gap: 10,
      }}
    >
      <div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between" }}>
        <div>
          <div style={{ fontWeight: 700, fontSize: 14 }}>{data.standard.name}</div>
          <div style={{ fontSize: 11, color: "var(--fg-faint)", marginTop: 2 }}>
            {data.standard.jurisdiction === "RU" ? "Россия" : "Международный"}
          </div>
        </div>
        <div style={{ fontWeight: 800, fontSize: 22, color: color.fg, fontFamily: "var(--font-mono)" }}>
          {pct.toFixed(0)}%
        </div>
      </div>

      <div
        aria-label="progress"
        style={{
          height: 8,
          borderRadius: 4,
          background: "var(--bg-elev-3)",
          overflow: "hidden",
          position: "relative",
        }}
      >
        <div
          style={{
            position: "absolute",
            top: 0,
            left: 0,
            bottom: 0,
            width: `${Math.max(pct, 0)}%`,
            background: color.fg,
            borderRadius: 4,
            transition: "width .25s",
          }}
        />
      </div>

      <div style={{ display: "flex", gap: 6, flexWrap: "wrap", fontSize: 11 }}>
        <Chip tone="success" mono>
          ✓ {data.covered_count}
        </Chip>
        <Chip tone="warn" mono>
          ◐ {data.partial_count}
        </Chip>
        <Chip tone="danger" mono>
          ✗ {data.uncovered_count}
        </Chip>
        <span style={{ color: "var(--fg-faint)", marginLeft: "auto" }}>
          из {data.total_count}
        </span>
      </div>
    </button>
  );
};

// ----- Список требований (drill-down) -----

const RequirementsList: React.FC<{ detail: AssetStandardCompliance }> = ({ detail }) => {
  // Группируем по категории.
  const groups = useMemo(() => {
    const m = new Map<string, RequirementStatus[]>();
    for (const r of detail.requirements) {
      const k = r.requirement.category || "Прочее";
      if (!m.has(k)) m.set(k, []);
      m.get(k)!.push(r);
    }
    return Array.from(m.entries());
  }, [detail]);

  const [filter, setFilter] = useState<"all" | "covered" | "partial" | "uncovered">("all");
  const [expanded, setExpanded] = useState<Set<number>>(new Set());

  const toggle = (id: number) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const matchFilter = (r: RequirementStatus): boolean => {
    if (filter === "all") return true;
    if (filter === "covered") return r.coverage >= 1.0;
    if (filter === "partial") return r.coverage > 0 && r.coverage < 1.0;
    return r.coverage <= 0;
  };

  return (
    <div>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 12, flexWrap: "wrap", gap: 12 }}>
        <div>
          <div style={{ fontSize: 13, fontWeight: 700 }}>{detail.standard.name}</div>
          <div style={{ fontSize: 11, color: "var(--fg-faint)", maxWidth: 600 }}>{detail.standard.full_name}</div>
        </div>
        <div style={{ display: "flex", gap: 6, fontSize: 12 }}>
          <FilterPill label="Все" active={filter === "all"} onClick={() => setFilter("all")} count={detail.total_count} />
          <FilterPill label="✓ Выполнено" tone="success" active={filter === "covered"} onClick={() => setFilter("covered")} count={detail.covered_count} />
          <FilterPill label="◐ Частично" tone="warn" active={filter === "partial"} onClick={() => setFilter("partial")} count={detail.partial_count} />
          <FilterPill label="✗ Не выполнено" tone="danger" active={filter === "uncovered"} onClick={() => setFilter("uncovered")} count={detail.uncovered_count} />
        </div>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        {groups.map(([cat, items]) => {
          const visible = items.filter(matchFilter);
          if (visible.length === 0) return null;
          return (
            <div key={cat}>
              <div
                style={{
                  fontSize: 11,
                  letterSpacing: 0.5,
                  textTransform: "uppercase",
                  color: "var(--fg-faint)",
                  marginBottom: 6,
                  fontWeight: 600,
                }}
              >
                {cat} <span style={{ opacity: 0.6 }}>· {visible.length}</span>
              </div>
              <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
                {visible.map((r) => (
                  <RequirementRow
                    key={r.requirement.id}
                    rs={r}
                    expanded={expanded.has(r.requirement.id)}
                    onToggle={() => toggle(r.requirement.id)}
                  />
                ))}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

const FilterPill: React.FC<{ label: string; active: boolean; count: number; tone?: "success" | "warn" | "danger"; onClick: () => void }> = ({
  label,
  active,
  count,
  tone,
  onClick,
}) => {
  const c = tone === "success"
    ? { fg: "var(--risk-low)", bg: "var(--risk-low-bg)", br: "var(--risk-low-br)" }
    : tone === "warn"
    ? { fg: "var(--risk-medium)", bg: "var(--risk-medium-bg)", br: "var(--risk-medium-br)" }
    : tone === "danger"
    ? { fg: "var(--risk-critical)", bg: "var(--risk-critical-bg)", br: "var(--risk-critical-br)" }
    : { fg: "var(--fg-muted)", bg: "var(--bg-elev-3)", br: "var(--border)" };
  return (
    <button
      onClick={onClick}
      style={{
        background: active ? c.bg : "transparent",
        color: c.fg,
        border: `1px solid ${active ? c.br : "var(--border)"}`,
        borderRadius: "var(--r-sm)",
        padding: "4px 10px",
        cursor: "pointer",
        fontSize: 12,
        fontWeight: 500,
        display: "inline-flex",
        gap: 6,
        alignItems: "center",
      }}
    >
      {label}
      <span style={{ opacity: 0.7, fontFamily: "var(--font-mono)" }}>{count}</span>
    </button>
  );
};

const RequirementRow: React.FC<{
  rs: RequirementStatus;
  expanded: boolean;
  onToggle: () => void;
}> = ({ rs, expanded, onToggle }) => {
  const r = rs.requirement;
  const colors = scoreColor(rs.coverage);
  const status: "covered" | "partial" | "uncovered" =
    rs.coverage >= 1.0 ? "covered" : rs.coverage > 0 ? "partial" : "uncovered";
  const statusIcon = status === "covered" ? "✓" : status === "partial" ? "◐" : "✗";
  const priorityLabel = ["", "критич.", "средн.", "низк."][r.priority] || "";

  return (
    <div
      style={{
        border: `1px solid var(--border)`,
        borderLeft: `3px solid ${colors.fg}`,
        borderRadius: "var(--r-sm)",
        background: "var(--bg-elev-1)",
        overflow: "hidden",
      }}
    >
      <button
        onClick={onToggle}
        style={{
          width: "100%",
          background: "transparent",
          border: "none",
          padding: "10px 12px",
          display: "flex",
          alignItems: "center",
          gap: 12,
          cursor: "pointer",
          textAlign: "left",
        }}
      >
        <span
          style={{
            display: "inline-flex",
            justifyContent: "center",
            alignItems: "center",
            width: 22,
            height: 22,
            borderRadius: "50%",
            background: colors.bg,
            color: colors.fg,
            fontWeight: 700,
            fontSize: 12,
            border: `1px solid ${colors.br}`,
          }}
        >
          {statusIcon}
        </span>
        <span style={{ fontFamily: "var(--font-mono)", fontWeight: 600, fontSize: 12, minWidth: 70, color: "var(--fg)" }}>
          {r.code}
        </span>
        <span style={{ flex: 1, color: "var(--fg)" }}>{r.title}</span>
        {r.priority === 1 && (
          <Chip tone="danger" mono>
            {priorityLabel}
          </Chip>
        )}
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 12, color: colors.fg, fontWeight: 700, minWidth: 40, textAlign: "right" }}>
          {(rs.coverage * 100).toFixed(0)}%
        </span>
        <span style={{ color: "var(--fg-faint)", marginLeft: 4 }}>{expanded ? "▾" : "▸"}</span>
      </button>

      <AnimatePresence initial={false}>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.15 }}
            style={{ overflow: "hidden" }}
          >
            <div
              style={{
                padding: "0 14px 12px",
                color: "var(--fg-muted)",
                fontSize: 13,
                lineHeight: 1.5,
                borderTop: "1px solid var(--border)",
                paddingTop: 10,
              }}
            >
              {r.description && <div style={{ marginBottom: 10 }}>{r.description}</div>}

              <ControlBlock
                title="Закрыто внедрёнными контролями"
                items={rs.covering_controls || []}
                emptyText="нет внедрённых контролей, закрывающих это требование"
                tone="success"
              />
              <ControlBlock
                title="Что ещё может закрыть"
                items={rs.missing_controls || []}
                emptyText="у этого требования нет дополнительных рекомендуемых контролей"
                tone="warn"
              />
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

const ControlBlock: React.FC<{
  title: string;
  items: Control[];
  emptyText: string;
  tone: "success" | "warn";
}> = ({ title, items, emptyText, tone }) => (
  <div style={{ marginTop: 8 }}>
    <div style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: 0.4, color: "var(--fg-faint)", marginBottom: 6 }}>
      {title}
    </div>
    {items.length === 0 ? (
      <div style={{ color: "var(--fg-faint)", fontSize: 12, fontStyle: "italic" }}>{emptyText}</div>
    ) : (
      <div style={{ display: "flex", gap: 6, flexWrap: "wrap" }}>
        {items.map((c) => (
          <Chip key={c.id} tone={tone}>
            {c.name}
          </Chip>
        ))}
      </div>
    )}
  </div>
);

// ----- утилита цвета по % -----

function scoreColor(s: number): { fg: string; bg: string; br: string } {
  if (s >= 0.8) return { fg: "var(--risk-low)", bg: "var(--risk-low-bg)", br: "var(--risk-low-br)" };
  if (s >= 0.5) return { fg: "var(--risk-medium)", bg: "var(--risk-medium-bg)", br: "var(--risk-medium-br)" };
  if (s >= 0.25) return { fg: "var(--risk-high)", bg: "var(--risk-high-bg)", br: "var(--risk-high-br)" };
  return { fg: "var(--risk-critical)", bg: "var(--risk-critical-bg)", br: "var(--risk-critical-br)" };
}

export default ComplianceSection;
