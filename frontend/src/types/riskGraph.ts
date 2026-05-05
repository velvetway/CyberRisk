export interface ThreatSource {
  id: number;
  code: string;
  name: string;
  description?: string;
}

export interface DestructiveAction {
  id: number;
  code: string;
  name: string;
  affects_confidentiality: boolean;
  affects_integrity: boolean;
  affects_availability: boolean;
  description?: string;
}

export interface ControlCoverage {
  id: number;
  name: string;
  coverage: number;
}

export interface VLNode {
  category_id: number;
  code: string; // VL1..VL6
  name: string;
  description?: string;
  coverage_controls: ControlCoverage[];
  uncovered: boolean;
  presence_count?: number; // CVE/БДУ-записи этой категории, открытые на активе
}

export interface AttackPath {
  asset: { id: number; name: string };
  threat: { id: number; name: string; bdu_id?: string };
  sources: ThreatSource[];
  vulnerable_links: VLNode[];
  destructive_actions: DestructiveAction[];
  w: number;
  q_threat: number;
  q_severity: number;
  q_reaction: number;
  z: number;
  level: 'low' | 'medium' | 'high' | 'critical';
}

export interface AssetAggregate {
  w_max: number;
  level: 'info' | 'low' | 'medium' | 'high' | 'critical';
  threat_count: number;
  uncovered_count: number;
}

export interface AssetAttackPathsResponse {
  asset: { id: number; name: string };
  aggregate: AssetAggregate;
  paths: AttackPath[];
}
