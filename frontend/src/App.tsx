import React from "react";
import { BrowserRouter, Routes, Route, NavLink } from "react-router-dom";
import { AuthProvider, useAuth } from "./context/AuthContext";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { AssetsPage } from "./pages/AssetsPage";
import { RiskPreviewPage } from "./pages/RiskPreviewPage";
import { RiskMapPage } from "./pages/RiskMapPage";
import { AssetFormPage } from "./pages/AssetFormPage";
import { AssetRiskProfilePage } from "./pages/AssetRiskProfilePage";
import { SoftwareCatalogPage } from "./pages/SoftwareCatalogPage";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import "./App.css";

const roleBadge: Record<string, { label: string; color: string }> = {
    admin: { label: "Админ", color: "#e74c3c" },
    auditor: { label: "Аудитор", color: "#f39c12" },
    viewer: { label: "Просмотр", color: "#3498db" },
};

const UserSection: React.FC = () => {
    const { user, logout } = useAuth();
    if (!user) return null;

    const badge = roleBadge[user.role] || { label: user.role, color: "#95a5a6" };

    return (
        <div className="user-section">
            <span className="user-name">{user.username}</span>
            <span className="role-badge" style={{ background: badge.color }}>
                {badge.label}
            </span>
            <button className="logout-btn" onClick={logout} title="Выход">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M6 2h6a2 2 0 012 2v8a2 2 0 01-2 2H6v-2h6V4H6V2zm1 3l3 3-3 3V9H2V7h5V5z" />
                </svg>
                Выход
            </button>
        </div>
    );
};

const AppLayout: React.FC = () => {
    const { isAuthenticated } = useAuth();

    if (!isAuthenticated) {
        return (
            <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegisterPage />} />
                <Route path="*" element={<LoginPage />} />
            </Routes>
        );
    }

    return (
        <div className="app-container">
            <header className="app-header">
                <div className="header-content">
                    <div className="logo-section">
                        <div className="logo-icon">
                            <svg width="28" height="28" viewBox="0 0 40 40" fill="none">
                                <rect width="40" height="40" rx="6" fill="#0a0a14" stroke="#00e5ff" strokeWidth="1.5" />
                                <path d="M20 10L25 15H15L20 10Z M20 30L15 25H25L20 30Z M10 20L15 15V25L10 20Z M30 20L25 25V15L30 20Z" fill="#00e5ff" opacity="0.9" />
                            </svg>
                        </div>
                        <div className="logo-text">
                            <div className="logo-title">CYBERRISK</div>
                            <div className="logo-subtitle">Threat Management</div>
                        </div>
                    </div>

                    <nav className="nav-menu">
                        <NavLink to="/assets" className={({ isActive }) => `nav-link ${isActive ? "nav-link-active" : ""}`}>
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                <path d="M2 3h12v2H2V3zm0 4h12v2H2V7zm0 4h12v2H2v-2z" />
                            </svg>
                            Активы
                        </NavLink>
                        <NavLink to="/software" className={({ isActive }) => `nav-link ${isActive ? "nav-link-active" : ""}`}>
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                <path d="M3 2h10a1 1 0 011 1v10a1 1 0 01-1 1H3a1 1 0 01-1-1V3a1 1 0 011-1zm1 2v8h8V4H4zm1 1h2v2H5V5zm3 0h3v1H8V5zm0 2h3v1H8V7zm-3 2h6v1H5V9z" />
                            </svg>
                            Справочник ПО
                        </NavLink>
                        <NavLink to="/risk/preview" className={({ isActive }) => `nav-link ${isActive ? "nav-link-active" : ""}`}>
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                <path d="M8 2l6 12H2L8 2z" />
                            </svg>
                            Симулятор риска
                        </NavLink>
                        <NavLink to="/risk/map" className={({ isActive }) => `nav-link ${isActive ? "nav-link-active" : ""}`}>
                            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                <rect x="1" y="1" width="6" height="6" />
                                <rect x="9" y="1" width="6" height="6" />
                                <rect x="1" y="9" width="6" height="6" />
                                <rect x="9" y="9" width="6" height="6" />
                            </svg>
                            Карта рисков
                        </NavLink>
                    </nav>

                    <UserSection />
                </div>
            </header>

            <main className="main-content">
                <Routes>
                    <Route path="/" element={<ProtectedRoute><AssetsPage /></ProtectedRoute>} />
                    <Route path="/assets" element={<ProtectedRoute><AssetsPage /></ProtectedRoute>} />
                    <Route path="/assets/new" element={<ProtectedRoute><AssetFormPage /></ProtectedRoute>} />
                    <Route path="/assets/edit/:id" element={<ProtectedRoute><AssetFormPage /></ProtectedRoute>} />
                    <Route path="/assets/:id/risks" element={<ProtectedRoute><AssetRiskProfilePage /></ProtectedRoute>} />
                    <Route path="/software" element={<ProtectedRoute><SoftwareCatalogPage /></ProtectedRoute>} />
                    <Route path="/risk/preview" element={<ProtectedRoute><RiskPreviewPage /></ProtectedRoute>} />
                    <Route path="/risk/map" element={<ProtectedRoute><RiskMapPage /></ProtectedRoute>} />
                    <Route path="*" element={<ProtectedRoute><AssetsPage /></ProtectedRoute>} />
                </Routes>
            </main>

            <footer className="app-footer">
                <div className="footer-content">
                    <div className="footer-text">
                        © {new Date().getFullYear()} CyberRisk Platform. Локальная система управления киберрисками.
                    </div>
                    <div className="footer-links">
                        <span className="footer-badge">v1.0.0</span>
                    </div>
                </div>
            </footer>
        </div>
    );
};

const App: React.FC = () => {
    return (
        <BrowserRouter>
            <AuthProvider>
                <AppLayout />
            </AuthProvider>
        </BrowserRouter>
    );
};

export default App;
