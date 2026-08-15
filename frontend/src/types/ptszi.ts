import { Asset } from "../types";

export interface ThreatSource {
  id: number;
  code: string;
  name: string;
  description?: string;
}

export interface PtsziThreat {
  id: number;
  code: string;
  name: string;
  description?: string;
  q_threat: number;
  q_severity: number;
  contours?: string[];
}

export interface PtsziVulnerableLink {
  id: number;
  code: string;
  name: string;
  description?: string;
}

export interface PtsziControl {
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
}

export interface PtsziUBIThreat {
  id: number;
  ubi_code: string;
  ubi_number: number;
  name: string;
  description?: string;
  source_raw?: string;
  impact_object?: string;
  impact_confidentiality: boolean;
  impact_integrity: boolean;
  impact_availability: boolean;
  max_potential: string;
  q_threat: number;
  q_severity: number;
  mapped_sources?: string[];
}

export interface PtsziControlCoverage {
  control: PtsziControl;
  coverage: number;
  implemented: boolean;
  effectiveness: number;
  resulting_coverage: number;
}

export interface PtsziPathVL {
  vulnerable_link: PtsziVulnerableLink;
  status: string;
  comment?: string;
  coverage: number;
  uncovered: boolean;
  controls: PtsziControlCoverage[];
}

export interface PtsziRecommendation {
  control_id: number;
  control_code: string;
  category: string;
  title: string;
  description: string;
  priority: string;
}

export interface PtsziAttackPath {
  asset: { id: number; name: string };
  asset_contour: string;
  threat: PtsziThreat;
  sources: ThreatSource[];
  vulnerable_links: PtsziPathVL[];
  destructive_actions: DestructiveAction[];
  ubi: PtsziUBIThreat[];
  q_threat: number;
  q_severity: number;
  q_reaction: number;
  z: number;
  w: number;
  level: "critical" | "high" | "medium" | "low";
  applicable: boolean;
  missing_controls: PtsziControl[];
  recommendations?: PtsziRecommendation[];
}

export interface PtsziAssetProfile {
  asset: Asset;
  security_contour: string;
  vulnerable_links: Array<{
    vulnerable_link: PtsziVulnerableLink;
    status: string;
    comment?: string;
  }>;
  controls: Array<{
    control: PtsziControl;
    effectiveness: number;
    comment?: string;
  }>;
  applicable_threats: PtsziAttackPath[];
}
