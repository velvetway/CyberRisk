import React, { useEffect, useMemo, useState } from "react";
import { useParams, useSearchParams, useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { authFetch } from "../api/client";
import { Btn, Card, Icon } from "../components/design";
import { AssetRiskHero } from "../components/risk/AssetRiskHero";
import { ThreatList } from "../components/risk/ThreatList";
import { AttackFlowSankey } from "../components/risk/AttackFlowSankey";
import { ControlPopover } from "../components/risk/ControlPopover";
import { WhatIfBar } from "../components/risk/WhatIfBar";
import { PtsziBreakdown } from "../components/risk/PtsziBreakdown";
import { recomputeW } from "../lib/riskFlow";
import type { AssetAttackPathsResponse, AttackPath, ControlCoverage } from "../types/riskGraph";

export const RiskGraphPage: React.FC = () => {
  const { assetId } = useParams<{ assetId: string }>();
  const [params, setParams] = useSearchParams();
  const navigate = useNavigate();

  const [response, setResponse] = useState<AssetAttackPathsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [disabledControls, setDisabledControls] = useState<Set<number>>(new Set());
  const [popoverFor, setPopoverFor] = useState<{ control: ControlCoverage; anchor: { x: number; y: number } } | null>(null);

  // Fetch all attack paths for the asset
  useEffect(() => {
    if (!assetId) return;
    setLoading(true);
    setError(null);
    authFetch(`/api/risk/asset/${assetId}/attack-paths`)
      .then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`)))
      .then((data: AssetAttackPathsResponse) => {
        setResponse(data);
        setLoading(false);
      })
      .catch(e => {
        setError(e.message);
        setLoading(false);
      });
  }, [assetId]);

  // Resolve selected threat from query or auto-pick max-W
  const requestedThreatId = params.get('threat');
  const selectedThreatId = useMemo<number | null>(() => {
    if (!response || response.paths.length === 0) return null;
    if (requestedThreatId) {
      const id = parseInt(requestedThreatId, 10);
      if (response.paths.some(p => p.threat.id === id)) return id;
      return response.paths[0].threat.id;
    }
    return response.paths.reduce((acc, p) => p.w > acc.w ? p : acc, response.paths[0]).threat.id;
  }, [response, requestedThreatId]);

  // Surface invalid ?threat= deep-links via a side effect (not inside the memo).
  useEffect(() => {
    if (!response || !requestedThreatId) return;
    const id = parseInt(requestedThreatId, 10);
    if (!response.paths.some(p => p.threat.id === id)) {
      toast.error(`Угроза id=${id} не найдена для этого актива`);
    }
  }, [response, requestedThreatId]);

  // Auto-sync ?threat= without push-state. Intentionally omit `params` from
  // the deps so a search-param mutation doesn't re-fire this effect; the
  // load-bearing inputs are selectedThreatId and requestedThreatId only.
  useEffect(() => {
    if (selectedThreatId !== null && String(selectedThreatId) !== requestedThreatId) {
      const next = new URLSearchParams();
      next.set('threat', String(selectedThreatId));
      setParams(next, { replace: true });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedThreatId, requestedThreatId]);

  // Reset what-if when switching threats
  useEffect(() => {
    setDisabledControls(new Set());
    setPopoverFor(null);
  }, [selectedThreatId]);

  const selectedPath: AttackPath | null = useMemo(() => {
    if (!response || selectedThreatId === null) return null;
    return response.paths.find(p => p.threat.id === selectedThreatId) ?? null;
  }, [response, selectedThreatId]);

  const simulation = useMemo(() => {
    if (!selectedPath) return null;
    if (disabledControls.size === 0) return null;
    return recomputeW(selectedPath, disabledControls);
  }, [selectedPath, disabledControls]);

  const allControlsForPath = useMemo(() => {
    if (!selectedPath) return new Map<number, ControlCoverage>();
    const m = new Map<number, ControlCoverage>();
    for (const vl of selectedPath.vulnerable_links) {
      for (const c of vl.coverage_controls) {
        if (!m.has(c.id)) m.set(c.id, c);
      }
    }
    return m;
  }, [selectedPath]);

  const handlePdf = async () => {
    if (!response || selectedThreatId === null) return;
    try {
      const res = await authFetch('/api/risk/report/pdf', {
        method: 'POST',
        body: JSON.stringify({ asset_id: response.asset.id, threat_id: selectedThreatId }),
      });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const blob = await res.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `risk_${response.asset.id}_${selectedThreatId}.pdf`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (e) {
      toast.error(`PDF: ${(e as Error).message}`);
    }
  };

  // ─────────────────── Render states ───────────────────

  if (!assetId) {
    return (
      <div style={{ padding: 40, textAlign: 'center', color: 'var(--fg-muted)' }}>
        Не указан актив.
      </div>
    );
  }

  if (loading) {
    return (
      <div style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ height: 80, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-md)' }} />
        <div style={{ height: 200, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-md)' }} />
        <div style={{ height: 460, background: 'var(--bg-elev-2)', borderRadius: 'var(--r-md)' }} />
      </div>
    );
  }

  if (error || !response) {
    return (
      <div style={{ padding: 40 }}>
        <Card title="Ошибка загрузки" dense>
          <div style={{ color: 'var(--risk-critical)' }}>{error ?? 'Нет данных'}</div>
          <Btn style={{ marginTop: 14 }} variant="outline" onClick={() => navigate(-1)} icon={<Icon name="arrowL" size={13} />}>
            Назад
          </Btn>
        </Card>
      </div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.3 }}
      style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}
    >
      <AssetRiskHero
        assetId={response.asset.id}
        assetName={response.asset.name}
        aggregate={response.aggregate}
        onBack={() => navigate(-1)}
        onPdf={selectedThreatId !== null ? handlePdf : undefined}
      />

      <ThreatList
        paths={response.paths}
        selectedThreatId={selectedThreatId}
        onSelect={(id) => {
          const next = new URLSearchParams(params);
          next.set('threat', String(id));
          setParams(next, { replace: true });
        }}
      />

      {selectedPath && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 340px', gap: 16, alignItems: 'start' }}>
          <Card pad={0}>
            <AttackFlowSankey
              path={selectedPath}
              disabledControlIds={disabledControls}
              onControlClick={(id, anchor) => {
                const c = allControlsForPath.get(id);
                if (c) setPopoverFor({ control: c, anchor });
              }}
            />
          </Card>

          <PtsziBreakdown
            path={selectedPath}
            simulated={simulation ?? undefined}
          />
        </div>
      )}

      {selectedPath && simulation && (
        <WhatIfBar
          disabledChips={Array.from(disabledControls).map(id => ({
            id,
            name: allControlsForPath.get(id)?.name ?? `C-${id}`,
          }))}
          baselineW={selectedPath.w}
          simulatedW={simulation.w}
          delta={simulation.delta}
          onReset={() => setDisabledControls(new Set())}
          onRemoveChip={(id) => setDisabledControls(s => {
            const next = new Set(s); next.delete(id); return next;
          })}
          onSaveNote={() => {
            const note = {
              assetId: response.asset.id,
              threatId: selectedPath.threat.id,
              disabledIds: Array.from(disabledControls),
              w: simulation.w,
              wBaseline: selectedPath.w,
              ts: Date.now(),
            };
            const key = 'risk:notes';
            const prev = JSON.parse(localStorage.getItem(key) ?? '[]');
            localStorage.setItem(key, JSON.stringify([note, ...prev]));
            toast.success('Заметка сохранена локально');
          }}
        />
      )}

      {popoverFor && (
        <ControlPopover
          control={popoverFor.control}
          disabled={disabledControls.has(popoverFor.control.id)}
          anchor={popoverFor.anchor}
          onToggle={(id) => {
            setDisabledControls(s => {
              const next = new Set(s);
              if (next.has(id)) next.delete(id); else next.add(id);
              return next;
            });
            setPopoverFor(null);
          }}
          onClose={() => setPopoverFor(null)}
        />
      )}
    </motion.div>
  );
};
