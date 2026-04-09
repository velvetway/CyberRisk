import React from "react";
import { BrowserRouter, Routes, Route, NavLink } from "react-router-dom";
import { AssetsPage } from "./pages/AssetsPage";
import { RiskPreviewPage } from "./pages/RiskPreviewPage";
import { RiskMapPage } from "./pages/RiskMapPage";
import { AssetFormPage } from "./pages/AssetFormPage";
import { AssetRiskProfilePage } from "./pages/AssetRiskProfilePage";
import { SoftwareCatalogPage } from "./pages/SoftwareCatalogPage";
import "./App.css";

const App: React.FC = () => {
    return (
        <BrowserRouter>
            <div className="app-container">
                <header className="app-header">
                    <div className="header-content">
                        <div className="logo-section">
                            <div className="logo-icon">
                                <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
                                    <rect width="32" height="32" rx="8" fill="url(#gradient)" />
                                    <path
                                        d="M16 8L20 12H12L16 8Z M16 24L12 20H20L16 24Z M8 16L12 12V20L8 16Z M24 16L20 20V12L24 16Z"
                                        fill="white"
                                        opacity="0.9"
                                    />
                                    <defs>
                                        <linearGradient id="gradient" x1="0" y1="0" x2="32" y2="32">
                                            <stop offset="0%" stopColor="#e74c3c" />
                                            <stop offset="100%" stopColor="#c0392b" />
                                        </linearGradient>
                                    </defs>
                                </svg>
                            </div>
                            <div className="logo-text">
                                <div className="logo-title">CyberRisk</div>
                                <div className="logo-subtitle">Управление киберрисками</div>
                            </div>
                        </div>

                        <nav className="nav-menu">
                            <NavLink
                                to="/assets"
                                className={({ isActive }) =>
                                    `nav-link ${isActive ? "nav-link-active" : ""}`
                                }
                            >
                                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                    <path d="M2 3h12v2H2V3zm0 4h12v2H2V7zm0 4h12v2H2v-2z" />
                                </svg>
                                Активы
                            </NavLink>

                            <NavLink
                                to="/software"
                                className={({ isActive }) =>
                                    `nav-link ${isActive ? "nav-link-active" : ""}`
                                }
                            >
                                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                    <path d="M3 2h10a1 1 0 011 1v10a1 1 0 01-1 1H3a1 1 0 01-1-1V3a1 1 0 011-1zm1 2v8h8V4H4zm1 1h2v2H5V5zm3 0h3v1H8V5zm0 2h3v1H8V7zm-3 2h6v1H5V9z" />
                                </svg>
                                Справочник ПО
                            </NavLink>

                            <NavLink
                                to="/risk/preview"
                                className={({ isActive }) =>
                                    `nav-link ${isActive ? "nav-link-active" : ""}`
                                }
                            >
                                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                    <path d="M8 2l6 12H2L8 2z" />
                                </svg>
                                Симулятор риска
                            </NavLink>

                            <NavLink
                                to="/risk/map"
                                className={({ isActive }) =>
                                    `nav-link ${isActive ? "nav-link-active" : ""}`
                                }
                            >
                                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                                    <rect x="1" y="1" width="6" height="6" />
                                    <rect x="9" y="1" width="6" height="6" />
                                    <rect x="1" y="9" width="6" height="6" />
                                    <rect x="9" y="9" width="6" height="6" />
                                </svg>
                                Карта рисков
                            </NavLink>
                        </nav>
                    </div>
                </header>

                <main className="main-content">
                    <Routes>
                        <Route path="/" element={<AssetsPage />} />
                        <Route path="/assets" element={<AssetsPage />} />
                        <Route path="/assets/new" element={<AssetFormPage />} />
                        <Route path="/assets/edit/:id" element={<AssetFormPage />} />
                        <Route path="/assets/:id/risks" element={<AssetRiskProfilePage />} />
                        <Route path="/software" element={<SoftwareCatalogPage />} />
                        <Route path="/risk/preview" element={<RiskPreviewPage />} />
                        <Route path="/risk/map" element={<RiskMapPage />} />
                    </Routes>
                </main>

                <footer className="app-footer">
                    <div className="footer-content">
                        <div className="footer-text">
                            © {new Date().getFullYear()} CyberRisk Platform. Локальная система
                            управления киберрисками.
                        </div>
                        <div className="footer-links">
                            <span className="footer-badge">v1.0.0</span>
                        </div>
                    </div>
                </footer>
            </div>
        </BrowserRouter>
    );
};

export default App;
