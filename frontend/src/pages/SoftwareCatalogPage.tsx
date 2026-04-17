import React, { useEffect, useMemo, useState } from "react";
import { motion } from "framer-motion";
import { authFetch } from "../api/client";
import { Software, SoftwareCategory } from "../types";
import { Btn, Chip, Icon, StatCard } from "../components/design";

type FilterRu = 'all' | 'ru' | 'foreign';

export const SoftwareCatalogPage: React.FC = () => {
  const [software, setSoftware] = useState<Software[]>([]);
  const [categories, setCategories] = useState<SoftwareCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [filterRu, setFilterRu] = useState<FilterRu>('all');
  const [filterCert, setFilterCert] = useState(false);

  useEffect(() => {
    Promise.all([
      authFetch('/api/software').then(r => r.ok ? r.json() : Promise.reject(new Error(`HTTP ${r.status}`))),
      authFetch('/api/software/categories').then(r => r.ok ? r.json() : []),
    ]).then(([s, c]) => {
      setSoftware(Array.isArray(s) ? s : []);
      setCategories(Array.isArray(c) ? c : []);
      setLoading(false);
    }).catch(e => { setErr(e.message); setLoading(false); });
  }, []);

  const filtered = useMemo(() => {
    return software.filter(s => {
      if (filterRu === 'ru' && !s.is_russian) return false;
      if (filterRu === 'foreign' && s.is_russian) return false;
      if (filterCert && !s.fstec_certified && !s.fsb_certified) return false;
      if (search.trim()) {
        const q = search.toLowerCase();
        const match = (s.name ?? '').toLowerCase().includes(q)
          || (s.vendor ?? '').toLowerCase().includes(q)
          || (s.registry_number ?? '').toLowerCase().includes(q);
        if (!match) return false;
      }
      return true;
    });
  }, [software, filterRu, filterCert, search]);

  const russianCount = software.filter(s => s.is_russian).length;
  const fstecCount = software.filter(s => s.fstec_certified).length;
  const fsbCount = software.filter(s => s.fsb_certified).length;
  const foreignCount = software.length - russianCount;

  return (
    <motion.div
      initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.3 }}
      style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}
    >
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 16 }}>
        <div>
          <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 4 }}>Справочник ПО · реестр Минцифры</div>
          <h1 style={{ margin: 0, fontSize: 'var(--text-2xl)', fontWeight: 600, letterSpacing: '-0.02em' }}>Российское ПО и СЗИ</h1>
          <div style={{ fontSize: 'var(--text-sm)', color: 'var(--fg-muted)', marginTop: 4 }}>
            {software.length} позиций · {russianCount} в реестре Минцифры · {fstecCount} с ФСТЭК-сертификатами
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <Btn variant="outline" icon={<Icon name="refresh" size={13} />}>Синхронизация реестра</Btn>
          <Btn variant="primary" icon={<Icon name="plus" size={13} />}>Добавить ПО</Btn>
        </div>
      </div>

      {err && (
        <div style={{ padding: 14, background: 'var(--risk-critical-bg)', border: '1px solid var(--risk-critical-br)', borderRadius: 'var(--r-md)', color: 'var(--risk-critical)', fontSize: 'var(--text-sm)' }}>⚠ {err}</div>
      )}

      {loading && !err && (
        <div style={{ textAlign: 'center', padding: 60, color: 'var(--fg-dim)' }}>Загрузка…</div>
      )}

      {!loading && !err && (
        <>
          {/* Stats */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12 }}>
            <StatCard label="В реестре Минцифры" value={russianCount} tone="low" mono sub={`${software.length > 0 ? Math.round(russianCount / software.length * 100) : 0}% каталога`} />
            <StatCard label="ФСТЭК-сертификаты" value={fstecCount} tone="accent" mono sub="УД-1 — УД-6, СКЗИ, МЭ, СОВ" />
            <StatCard label="ФСБ-сертификаты" value={fsbCount} tone="info" mono sub="КС1 — КВ2, СКЗИ" />
            <StatCard label="Зарубежное ПО" value={foreignCount} tone="critical" mono sub="Требуют замены" />
          </div>

          {/* Filters */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '8px 12px', background: 'var(--bg-elev-1)', border: '1px solid var(--border)', borderRadius: 'var(--r-md)' }}>
            <Icon name="search" size={14} color="var(--fg-dim)" />
            <input placeholder="Поиск: название, вендор, номер реестра..." value={search} onChange={e => setSearch(e.target.value)} style={{ flex: 1, background: 'transparent', border: 'none', outline: 'none', fontSize: 'var(--text-sm)', color: 'var(--fg)' }} />
            <div style={{ width: 1, height: 20, background: 'var(--border)' }} />
            {([{ v: 'all', l: 'Все' }, { v: 'ru', l: '🇷🇺 Реестр' }, { v: 'foreign', l: 'Зарубежное' }] as const).map(o => (
              <button key={o.v} onClick={() => setFilterRu(o.v as FilterRu)} style={{
                height: 24, padding: '0 10px',
                background: filterRu === o.v ? 'var(--bg-active)' : 'transparent',
                color: filterRu === o.v ? 'var(--fg)' : 'var(--fg-muted)',
                border: `1px solid ${filterRu === o.v ? 'var(--border-strong)' : 'transparent'}`,
                borderRadius: 'var(--r-xs)', fontSize: 'var(--text-xs)', cursor: 'pointer',
              }}>{o.l}</button>
            ))}
            <button onClick={() => setFilterCert(!filterCert)} style={{
              height: 24, padding: '0 10px',
              background: filterCert ? 'var(--accent-ghost)' : 'transparent',
              color: filterCert ? 'var(--accent)' : 'var(--fg-muted)',
              border: 'none', borderRadius: 'var(--r-xs)', fontSize: 'var(--text-xs)', cursor: 'pointer',
              display: 'inline-flex', alignItems: 'center', gap: 4,
            }}><Icon name="award" size={11} />Только сертифицированные</button>
          </div>

          {/* Cards */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(330px, 1fr))', gap: 12 }}>
            {filtered.map(s => {
              const categoryName = s.category_name ?? categories.find(c => c.id === s.category_id)?.name ?? '—';
              return (
                <div key={s.id} style={{
                  padding: 16, background: 'var(--bg-elev-1)', border: '1px solid var(--border)',
                  borderRadius: 'var(--r-lg)', cursor: 'pointer',
                  transition: 'var(--transition)',
                  display: 'flex', flexDirection: 'column', gap: 10
                }}
                  onMouseEnter={e => { e.currentTarget.style.borderColor = 'var(--border-strong)'; e.currentTarget.style.transform = 'translateY(-1px)'; }}
                  onMouseLeave={e => { e.currentTarget.style.borderColor = 'var(--border)'; e.currentTarget.style.transform = 'translateY(0)'; }}
                >
                  <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 10 }}>
                    <div>
                      <div style={{ fontSize: 'var(--text-md)', fontWeight: 500, marginBottom: 2 }}>{s.name}</div>
                      <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-dim)' }}>{s.vendor}</div>
                    </div>
                    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 4 }}>
                      {s.is_russian ? <Chip tone="success" mono>РЕЕСТР</Chip> : <Chip tone="danger" mono>FOREIGN</Chip>}
                    </div>
                  </div>
                  <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
                    <Chip tone="ghost">{categoryName}</Chip>
                    {s.fstec_certified && <Chip tone="accent" mono>ФСТЭК{s.fstec_protection_class ? ` ${s.fstec_protection_class}` : ''}</Chip>}
                    {s.fsb_certified && <Chip tone="accent" mono>ФСБ{s.fsb_protection_class ? ` ${s.fsb_protection_class}` : ''}</Chip>}
                  </div>
                  {s.is_russian && s.registry_number && (
                    <div style={{ fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', fontFamily: 'var(--font-mono)' }}>
                      Минцифры: №{s.registry_number}
                    </div>
                  )}
                  {s.description && (
                    <div style={{ paddingTop: 8, borderTop: '1px dashed var(--border)', fontSize: 'var(--text-xs)', color: 'var(--fg-muted)', lineHeight: 1.4 }}>
                      {s.description.length > 120 ? s.description.slice(0, 119) + '…' : s.description}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
          {filtered.length === 0 && (
            <div style={{ textAlign: 'center', padding: 40, color: 'var(--fg-dim)' }}>Ничего не найдено по заданным фильтрам</div>
          )}
        </>
      )}
    </motion.div>
  );
};
