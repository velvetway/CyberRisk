import React from "react";
import { motion } from "framer-motion";
import { Btn, Card, Icon, RiskBadge } from "../design";
import type { AssetAggregate } from "../../types/riskGraph";

type Level = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface AssetRiskHeroProps {
  assetId: number;
  assetName: string;
  aggregate: AssetAggregate;
  onPdf?: () => void;
  onParams?: () => void;
  onBack?: () => void;
}

export const AssetRiskHero: React.FC<AssetRiskHeroProps> = ({
  assetId, assetName, aggregate, onPdf, onParams, onBack,
}) => {
  const level = aggregate.level as Level;

  return (
    <motion.div
      initial={{ opacity: 0, y: -4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <Card pad={0}>
        <div style={{
          padding: '16px 20px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: 16,
        }}>
          <div>
            <div style={{
              fontSize: 'var(--text-xs)', color: 'var(--fg-dim)',
              fontFamily: 'var(--font-mono)',
              textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4,
            }}>
              Граф атаки · ПТСЗИ
            </div>
            <h1 style={{
              margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600,
              letterSpacing: '-0.02em',
            }}>
              <span className="mono" style={{ color: 'var(--fg-muted)' }}>
                A-{String(assetId).padStart(3, '0')}
              </span>{' '}
              «{assetName}»
            </h1>
            <div style={{
              marginTop: 8, display: 'flex', gap: 16, alignItems: 'center',
              fontSize: 'var(--text-sm)', color: 'var(--fg-muted)',
            }}>
              <span className="num" style={{ color: 'var(--fg)', fontWeight: 600 }}>
                W = {aggregate.w_max.toFixed(2)}
              </span>
              <RiskBadge level={level} />
              <span>{aggregate.threat_count} угроз</span>
              <span style={{ color: aggregate.uncovered_count > 0 ? 'var(--risk-critical)' : 'var(--fg-muted)' }}>
                {aggregate.uncovered_count} непокрыто
              </span>
            </div>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            {onBack && (
              <Btn variant="outline" icon={<Icon name="arrowL" size={13} />} onClick={onBack}>
                Назад
              </Btn>
            )}
            {onParams && (
              <Btn variant="outline" icon={<Icon name="sliders" size={13} />} onClick={onParams}>
                Параметры
              </Btn>
            )}
            {onPdf && (
              <Btn variant="primary" icon={<Icon name="file" size={13} />} onClick={onPdf}>
                PDF сценария
              </Btn>
            )}
          </div>
        </div>
      </Card>
    </motion.div>
  );
};
