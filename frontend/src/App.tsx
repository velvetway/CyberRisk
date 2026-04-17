import React, { useEffect, useState } from "react";
import { BrowserRouter, Routes, Route, useLocation, Navigate } from "react-router-dom";
import { Toaster } from "react-hot-toast";

import { AuthProvider, useAuth } from "./context/AuthContext";
import { ProtectedRoute } from "./components/ProtectedRoute";

import { Sidebar, TopBar, CommandPalette, NAV_ITEMS, findActiveNavItem } from "./components/design";

import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { AssetsPage } from "./pages/AssetsPage";
import { AssetFormPage } from "./pages/AssetFormPage";
import { AssetRiskProfilePage } from "./pages/AssetRiskProfilePage";
import { SoftwareCatalogPage } from "./pages/SoftwareCatalogPage";
import { RiskPreviewPage } from "./pages/RiskPreviewPage";
import { RiskMapPage } from "./pages/RiskMapPage";
import { RiskGraphPage } from "./pages/RiskGraphPage";
import { DashboardPage } from "./pages/DashboardPage";

type Theme = 'dark' | 'light';
type Accent = 'indigo' | 'cyan' | 'emerald' | 'amber' | 'crimson';
type Density = 'default' | 'comfortable';

interface PersistedState {
  theme?: Theme;
  accent?: Accent;
  density?: Density;
}

function loadState(): PersistedState {
  try { return JSON.parse(localStorage.getItem('cr-state') || '{}'); }
  catch { return {}; }
}

const BC_MAP: Record<string, string[]> = {
  dashboard:  ['Платформа', 'Обзор'],
  riskmap:    ['Платформа', 'Карта рисков'],
  graph:      ['Платформа', 'Граф атаки'],
  assets:     ['Платформа', 'Реестр активов'],
  software:   ['Справочники', 'ПО (Минцифры)'],
  simulator:  ['Платформа', 'Симулятор риска'],
  threats:    ['Справочники', 'Каталог угроз'],
  vulns:      ['Справочники', 'Уязвимости'],
  reports:    ['Справочники', 'Отчёты'],
  settings:   ['Система', 'Настройки'],
};

// Silence unused-import lint for NAV_ITEMS (kept for future wiring / consistency).
void NAV_ITEMS;

const PageShell: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [theme, setTheme] = useState<Theme>(() => loadState().theme ?? 'dark');
  const [accent, setAccent] = useState<Accent>(() => loadState().accent ?? 'indigo');
  const [density, setDensity] = useState<Density>(() => loadState().density ?? 'default');
  const [collapsed, setCollapsed] = useState(false);
  const [cmdOpen, setCmdOpen] = useState(false);
  const location = useLocation();

  // Silence unused-setter lint — setters retained for future settings panel.
  void setAccent;
  void setDensity;

  useEffect(() => {
    localStorage.setItem('cr-state', JSON.stringify({ theme, accent, density }));
    document.documentElement.setAttribute('data-theme', theme);
    document.documentElement.setAttribute('data-accent', accent);
    document.documentElement.setAttribute('data-density', density);
  }, [theme, accent, density]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setCmdOpen(v => !v);
      }
      if (e.key === 'Escape') setCmdOpen(false);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  const activeItem = findActiveNavItem(location.pathname);
  const breadcrumbs = activeItem ? BC_MAP[activeItem.id] ?? ['Платформа'] : ['Платформа'];

  return (
    <div style={{ display: 'flex', minHeight: '100vh', color: 'var(--fg)', background: 'var(--bg)' }}>
      <Sidebar collapsed={collapsed} onToggleCollapsed={() => setCollapsed(v => !v)} />
      <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
        <TopBar
          breadcrumbs={breadcrumbs}
          onCmdK={() => setCmdOpen(true)}
          onThemeToggle={() => setTheme(t => (t === 'dark' ? 'light' : 'dark'))}
          theme={theme}
        />
        <main style={{ flex: 1, overflow: 'auto' }}>{children}</main>
      </div>
      <CommandPalette open={cmdOpen} onClose={() => setCmdOpen(false)} />
    </div>
  );
};

const LayoutGuard: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { token } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  return <PageShell>{children}</PageShell>;
};

// ProtectedRoute kept referenced so tree-shaking doesn't warn on unused import.
void ProtectedRoute;

function RoutedApp() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />

      {/* Protected pages wrapped in PageShell */}
      <Route path="/" element={<LayoutGuard><DashboardPage /></LayoutGuard>} />
      <Route path="/assets" element={<LayoutGuard><AssetsPage /></LayoutGuard>} />
      <Route path="/assets/new" element={<LayoutGuard><AssetFormPage /></LayoutGuard>} />
      <Route path="/assets/edit/:id" element={<LayoutGuard><AssetFormPage /></LayoutGuard>} />
      <Route path="/assets/:id/risks" element={<LayoutGuard><AssetRiskProfilePage /></LayoutGuard>} />
      <Route path="/software" element={<LayoutGuard><SoftwareCatalogPage /></LayoutGuard>} />
      <Route path="/risk/preview" element={<LayoutGuard><RiskPreviewPage /></LayoutGuard>} />
      <Route path="/risk/map" element={<LayoutGuard><RiskMapPage /></LayoutGuard>} />
      <Route path="/risk/graph/:assetId" element={<LayoutGuard><RiskGraphPage /></LayoutGuard>} />
      <Route path="*" element={<LayoutGuard><AssetsPage /></LayoutGuard>} />
    </Routes>
  );
}

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <RoutedApp />
        <Toaster position="top-right" toastOptions={{
          style: {
            background: 'var(--bg-elev-2)',
            color: 'var(--fg)',
            border: '1px solid var(--border)',
            fontSize: 'var(--text-sm)'
          }
        }} />
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
