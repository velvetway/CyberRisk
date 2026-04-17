import React, { useState } from "react";
import { motion } from "framer-motion";
import { Card, PtsziFormula, RiskBadge } from "../components/design";

type Level = 'critical' | 'high' | 'medium' | 'low';

function levelFromW(w: number): Level {
  if (w >= 0.75) return 'critical';
  if (w >= 0.50) return 'high';
  if (w >= 0.25) return 'medium';
  return 'low';
}

const Slider: React.FC<{
  value: number;
  onChange: (v: number) => void;
  label: string;
  help: string;
  color: string;
}> = ({ value, onChange, label, help, color }) => (
  <div>
    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6, alignItems: 'baseline' }}>
      <span style={{ fontSize: 'var(--text-sm)', color: 'var(--fg)' }}>{label}</span>
      <span className="num" style={{ fontSize: 'var(--text-md)', color, fontWeight: 600 }}>{value.toFixed(2)}</span>
    </div>
    <input
      type="range" min="0" max="1" step="0.01"
      value={value}
      onChange={e => onChange(parseFloat(e.target.value))}
      style={{ width: '100%', accentColor: color }}
    />
    <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', marginTop: 4 }}>{help}</div>
  </div>
);

export const RiskPreviewPage: React.FC = () => {
  const [qTh, setQTh] = useState(0.8);
  const [qVu, setQVu] = useState(0.9);
  const [qRe, setQRe] = useState(0.2);
  const [z, setZ] = useState(0.95);

  const W = ((qTh + qVu + (1 - qRe)) / 3) * z;
  const level = levelFromW(W);
  const score = (W * 25).toFixed(1);

  return (
    <motion.div
      initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}
      style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div>
          <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4 }}>Симулятор риска · «что если»</div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600, letterSpacing: '-0.02em' }}>Интерактивная модель ПТСЗИ</h1>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-muted)', marginTop: 4 }}>
            Подвигайте ползунки чтобы смоделировать вес сценария
          </div>
        </div>
        <PtsziFormula size="lg" align="left" />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <Card title="Входные параметры" dense>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
            <Slider
              value={qTh} onChange={setQTh}
              label="Q^th · потенциал угрозы"
              help="Опасность и вектор по БДУ"
              color="var(--risk-critical)"
            />
            <Slider
              value={qVu} onChange={setQVu}
              label="q · вес уязвимости"
              help="CVSS / эксплуатируемость"
              color="var(--risk-high)"
            />
            <Slider
              value={qRe} onChange={setQRe}
              label="Q^re · зрелость реакции СЗИ"
              help="Выше = лучше защита (в формуле 1−Q^re)"
              color="var(--risk-info)"
            />
            <Slider
              value={z} onChange={setZ}
              label="Z · вес действия"
              help="Контур актива: 0.5 (dev/isolated) → 1.0 (prod)"
              color="var(--risk-medium)"
            />
          </div>
        </Card>

        <Card title="Результат" dense>
          <div style={{
            padding: 24, background: `var(--risk-${level}-bg)`,
            border: `1px solid var(--risk-${level}-br)`,
            borderRadius: 'var(--r-md)', textAlign: 'center', marginBottom: 16
          }}>
            <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 10 }}>Вес сценария W</div>
            <div className="num" style={{ fontSize: 56, fontWeight: 600, color: `var(--risk-${level})`, letterSpacing: '-0.03em', lineHeight: 1 }}>
              {W.toFixed(3)}
            </div>
            <div style={{ marginTop: 14 }}>
              <RiskBadge level={level} score={parseFloat(score)} />
            </div>
            <div className="num" style={{ fontSize: 11, color: 'var(--fg-dim)', marginTop: 14 }}>
              Адаптированный score: {score} / 25
            </div>
          </div>

          <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', textTransform: 'uppercase', letterSpacing: '0.06em', fontWeight: 500, marginBottom: 8 }}>Пошаговый расчёт</div>
          <div style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', lineHeight: 1.7, background: 'var(--bg-elev-2)', padding: '10px 12px', borderRadius: 'var(--r-sm)' }}>
            <div>step 1: (1 − Q^re) = 1 − {qRe.toFixed(2)} = <span style={{ color: 'var(--fg)' }}>{(1 - qRe).toFixed(2)}</span></div>
            <div>step 2: Q^th + q + (1−Q^re) = {qTh.toFixed(2)} + {qVu.toFixed(2)} + {(1 - qRe).toFixed(2)} = <span style={{ color: 'var(--fg)' }}>{(qTh + qVu + (1 - qRe)).toFixed(2)}</span></div>
            <div>step 3: / 3 = <span style={{ color: 'var(--fg)' }}>{((qTh + qVu + (1 - qRe)) / 3).toFixed(3)}</span></div>
            <div>step 4: × Z = × {z.toFixed(2)} = <span style={{ color: 'var(--accent)', fontWeight: 600 }}>{W.toFixed(3)}</span></div>
          </div>
        </Card>
      </div>
    </motion.div>
  );
};
