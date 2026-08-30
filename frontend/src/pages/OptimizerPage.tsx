import React, { useEffect, useMemo, useState } from "react";
import { useLocation } from "react-router-dom";
import { api } from "../api/client";
import { Asset } from "../types";
import { AssetScale, OptimizerPlan, Roadmap } from "../types/optimizer";
import { Btn, Card, Chip, Icon, StatCard } from "../components/design";

const money = (v: number) =>
  new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 0 }).format(v) + " ₽";

const UNIT_LABEL: Record<string, string> = {
  node: "за рабочее место",
  server: "за сервер",
  appliance: "за шасси",
  bundle: "за комплект",
};

/** Кривая суммарного веса угроз по месяцам горизонта. */
const RiskCurve: React.FC<{ roadmap: Roadmap }> = ({ roadmap }) => {
  const months = roadmap.horizon_years * 12;
  const w = 560;
  const h = 150;
  const padL = 38;
  const padB = 22;

  // Ряд восстанавливаем по годам: внутри года риск меняется, когда
  // внедряется купленное, поэтому берём начало и конец каждого периода.
  const points: { m: number; v: number }[] = [{ m: 0, v: roadmap.baseline_w }];
  roadmap.periods.forEach((p, i) => {
    points.push({ m: i * 12, v: p.w_at_start });
    points.push({ m: (i + 1) * 12 - 1, v: p.w_at_end });
  });

  const maxW = Math.max(roadmap.baseline_w, ...points.map(p => p.v)) || 1;
  const x = (m: number) => padL + (m / Math.max(months - 1, 1)) * (w - padL - 8);
  const y = (v: number) => 8 + (1 - v / maxW) * (h - padB - 8);

  const path = points.map((p, i) => `${i === 0 ? "M" : "L"} ${x(p.m)} ${y(p.v)}`).join(" ");
  const area = `${path} L ${x(months - 1)} ${h - padB} L ${padL} ${h - padB} Z`;

  return (
    <svg viewBox={`0 0 ${w} ${h}`} style={{ width: "100%", height: "auto" }} role="img"
         aria-label="Кривая риска по месяцам горизонта планирования">
      {/* Уровень риска без вложений — то, с чем сравниваем */}
      <line x1={padL} y1={y(roadmap.baseline_w)} x2={w - 8} y2={y(roadmap.baseline_w)}
            stroke="var(--danger)" strokeDasharray="4 3" strokeWidth={1} opacity={0.6} />
      <text x={w - 10} y={y(roadmap.baseline_w) - 5} textAnchor="end"
            fontSize={9} fill="var(--danger)">без вложений</text>

      <path d={area} fill="var(--accent)" opacity={0.12} />
      <path d={path} fill="none" stroke="var(--accent)" strokeWidth={2} />

      {roadmap.periods.map((p, i) => (
        <g key={p.year}>
          <line x1={x(i * 12)} y1={8} x2={x(i * 12)} y2={h - padB}
                stroke="var(--border)" strokeWidth={1} />
          <text x={x(i * 12) + 4} y={h - 8} fontSize={9} fill="var(--fg-dim)">
            год {p.year}
          </text>
        </g>
      ))}

      <text x={4} y={y(maxW) + 4} fontSize={9} fill="var(--fg-dim)">{maxW.toFixed(2)}</text>
      <text x={4} y={h - padB} fontSize={9} fill="var(--fg-dim)">0</text>
    </svg>
  );
};

