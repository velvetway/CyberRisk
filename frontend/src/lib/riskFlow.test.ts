import { buildSankeyGraph, recomputeW } from './riskFlow';
import type { AttackPath } from '../types/riskGraph';

const minimalPath = (overrides: Partial<AttackPath> = {}): AttackPath => ({
  asset: { id: 1, name: 'A' },
  threat: { id: 10, name: 'T' },
  sources: [{ id: 1, code: 'S1', name: 'Внешний нарушитель' }],
  vulnerable_links: [],
  destructive_actions: [{
    id: 1, code: 'DA1', name: 'Утечка',
    affects_confidentiality: true, affects_integrity: false, affects_availability: false,
  }],
  w: 0.5,
  q_threat: 0.5,
  q_severity: 0.5,
  q_reaction: 0,
  z: 1.0,
  level: 'high',
  ...overrides,
});

describe('buildSankeyGraph', () => {
  test('one VL with one half-coverage control: per-VL flow conserves', () => {
    const path = minimalPath({
      vulnerable_links: [{
        vulnerability_id: 1, name: 'CVE-1', severity: 5,
        coverage_controls: [{ id: 100, name: 'WAF', coverage: 0.5 }],
        uncovered: false,
      }],
    });
    const g = buildSankeyGraph(path);
    const vlIn = g.links.filter(l => l.target === 'VL1').reduce((a, l) => a + l.value, 0);
    const vlOut = g.links.filter(l => l.source === 'VL1').reduce((a, l) => a + l.value, 0);
    expect(Math.abs(vlIn - vlOut)).toBeLessThan(1e-9);
  });

  test('two controls with combined coverage > 1 are clamped, all flow blocked', () => {
    const path = minimalPath({
      vulnerable_links: [{
        vulnerability_id: 1, name: 'CVE', severity: 1,
        coverage_controls: [
          { id: 1, name: 'A', coverage: 0.7 },
          { id: 2, name: 'B', coverage: 0.6 },
        ],
        uncovered: false,
      }],
    });
    const g = buildSankeyGraph(path);
    const cFlow = g.links.filter(l => l.kind === 'VL->C').reduce((a, l) => a + l.value, 0);
    const daFlow = g.links.filter(l => l.kind === 'VL->DA').reduce((a, l) => a + l.value, 0);
    expect(cFlow + daFlow).toBeCloseTo(1, 9);
    expect(daFlow).toBeCloseTo(0, 9);
  });

  test('VL with no controls — all flow goes to DA', () => {
    const path = minimalPath({
      vulnerable_links: [{
        vulnerability_id: 1, name: 'CVE', severity: 1,
        coverage_controls: [], uncovered: true,
      }],
    });
    const g = buildSankeyGraph(path);
    const cFlow = g.links.filter(l => l.kind === 'VL->C').length;
    const daFlow = g.links.filter(l => l.kind === 'VL->DA').reduce((a, l) => a + l.value, 0);
    expect(cFlow).toBe(0);
    expect(daFlow).toBeCloseTo(1, 9);
  });

  test('disabled control does not produce VL->C edges; flow shifts to DA', () => {
    const path = minimalPath({
      vulnerable_links: [{
        vulnerability_id: 1, name: 'CVE', severity: 1,
        coverage_controls: [{ id: 1, name: 'A', coverage: 1.0 }],
        uncovered: false,
      }],
    });
    const g = buildSankeyGraph(path, new Set([1]));
    expect(g.links.filter(l => l.kind === 'VL->C').length).toBe(0);
    const daFlow = g.links.filter(l => l.kind === 'VL->DA').reduce((a, l) => a + l.value, 0);
    expect(daFlow).toBeCloseTo(1, 9);
  });

  test('S->ST equal split when multiple sources', () => {
    const path = minimalPath({
      sources: [
        { id: 1, code: 'S1', name: 'Внешний' },
        { id: 2, code: 'S2', name: 'Внутр.' },
      ],
    });
    const g = buildSankeyGraph(path);
    const stIn = g.links.filter(l => l.target === 'ST').map(l => l.value);
    expect(stIn).toHaveLength(2);
    expect(stIn[0]).toBeCloseTo(0.5, 9);
    expect(stIn[1]).toBeCloseTo(0.5, 9);
  });

  test('ST->VL split by relative severity sums to 1', () => {
    const path = minimalPath({
      vulnerable_links: [
        { vulnerability_id: 1, name: 'V1', severity: 8, coverage_controls: [], uncovered: true },
        { vulnerability_id: 2, name: 'V2', severity: 2, coverage_controls: [], uncovered: true },
      ],
    });
    const g = buildSankeyGraph(path);
    const total = g.links.filter(l => l.kind === 'ST->VL').reduce((a, l) => a + l.value, 0);
    expect(total).toBeCloseTo(1, 9);
    const v1 = g.links.find(l => l.target === 'VL1')!;
    expect(v1.value).toBeCloseTo(0.8, 9);
  });
});

describe('recomputeW', () => {
  test('baseline (no disabled): W matches formula', () => {
    const path = minimalPath({
      q_threat: 0.6, q_severity: 0.4, q_reaction: 1.0, z: 1.0,
      w: (0.6 + 0.4 + 0) / 3,
      vulnerable_links: [{
        vulnerability_id: 1, name: 'V', severity: 1,
        coverage_controls: [{ id: 1, name: 'A', coverage: 1.0 }],
        uncovered: false,
      }],
    });
    const r = recomputeW(path, new Set());
    expect(r.qReaction).toBeCloseTo(1.0, 9);
    expect(r.w).toBeCloseTo((0.6 + 0.4 + 0) / 3, 9);
    expect(Math.abs(r.delta)).toBeLessThan(1e-9);
  });

  test('disabling the only covering control drops q_reaction to 0', () => {
    const path = minimalPath({
      q_threat: 0.6, q_severity: 0.4, q_reaction: 1.0, z: 1.0, w: (0.6 + 0.4) / 3,
      vulnerable_links: [{
        vulnerability_id: 1, name: 'V', severity: 1,
        coverage_controls: [{ id: 1, name: 'A', coverage: 1.0 }],
        uncovered: false,
      }],
    });
    const r = recomputeW(path, new Set([1]));
    expect(r.qReaction).toBe(0);
    expect(r.w).toBeCloseTo((0.6 + 0.4 + 1.0) / 3, 9);
    expect(r.delta).toBeGreaterThan(0);
  });

  test('disabling one of two controls on same VL keeps it covered', () => {
    const path = minimalPath({
      q_threat: 0.5, q_severity: 0.5, q_reaction: 1.0, z: 1.0, w: (0.5 + 0.5) / 3,
      vulnerable_links: [{
        vulnerability_id: 1, name: 'V', severity: 1,
        coverage_controls: [
          { id: 1, name: 'A', coverage: 0.4 },
          { id: 2, name: 'B', coverage: 0.6 },
        ],
        uncovered: false,
      }],
    });
    const r = recomputeW(path, new Set([1])); // disable A; B (cov=0.6>0) still covers
    expect(r.qReaction).toBe(1.0);
    expect(Math.abs(r.delta)).toBeLessThan(1e-9);
  });

  test('empty VLs: q_reaction=0, W is full pass-through', () => {
    const path = minimalPath({
      q_threat: 0.4, q_severity: 0.3, q_reaction: 0, z: 1.0,
      vulnerable_links: [],
      w: (0.4 + 0.3 + 1) / 3,
    });
    const r = recomputeW(path, new Set());
    expect(r.qReaction).toBe(0);
    expect(r.w).toBeCloseTo((0.4 + 0.3 + 1) / 3, 9);
  });
});
