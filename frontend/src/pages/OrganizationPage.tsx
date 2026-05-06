// OrganizationPage — сводный «дашборд организации».
// Закрывает требование «не один актив = один отчёт, а вся организация»:
// агрегированные метрики по всем активам, таблица активов, топ
// критических рисков, скачивание единого PDF.
import React, { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  api,
  authFetch,
  OrganizationAssetRow,
  OrganizationCriticalRisk,
  OrganizationOverview,
} from "../api/client";
import { Btn, Card, Chip, Icon, RiskBadge } from "../components/design";

const LEVEL_TONE: Record<string, "success" | "warn" | "danger" | "neutral"> = {
  critical: "danger",
  high: "warn",
  medium: "warn",
  low: "success",
};

export const OrganizationPage: React.FC = () => {
  const navigate = useNavigate();
  const [overview, setOverview] = useState<OrganizationOverview | null>(null);
  const [matrix, setMatrix] = useState<OrganizationAssetRow[]>([]);
  const [critical, setCritical] = useState<OrganizationCriticalRisk[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      api.getOrganizationOverview(),
      api.getOrganizationAssetMatrix(),
      api.getOrganizationCriticalRisks(15),
    ])
      .then(([o, m, c]) => {
        setOverview(o);
        setMatrix(m);
        setCritical(c);
      })
      .catch((e: Error) => setErr(e.message))
      .finally(() => setLoading(false));
  }, []);

  const downloadPDF = async () => {
    try {
      const res = await authFetch("/api/organization/report.pdf");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = "organization-report.pdf";
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (e) {
      setErr(`PDF: ${(e as Error).message}`);
    }
  };

  if (loading) return <div style={{ padding: 24 }}>Загружаем сводку…</div>;
  if (err) return <div style={{ padding: 24, color: "var(--risk-critical)" }}>Ошибка: {err}</div>;
  if (!overview) return <div style={{ padding: 24 }}>Нет данных</div>;

  return (
    <div style={{ padding: "20px 24px", maxWidth: 1400, margin: "0 auto" }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: 16 }}>
        <div>
          <h1 style={{ fontSize: 24, fontWeight: 700, margin: 0, color: "var(--fg)" }}>Организация</h1>
          <div style={{ color: "var(--fg-muted)", fontSize: 13, marginTop: 4 }}>
            Сводные метрики по всем активам · модель риска ПТСЗИ
          </div>
        </div>
        <Btn variant="primary" onClick={downloadPDF} icon={<Icon name="file" />}>
          Скачать сводный PDF
        </Btn>
      </div>

      {/* ----- METRIC TILES ----- */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: 12, marginBottom: 16 }}>
        <MetricTile label="Активов" value={overview.total_assets.toString()} accent="#1971c2" />
        <MetricTile label="Изолированных" value={`${overview.isolated_assets} / ${overview.total_assets}`} accent="#7048e8" />
        <MetricTile label="W max" value={overview.w_max.toFixed(2)} accent={wColor(overview.w_max)} sub={overview.w_max_asset} />
        <MetricTile label="Средний W_max" value={overview.avg_w_per_asset.toFixed(2)} accent={wColor(overview.avg_w_per_asset)} />
        <MetricTile label="Внедрено СЗИ" value={overview.total_controls.toString()} accent="#2f9e44" />
        <MetricTile label="Непокрытых VL" value={overview.uncovered_vls.toString()} accent={overview.uncovered_vls > 0 ? "#c92a2a" : "#2f9e44"} />
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16, marginBottom: 16 }}>
        {/* ----- RISK DISTRIBUTION ----- */}
        <Card title={<span><Icon name="alert" /> Распределение по уровню риска</span>}>
          <RiskDistributionBar dist={overview.risk_distribution} total={overview.total_assets} />
        </Card>

        {/* ----- COMPLIANCE BY STANDARD ----- */}
        <Card title={<span><Icon name="award" /> Среднее соответствие стандартам</span>}>
          {overview.compliance_by_standard.length === 0 ? (
            <div style={{ color: "var(--fg-faint)" }}>Стандарты не настроены.</div>
          ) : (
            <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
              {overview.compliance_by_standard.map((s) => (
                <ComplianceBar key={s.standard.code} data={s} />
              ))}
            </div>
          )}
        </Card>
      </div>

      {/* ----- ASSET MATRIX ----- */}
      <Card title={<span><Icon name="layers" /> Активы организации ({matrix.length})</span>} style={{ marginBottom: 16 }}>
        <div style={{ overflowX: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
            <thead>
              <tr style={{ borderBottom: "1px solid var(--border)", color: "var(--fg-faint)", fontSize: 11, textTransform: "uppercase", letterSpacing: 0.4 }}>
                <th style={th}>Актив</th>
                <th style={th}>Тип</th>
                <th style={th}>Среда</th>
                <th style={th}>Контур</th>
                <th style={{ ...th, textAlign: "right" }}>W max</th>
                <th style={th}>Уровень</th>
                <th style={{ ...th, textAlign: "right" }}>Угроз</th>
                <th style={{ ...th, textAlign: "right" }}>СЗИ</th>
              </tr>
            </thead>
            <tbody>
              {matrix.map((m) => (
                <tr
                  key={m.asset_id}
                  onClick={() => navigate(`/assets/${m.asset_id}`)}
                  style={{
                    borderBottom: "1px solid var(--border)",
                    cursor: "pointer",
                    transition: "var(--transition)",
                  }}
                  onMouseEnter={(e) => ((e.currentTarget as HTMLTableRowElement).style.background = "var(--bg-elev-2)")}
                  onMouseLeave={(e) => ((e.currentTarget as HTMLTableRowElement).style.background = "transparent")}
                >
                  <td style={td}>{m.name}</td>
                  <td style={td}>{m.type_name || "—"}</td>
                  <td style={td}>{m.environment || "—"}</td>
                  <td style={td}>{m.is_isolated ? "изолирован" : "общий"}</td>
                  <td style={{ ...td, textAlign: "right", fontFamily: "var(--font-mono)", fontWeight: 600 }}>
                    {m.w_max.toFixed(2)}
                  </td>
                  <td style={td}>
                    <Chip tone={LEVEL_TONE[m.level] || "neutral"} mono>{levelLabel(m.level)}</Chip>
                  </td>
                  <td style={{ ...td, textAlign: "right", fontFamily: "var(--font-mono)" }}>{m.threat_count}</td>
                  <td style={{ ...td, textAlign: "right", fontFamily: "var(--font-mono)" }}>{m.control_count}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* ----- CRITICAL RISKS ----- */}
      <Card title={<span><Icon name="alert" /> Топ критических рисков (asset × threat)</span>}>
        {critical.length === 0 ? (
          <div style={{ color: "var(--fg-faint)" }}>Критических рисков не выявлено.</div>
        ) : (
          <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
            {critical.map((r, i) => (
              <div
                key={`${r.asset_id}-${r.threat_id}`}
                onClick={() => navigate(`/risk/graph/${r.asset_id}`)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 12,
                  padding: "10px 12px",
                  border: "1px solid var(--border)",
                  borderLeft: `3px solid ${wColor(r.w)}`,
                  borderRadius: "var(--r-sm)",
                  background: "var(--bg-elev-1)",
                  cursor: "pointer",
                }}
              >
                <span style={{ minWidth: 26, fontFamily: "var(--font-mono)", color: "var(--fg-faint)" }}>
                  #{i + 1}
                </span>
                <RiskBadge level={(r.level as any) || "low"} score={r.w} compact />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 600, fontSize: 13 }}>{r.asset_name}</div>
                  <div style={{ fontSize: 12, color: "var(--fg-muted)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                    ← {r.bdu_id ? <span style={{ fontFamily: "var(--font-mono)", marginRight: 6 }}>{r.bdu_id}</span> : null}
                    {r.threat_name}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  );
};

const th: React.CSSProperties = { padding: "8px 10px", textAlign: "left", fontWeight: 600 };
const td: React.CSSProperties = { padding: "8px 10px" };

const MetricTile: React.FC<{ label: string; value: string; sub?: string; accent: string }> = ({
  label,
  value,
  sub,
  accent,
}) => (
  <div
    style={{
      background: "var(--bg-elev-1)",
      border: "1px solid var(--border)",
      borderTop: `3px solid ${accent}`,
      borderRadius: "var(--r-md)",
      padding: 14,
      display: "flex",
      flexDirection: "column",
      gap: 4,
    }}
  >
    <div style={{ fontSize: 11, color: "var(--fg-faint)", textTransform: "uppercase", letterSpacing: 0.4 }}>{label}</div>
    <div style={{ fontSize: 24, fontWeight: 700, fontFamily: "var(--font-mono)", color: accent }}>{value}</div>
    {sub && <div style={{ fontSize: 11, color: "var(--fg-muted)", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{sub}</div>}
  </div>
);

const RiskDistributionBar: React.FC<{ dist: Record<string, number>; total: number }> = ({ dist, total }) => {
  if (total === 0) return <div style={{ color: "var(--fg-faint)" }}>Нет активов</div>;
  const items: { level: string; label: string; color: string; count: number }[] = [
    { level: "critical", label: "Критические", color: "#c92a2a", count: dist.critical || 0 },
    { level: "high", label: "Высокие", color: "#e8590c", count: dist.high || 0 },
    { level: "medium", label: "Средние", color: "#f59f00", count: dist.medium || 0 },
    { level: "low", label: "Низкие", color: "#2f9e44", count: dist.low || 0 },
  ];

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 10 }}>
      <div style={{ display: "flex", height: 14, borderRadius: 6, overflow: "hidden", border: "1px solid var(--border)" }}>
        {items.map((it) => {
          const pct = (100 * it.count) / total;
          if (pct === 0) return null;
          return (
            <div key={it.level} style={{ flex: pct, background: it.color }} title={`${it.label}: ${it.count}`} />
          );
        })}
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 4, fontSize: 12 }}>
        {items.map((it) => (
          <div key={it.level} style={{ display: "flex", alignItems: "center", gap: 8 }}>
            <span style={{ width: 10, height: 10, borderRadius: 2, background: it.color }} />
            <span style={{ flex: 1 }}>{it.label}</span>
            <span style={{ fontFamily: "var(--font-mono)", fontWeight: 600 }}>{it.count}</span>
            <span style={{ width: 40, textAlign: "right", color: "var(--fg-faint)" }}>
              {((100 * it.count) / total).toFixed(0)}%
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

const ComplianceBar: React.FC<{ data: { standard: { name: string; code: string }; avg_score: number; min_score: number; max_score: number; assets_count: number } }> = ({ data }) => {
  const pct = data.avg_score * 100;
  const c = wColor(1 - data.avg_score); // больше = хуже соответствует красной шкале → инвертируем для compliance
  const color = pct >= 80 ? "#2f9e44" : pct >= 50 ? "#f59f00" : pct >= 25 ? "#e8590c" : "#c92a2a";
  void c;
  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
      <div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between" }}>
        <span style={{ fontWeight: 600, fontSize: 13 }}>{data.standard.name}</span>
        <span style={{ fontWeight: 700, fontFamily: "var(--font-mono)", color }}>{pct.toFixed(0)}%</span>
      </div>
      <div style={{ height: 6, background: "var(--bg-elev-3)", borderRadius: 3, overflow: "hidden" }}>
        <div style={{ width: `${pct}%`, height: "100%", background: color, transition: "width .3s" }} />
      </div>
      <div style={{ display: "flex", gap: 12, fontSize: 11, color: "var(--fg-faint)" }}>
        <span>min: {(data.min_score * 100).toFixed(0)}%</span>
        <span>max: {(data.max_score * 100).toFixed(0)}%</span>
        <span style={{ marginLeft: "auto" }}>{data.assets_count} активов</span>
      </div>
    </div>
  );
};

function wColor(w: number): string {
  if (w >= 0.75) return "#c92a2a";
  if (w >= 0.5) return "#e8590c";
  if (w >= 0.25) return "#f59f00";
  return "#2f9e44";
}

function levelLabel(level: string): string {
  switch (level) {
    case "critical": return "Критич.";
    case "high":     return "Высокий";
    case "medium":   return "Средний";
    case "low":      return "Низкий";
    default:         return level;
  }
}

export default OrganizationPage;