export const OptimizerPage: React.FC = () => {
  const location = useLocation();
  const [assets, setAssets] = useState<Asset[]>([]);
  const [assetId, setAssetId] = useState<number | null>(null);

  const [budget, setBudget] = useState(1_000_000);
  const [budgetPerYear, setBudgetPerYear] = useState(300_000);
  const [years, setYears] = useState(3);
  const [maxClass, setMaxClass] = useState<number | "">("");
  const [scale, setScale] = useState<AssetScale>({ workstations: 50, servers: 3, appliances: 1 });

  const [plan, setPlan] = useState<OptimizerPlan | null>(null);
  const [roadmap, setRoadmap] = useState<Roadmap | null>(null);
  const [mode, setMode] = useState<"budget" | "roadmap">("budget");
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    api.getAssets()
      .then(list => {
        setAssets(list);
        const requested = Number(new URLSearchParams(location.search).get("asset"));
        const found = list.find(a => a.id === requested);
        setAssetId(found ? found.id : list[0]?.id ?? null);
      })
      .catch(e => setErr(e.message));
  }, [location.search]);

  const run = async () => {
    if (!assetId) return;
    setLoading(true);
    setErr(null);
    try {
      const cls = maxClass === "" ? undefined : Number(maxClass);
      if (mode === "budget") {
        setPlan(await api.optimizeAsset(assetId, budget, scale, cls));
        setRoadmap(null);
      } else {
        setRoadmap(await api.getAssetRoadmap(assetId, budgetPerYear, years, scale, cls));
        setPlan(null);
      }
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setLoading(false);
    }
  };

  const reductionPct = useMemo(() => {
    if (plan) return plan.baseline_w > 0 ? (100 * plan.total_delta) / plan.baseline_w : 0;
    if (roadmap) return roadmap.baseline_area > 0 ? (100 * roadmap.area_reduction) / roadmap.baseline_area : 0;
    return 0;
  }, [plan, roadmap]);

  const warnings = plan?.warnings ?? roadmap?.warnings ?? [];
  const skipped = plan?.skipped ?? roadmap?.skipped ?? [];

  return (
    <div style={{ display: "grid", gap: 16 }}>
      <Card
        title="Подбор комплекса средств защиты"
        subtitle="Максимальное снижение риска в рамках бюджета"
        action={
          <div style={{ display: "flex", gap: 6 }}>
            <Btn size="sm" variant={mode === "budget" ? "primary" : "ghost"} onClick={() => setMode("budget")}>
              Единый бюджет
            </Btn>
            <Btn size="sm" variant={mode === "roadmap" ? "primary" : "ghost"} onClick={() => setMode("roadmap")}>
              Дорожная карта
            </Btn>
          </div>
        }
      >
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))", gap: 12 }}>
          <label style={{ display: "grid", gap: 4 }}>
            <span style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)" }}>Актив</span>
            <select value={assetId ?? ""} onChange={e => setAssetId(Number(e.target.value))}
                    style={{ padding: "6px 8px" }}>
              {assets.map(a => <option key={a.id} value={a.id}>{a.name}</option>)}
            </select>
          </label>

          {mode === "budget" ? (
            <label style={{ display: "grid", gap: 4 }}>
              <span style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)" }}>Бюджет, ₽</span>
              <input type="number" value={budget} min={0} step={50_000}
                     onChange={e => setBudget(Number(e.target.value))} style={{ padding: "6px 8px" }} />
            </label>
          ) : (
            <>
              <label style={{ display: "grid", gap: 4 }}>
                <span style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)" }}>Бюджет в год, ₽</span>
                <input type="number" value={budgetPerYear} min={0} step={50_000}
                       onChange={e => setBudgetPerYear(Number(e.target.value))} style={{ padding: "6px 8px" }} />
              </label>
              <label style={{ display: "grid", gap: 4 }}>
                <span style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)" }}>Горизонт, лет</span>
                <input type="number" value={years} min={1} max={5}
                       onChange={e => setYears(Number(e.target.value))} style={{ padding: "6px 8px" }} />
              </label>
            </>
          )}

          <label style={{ display: "grid", gap: 4 }}>
            <span style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)" }}>Класс защиты не ниже</span>
            <select value={maxClass} onChange={e => setMaxClass(e.target.value === "" ? "" : Number(e.target.value))}
                    style={{ padding: "6px 8px" }}>
              <option value="">любой</option>
              {[1, 2, 3, 4, 5, 6].map(c => <option key={c} value={c}>{c} класс</option>)}
            </select>
          </label>
        </div>

        <div style={{ marginTop: 12, paddingTop: 12, borderTop: "1px solid var(--border)" }}>
          <div style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)", marginBottom: 8 }}>
            Масштаб актива — на него умножается цена за единицу лицензирования
          </div>
          <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(130px, 1fr))", gap: 12 }}>
            {([
              ["workstations", "Рабочих мест"],
              ["servers", "Серверов"],
              ["appliances", "Шлюзов"],
            ] as const).map(([key, label]) => (
              <label key={key} style={{ display: "grid", gap: 4 }}>
                <span style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)" }}>{label}</span>
                <input type="number" min={1} value={scale[key]}
                       onChange={e => setScale({ ...scale, [key]: Math.max(1, Number(e.target.value)) })}
                       style={{ padding: "6px 8px" }} />
              </label>
            ))}
          </div>
        </div>

        <div style={{ marginTop: 14 }}>
          <Btn variant="primary" onClick={run} disabled={loading || !assetId}
               icon={<Icon name="zap" size={13} />}>
            {loading ? "Считаем…" : "Подобрать"}
          </Btn>
        </div>

        {err && (
          <div style={{ marginTop: 10, color: "var(--danger)", fontSize: "var(--text-sm)" }}>{err}</div>
        )}
      </Card>

      {(plan || roadmap) && (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))", gap: 12 }}>
          <StatCard label="Риск до" value={(plan?.baseline_w ?? roadmap?.baseline_w ?? 0).toFixed(3)} tone="critical" mono />
          <StatCard label="Риск после" value={(plan?.resulting_w ?? roadmap?.final_w ?? 0).toFixed(3)} tone="low" mono />
          <StatCard
            label={roadmap ? "Площадь под кривой" : "Снижение риска"}
            value={`−${reductionPct.toFixed(1)}%`}
            sub={roadmap ? `${roadmap.baseline_area.toFixed(2)} → ${roadmap.risk_area.toFixed(2)} W·год` : undefined}
            tone="accent"
            mono
          />
          <StatCard label="Затраты" value={money(plan?.total_cost ?? roadmap?.total_cost ?? 0)} tone="info" mono />
        </div>
      )}

      {plan && plan.exhaustive_checked && (
        <Card dense>
          <div style={{ display: "flex", alignItems: "center", gap: 8, fontSize: "var(--text-sm)" }}>
            <Chip tone={plan.greedy_is_optimal ? "success" : "warn"} mono>
              {plan.greedy_is_optimal ? "оптимум достигнут" : "жадный не оптимален"}
            </Chip>
            <span style={{ color: "var(--fg-dim)" }}>
              сверено полным перебором: лучший возможный результат −{plan.exhaustive_delta?.toFixed(4)},
              получено −{plan.total_delta.toFixed(4)}
            </span>
          </div>
        </Card>
      )}

      {roadmap && (
        <Card title="Кривая риска" subtitle="Площадь под кривой — то, что минимизирует планировщик" dense>
          <RiskCurve roadmap={roadmap} />
        </Card>
      )}

      {plan && plan.steps.length > 0 && (
        <Card title="План закупок" subtitle={`${plan.steps.length} средств, по убыванию отдачи на рубль`} dense>
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "var(--text-sm)" }}>
              <thead>
                <tr style={{ textAlign: "left", color: "var(--fg-dim)", fontSize: "var(--text-xs)" }}>
                  <th style={{ padding: "6px 8px" }}>Метод</th>
                  <th style={{ padding: "6px 8px" }}>Средство</th>
                  <th style={{ padding: "6px 8px", textAlign: "right" }}>Цена</th>
                  <th style={{ padding: "6px 8px", textAlign: "right" }}>Кол-во</th>
                  <th style={{ padding: "6px 8px", textAlign: "right" }}>Итого</th>
                  <th style={{ padding: "6px 8px", textAlign: "right" }}>ΔW</th>
                </tr>
              </thead>
              <tbody>
                {plan.steps.map((s, i) => (
                  <tr key={i} style={{ borderTop: "1px solid var(--border)" }}>
                    <td style={{ padding: "8px" }}><Chip tone="accent" mono>{s.candidate.control_code}</Chip></td>
                    <td style={{ padding: "8px" }}>
                      <div>{s.candidate.product_name}</div>
                      <div style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)" }}>
                        {s.candidate.vendor}
                        {s.candidate.protection_class ? ` · ${s.candidate.protection_class} класс` : ""}
                        {s.candidate.valid_until ? ` · до ${s.candidate.valid_until}` : " · бессрочно"}
                      </div>
                    </td>
                    <td style={{ padding: "8px", textAlign: "right" }} className="mono">
                      {money(s.candidate.cost_max)}
                      <div style={{ fontSize: "var(--text-xs)", color: "var(--fg-dim)" }}>
                        {UNIT_LABEL[s.candidate.pricing_unit] ?? s.candidate.pricing_unit}
                      </div>
                    </td>
                    <td style={{ padding: "8px", textAlign: "right" }} className="mono">×{s.candidate.units}</td>
                    <td style={{ padding: "8px", textAlign: "right", fontWeight: 600 }} className="mono">
                      {money(s.candidate.total_cost)}
                    </td>
                    <td style={{ padding: "8px", textAlign: "right", color: "var(--success)" }} className="mono">
                      −{s.delta_w.toFixed(4)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {roadmap && roadmap.periods.map(p => (
        <Card key={p.year} title={`Год ${p.year}`}
              subtitle={`${money(p.spent)} · риск ${p.w_at_start.toFixed(3)} → ${p.w_at_end.toFixed(3)}`} dense>
          {p.purchases.length === 0 ? (
            <div style={{ color: "var(--fg-muted)", fontSize: "var(--text-sm)" }}>
              Закупок нет — бюджета не хватает либо закрывать больше нечего.
            </div>
          ) : (
            <div style={{ display: "grid", gap: 8 }}>
              {p.purchases.map((buy, i) => (
                <div key={i} style={{ display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
                  <Chip tone="accent" mono>{buy.candidate.control_code}</Chip>
                  <span style={{ flex: 1, minWidth: 200 }}>{buy.candidate.product_name}</span>
                  <Chip tone="ghost">
                    внедрение {buy.deploy_months} мес · работает с {buy.active_from_month + 1}-го
                  </Chip>
                  {buy.expires_at_month !== undefined && (
                    <Chip tone="warn">сертификат истекает на {buy.expires_at_month + 1}-м мес</Chip>
                  )}
                  <span className="mono" style={{ fontWeight: 600 }}>{money(buy.cost)}</span>
                </div>
              ))}
            </div>
          )}
        </Card>
      ))}

      {warnings.length > 0 && (
        <Card title="Оговорки к расчёту" dense>
          <ul style={{ margin: 0, paddingLeft: 18, display: "grid", gap: 6, fontSize: "var(--text-sm)", color: "var(--fg-muted)" }}>
            {warnings.map((w, i) => <li key={i}>{w}</li>)}
          </ul>
        </Card>
      )}

      {skipped.length > 0 && (
        <Card title="Методы без подходящих средств"
              subtitle="Закрыть нечем: сертифицированного средства с известной ценой не нашлось" dense>
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
            {skipped.map((s, i) => (
              <Chip key={i} tone="ghost" mono>{s.candidate.control_code}</Chip>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
};
