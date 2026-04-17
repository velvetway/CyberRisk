import React from 'react';
import {
  Shield, LayoutGrid, Layers, Target, Share2, Package, Bell, Search, User,
  Settings, Plus, Filter, Download, Upload, ChevronRight, ChevronDown, ChevronUp,
  X, Check, AlertTriangle, BookOpen, Zap, Server, Database, Cloud, Sun, Moon,
  Eye, Trash2, Edit, SlidersHorizontal, Menu, Command, ExternalLink, ArrowRight,
  ArrowLeft, FileText, Lock, RefreshCw, HelpCircle, Flag, Award, Copy, Table,
  GitBranch, Activity, Link, Circle, LogOut, Compass,
} from 'lucide-react';

export type IconName =
  | 'shield' | 'grid' | 'layers' | 'target' | 'flow' | 'package' | 'bell' | 'search'
  | 'user' | 'settings' | 'plus' | 'filter' | 'download' | 'upload'
  | 'chevronR' | 'chevronD' | 'chevronU' | 'x' | 'check' | 'alert' | 'book' | 'zap'
  | 'server' | 'database' | 'cloud' | 'sun' | 'moon' | 'eye' | 'trash' | 'edit'
  | 'sliders' | 'menu' | 'command' | 'external' | 'arrowR' | 'arrowL' | 'file'
  | 'lock' | 'refresh' | 'help' | 'russia' | 'award' | 'copy' | 'table'
  | 'layoutGrid' | 'git' | 'activity' | 'link' | 'dot' | 'logout' | 'compass';

type LucideCmp = React.ComponentType<{
  size?: number | string;
  color?: string;
  strokeWidth?: number | string;
}>;

const map: Record<IconName, LucideCmp> = {
  shield: Shield as LucideCmp,
  grid: LayoutGrid as LucideCmp,
  layers: Layers as LucideCmp,
  target: Target as LucideCmp,
  flow: Share2 as LucideCmp,
  package: Package as LucideCmp,
  bell: Bell as LucideCmp,
  search: Search as LucideCmp,
  user: User as LucideCmp,
  settings: Settings as LucideCmp,
  plus: Plus as LucideCmp,
  filter: Filter as LucideCmp,
  download: Download as LucideCmp,
  upload: Upload as LucideCmp,
  chevronR: ChevronRight as LucideCmp,
  chevronD: ChevronDown as LucideCmp,
  chevronU: ChevronUp as LucideCmp,
  x: X as LucideCmp,
  check: Check as LucideCmp,
  alert: AlertTriangle as LucideCmp,
  book: BookOpen as LucideCmp,
  zap: Zap as LucideCmp,
  server: Server as LucideCmp,
  database: Database as LucideCmp,
  cloud: Cloud as LucideCmp,
  sun: Sun as LucideCmp,
  moon: Moon as LucideCmp,
  eye: Eye as LucideCmp,
  trash: Trash2 as LucideCmp,
  edit: Edit as LucideCmp,
  sliders: SlidersHorizontal as LucideCmp,
  menu: Menu as LucideCmp,
  command: Command as LucideCmp,
  external: ExternalLink as LucideCmp,
  arrowR: ArrowRight as LucideCmp,
  arrowL: ArrowLeft as LucideCmp,
  file: FileText as LucideCmp,
  lock: Lock as LucideCmp,
  refresh: RefreshCw as LucideCmp,
  help: HelpCircle as LucideCmp,
  russia: Flag as LucideCmp,
  award: Award as LucideCmp,
  copy: Copy as LucideCmp,
  table: Table as LucideCmp,
  layoutGrid: LayoutGrid as LucideCmp,
  git: GitBranch as LucideCmp,
  activity: Activity as LucideCmp,
  link: Link as LucideCmp,
  dot: Circle as LucideCmp,
  logout: LogOut as LucideCmp,
  compass: Compass as LucideCmp,
};

export interface IconProps {
  name: IconName;
  size?: number;
  color?: string;
  strokeWidth?: number;
}

export const Icon: React.FC<IconProps> = ({
  name,
  size = 16,
  color = 'currentColor',
  strokeWidth = 1.75,
}) => {
  const Cmp = map[name];
  if (!Cmp) return null;
  return <Cmp size={size} color={color} strokeWidth={strokeWidth} />;
};
