import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import "./AuthPages.css";

export const LoginPage: React.FC = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [error, setError] = useState("");
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        setLoading(true);
        try {
            await login(username, password);
            navigate("/", { replace: true });
        } catch (err: any) {
            setError(err.message || "Ошибка входа");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-page">
            <div className="auth-grid-bg" />
            <div className="auth-scanlines" />
            <div className="auth-glow auth-glow--top" />
            <div className="auth-glow auth-glow--bottom" />

            <div className="auth-card">
                <div className="auth-card__header">
                    <div className="auth-logo">
                        <div className="auth-logo__icon">
                            <svg width="40" height="40" viewBox="0 0 40 40" fill="none">
                                <rect width="40" height="40" rx="6" fill="#0a0a0f" stroke="#00e5ff" strokeWidth="1.5" />
                                <path d="M20 10L25 15H15L20 10Z M20 30L15 25H25L20 30Z M10 20L15 15V25L10 20Z M30 20L25 25V15L30 20Z" fill="#00e5ff" opacity="0.9" />
                                <circle cx="20" cy="20" r="3" fill="none" stroke="#00e5ff" strokeWidth="1" opacity="0.5" />
                            </svg>
                        </div>
                        <div className="auth-logo__text">
                            <span className="auth-logo__title">CYBERRISK</span>
                            <span className="auth-logo__sub">THREAT MANAGEMENT SYSTEM</span>
                        </div>
                    </div>
                    <div className="auth-card__divider" />
                    <h2 className="auth-card__title">Авторизация</h2>
                    <p className="auth-card__desc">Введите учётные данные для входа в систему</p>
                </div>

                {error && (
                    <div className="auth-alert auth-alert--error">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                            <path d="M8 1a7 7 0 100 14A7 7 0 008 1zm0 10.5a.75.75 0 110-1.5.75.75 0 010 1.5zM8.75 4.5v4a.75.75 0 01-1.5 0v-4a.75.75 0 011.5 0z"/>
                        </svg>
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="auth-form">
                    <div className="auth-field">
                        <label className="auth-label">
                            <span className="auth-label__icon">&#x276F;</span>
                            Пользователь
                        </label>
                        <input
                            className="auth-input"
                            type="text"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="username"
                            required
                            autoFocus
                            autoComplete="username"
                        />
                    </div>
                    <div className="auth-field">
                        <label className="auth-label">
                            <span className="auth-label__icon">&#x276F;</span>
                            Пароль
                        </label>
                        <input
                            className="auth-input"
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="&#8226;&#8226;&#8226;&#8226;&#8226;&#8226;&#8226;&#8226;"
                            required
                            autoComplete="current-password"
                        />
                    </div>
                    <button className="auth-btn" type="submit" disabled={loading}>
                        {loading ? (
                            <span className="auth-btn__loading">
                                <span className="auth-spinner" />
                                Вход...
                            </span>
                        ) : (
                            <>
                                <span>Войти в систему</span>
                                <svg width="18" height="18" viewBox="0 0 18 18" fill="currentColor">
                                    <path d="M3 9h10M10 5l4 4-4 4" stroke="currentColor" strokeWidth="2" fill="none" strokeLinecap="round"/>
                                </svg>
                            </>
                        )}
                    </button>
                </form>

                <div className="auth-footer">
                    <span className="auth-footer__text">Нет аккаунта?</span>
                    <Link to="/register" className="auth-footer__link">Зарегистрироваться</Link>
                </div>

                <div className="auth-card__status">
                    <span className="auth-status-dot" />
                    <span>Система активна</span>
                    <span className="auth-status-ver">v1.0.0</span>
                </div>
            </div>
        </div>
    );
};
