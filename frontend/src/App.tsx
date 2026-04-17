import React from "react";
import { BrowserRouter, Routes, Route, NavLink } from "react-router-dom";
import { AuthProvider, useAuth } from "./context/AuthContext";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { AssetsPage } from "./pages/AssetsPage";
import { RiskPreviewPage } from "./pages/RiskPreviewPage";
import { RiskMapPage } from "./pages/RiskMapPage";
import { RiskGraphPage } from "./pages/RiskGraphPage";
import { AssetFormPage } from "./pages/AssetFormPage";
import { AssetRiskProfilePage } from "./pages/AssetRiskProfilePage";
import { SoftwareCatalogPage } from "./pages/SoftwareCatalogPage";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { Toaster } from "react-hot-toast";
import { Shield, LayoutDashboard, Map, TestTube, BookOpen, LogOut, User } from "lucide-react";
import "./App.css";

const roleBadge: Record<string, { label: string; color: string }> = {
    admin: { label: "Админ", color: "var(--danger)" },
    auditor: { label: "Аудитор", color: "var(--warning)" },
    viewer: { label: "Просмотр", color: "var(--info)" },
};

const UserSection: React.FC = () => {
    const { user, logout } = useAuth();
    if (!user) return null;

    const badge = roleBadge[user.role] || { label: user.role, color: "#94a3b8" };

    return (
        <div className="user-section">
            <div className="user-info">
                <div className="user-avatar">
                    <User size={20} />
                </div>
                <div className="user-details">
                    <span className="user-name">{user.username}</span>
                    <span className="role-badge" style={{ background: badge.color }}>
                        {badge.label}
                    </span>
                </div>
            </div>
            <button className="logout-btn" onClick={logout} title="Выход">
                <LogOut size={16} />
                <span>Выход</span>
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
            <aside className="app-sidebar">
                <div className="sidebar-header">
                    <div className="logo-icon">
                        <Shield size={20} strokeWidth={2.5} />
                    </div>
                    <div className="logo-text">
                        <div className="logo-title">CYBERRISK</div>
                        <div className="logo-subtitle">Управление рисками</div>
                    </div>
                </div>

                <nav className="nav-menu">
                    <NavLink to="/assets" className={({ isActive }) => `nav-link ${isActive ? "nav-link-active" : ""}`}>
                        <LayoutDashboard size={20} />
                        <span>Активы</span>
                    </NavLink>
                    <NavLink to="/software" className={({ isActive }) => `nav-link ${isActive ? "nav-link-active" : ""}`}>
                        <BookOpen size={20} />
                        <span>Справочник ПО</span>
                    </NavLink>
                    <NavLink to="/risk/preview" className={({ isActive }) => `nav-link ${isActive ? "nav-link-active" : ""}`}>
                        <TestTube size={20} />
                        <span>Симулятор риска</span>
                    </NavLink>
                    <NavLink to="/risk/map" className={({ isActive }) => `nav-link ${isActive ? "nav-link-active" : ""}`}>
                        <Map size={20} />
                        <span>Карта рисков</span>
                    </NavLink>
                </nav>

                <div className="sidebar-footer">
                    <UserSection />
                </div>
            </aside>

            <div className="main-wrapper">
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
                        <Route path="/risk/graph/:assetId" element={<ProtectedRoute><RiskGraphPage /></ProtectedRoute>} />
                        <Route path="*" element={<ProtectedRoute><AssetsPage /></ProtectedRoute>} />
                    </Routes>
                </main>

                <footer className="app-footer">
                    <div className="footer-content">
                        <div className="footer-text">
                            © {new Date().getFullYear()} CyberRisk Platform. Корпоративная система управления киберрисками.
                        </div>
                        <div className="footer-links">
                            <span className="footer-badge">v1.0.0</span>
                        </div>
                    </div>
                </footer>
            </div>
        </div>
    );
};

const App: React.FC = () => {
    return (
        <BrowserRouter>
            <AuthProvider>
                <AppLayout />
                <Toaster 
                    position="top-right" 
                    toastOptions={{
                        style: {
                            background: 'var(--raised)',
                            color: 'var(--ink)',
                            border: '1px solid var(--perimeter)',
                            boxShadow: 'var(--shadow-md)',
                            fontSize: '14px',
                            fontWeight: 500
                        },
                        success: {
                            iconTheme: {
                                primary: 'var(--success)',
                                secondary: '#fff',
                            },
                        },
                        error: {
                            iconTheme: {
                                primary: 'var(--danger)',
                                secondary: '#fff',
                            },
                        },
                    }} 
                />
            </AuthProvider>
        </BrowserRouter>
    );
};

export default App;
