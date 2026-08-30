import type { AttackPath, ControlCoverage } from '../types/riskGraph';

export type SankeyNodeKind = 'S' | 'ST' | 'VL' | 'C' | 'DA';
export type SankeyLinkKind = 'S->ST' | 'ST->VL' | 'VL->C' | 'VL->DA';

export interface SankeyNodeData {
  id: string;
  kind: SankeyNodeKind;
  label: string;
  meta?: {
    coverage?: number;
    disabled?: boolean;
    uncovered?: boolean;
    code?: string;
  };
}

export interface SankeyLinkData {
  source: string;
  target: string;
  value: number;
  kind: SankeyLinkKind;
}

export interface SankeyGraph {
  nodes: SankeyNodeData[];
  links: SankeyLinkData[];
}

export interface RecomputeResult {
  qReaction: number;
  w: number;
  delta: number; // w − path.w; положительная дельта = риск вырос
}

const clamp01 = (v: number): number => (v < 0 ? 0 : v > 1 ? 1 : v);
const clampZ = (z: number): number => (z < 0.5 ? 0.5 : z > 1 ? 1 : z);

/**
 * Строит Sankey-граф S → ST → VL → {C | DA} с честной flow-семантикой.
 * Поток нормирован на 1.0. Подробное описание формул — в spec'е.
 */
export function buildSankeyGraph(
  path: AttackPath,
  disabledControlIds: Set<number> = new Set(),
): SankeyGraph {
  const nodes: SankeyNodeData[] = [];
  const links: SankeyLinkData[] = [];

  // S nodes
  for (const s of path.sources) {
    nodes.push({ id: `S${s.id}`, kind: 'S', label: s.code, meta: { code: s.code } });
  }

  // ST node (single)
  nodes.push({
    id: 'ST',
    kind: 'ST',
    label: path.threat.name,
    meta: { code: path.threat.bdu_id },
  });

  // VL nodes (категории VL1..VL6 из диплома)
  for (const vl of path.vulnerable_links) {
    nodes.push({
      id: `VL${vl.category_id}`,
      kind: 'VL',
      label: vl.name,
      meta: { uncovered: vl.uncovered, code: vl.code },
    });
  }

  // C nodes — дедуп по id (один контроль может покрывать несколько VL)
  const allControls = new Map<number, ControlCoverage>();
  for (const vl of path.vulnerable_links) {
    for (const c of vl.coverage_controls) {
      if (!allControls.has(c.id)) allControls.set(c.id, c);
    }
  }
  allControls.forEach((c) => {
    nodes.push({
      id: `C${c.id}`,
      kind: 'C',
      label: c.name,
      meta: { coverage: c.coverage, disabled: disabledControlIds.has(c.id) },
    });
  });

  // DA nodes
  for (const da of path.destructive_actions) {
    nodes.push({ id: `DA${da.id}`, kind: 'DA', label: da.name, meta: { code: da.code } });
  }

  // S → ST: равномерное распределение
  if (path.sources.length > 0) {
    const w = 1 / path.sources.length;
    for (const s of path.sources) {
      links.push({ source: `S${s.id}`, target: 'ST', value: w, kind: 'S->ST' });
    }
  }

  // ST → VL: равномерное распределение по VL-категориям. У категорий нет
  // «severity» — они одинаково применимы; вес угрозы распределяется поровну.
  const vlCount = path.vulnerable_links.length;
  if (vlCount > 0) {
    const perVL = 1 / vlCount;
    for (const vl of path.vulnerable_links) {
      links.push({
        source: 'ST',
        target: `VL${vl.category_id}`,
        value: perVL,
        kind: 'ST->VL',
      });
    }
  }

  // Per-VL: VL → C (только enabled), VL → DA (passthrough)
  for (const vl of path.vulnerable_links) {
    const inflow = vlCount > 0 ? 1 / vlCount : 0;
    const enabled = vl.coverage_controls.filter(c => !disabledControlIds.has(c.id));
    const sumC = enabled.reduce((a, c) => a + c.coverage, 0);
    const cov = Math.min(1, sumC);

    if (sumC > 0) {
      for (const c of enabled) {
        const share = c.coverage / sumC;
        const value = share * cov * inflow;
        if (value > 0) {
          links.push({
            source: `VL${vl.category_id}`,
            target: `C${c.id}`,
            value,
            kind: 'VL->C',
          });
        }
      }
    }

    const passthrough = inflow * (1 - cov);
    if (passthrough > 0 && path.destructive_actions.length > 0) {
      const perDa = passthrough / path.destructive_actions.length;
      for (const da of path.destructive_actions) {
        links.push({
          source: `VL${vl.category_id}`,
          target: `DA${da.id}`,
          value: perDa,
          kind: 'VL->DA',
        });
      }
    }
  }

  return { nodes, links };
}

/**
 * Пересчитывает W при отключении набора контролей. Зеркалит бэкенд-формулу
 * QReactionFromVLs: VL считается «закрытой», если у неё есть хотя бы один
 * включённый контроль с coverage > 0.
 */
export function recomputeW(
  path: AttackPath,
  disabledControlIds: Set<number>,
): RecomputeResult {
  const vls = path.vulnerable_links;
  let qReaction = 0;
  if (vls.length > 0) {
    let covered = 0;
    for (const vl of vls) {
      const hasActive = vl.coverage_controls.some(
        c => c.coverage > 0 && !disabledControlIds.has(c.id),
      );
      if (hasActive) covered++;
    }
    qReaction = covered / vls.length;
  }

  const qT = clamp01(path.q_threat);
  const qS = clamp01(path.q_severity);
  const qR = clamp01(qReaction);
  const z = clampZ(path.z);
  const w = ((qT + qS + (1 - qR)) / 3) * z;

  return {
    qReaction,
    w,
    delta: w - path.w,
  };
}
