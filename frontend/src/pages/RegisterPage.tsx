import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import "./AuthPages.css";

export const RegisterPage: React.FC = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [role, setRole] = useState("viewer");
    const [error, setError] = useState("");
    const [success, setSuccess] = useState("");
    const [loading, setLoading] = useState(false);
    const { register } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError("");
        setSuccess("");

        if (password.length < 6) {
            setError("Пароль должен быть не менее 6 символов");
            return;
        }

        setLoading(true);
        try {
            await register(username, password, role);
            setSuccess("Регистрация выполнена. Перенаправление...");
            setTimeout(() => navigate("/login"), 1200);
        } catch (err: any) {
            setError(err.message || "Ошибка регистрации");
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
                    <h2 className="auth-card__title">Регистрация</h2>
                    <p className="auth-card__desc">Создание нового оператора системы</p>
                </div>

                {error && (
                    <div className="auth-alert auth-alert--error">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                            <path d="M8 1a7 7 0 100 14A7 7 0 008 1zm0 10.5a.75.75 0 110-1.5.75.75 0 010 1.5zM8.75 4.5v4a.75.75 0 01-1.5 0v-4a.75.75 0 011.5 0z"/>
                        </svg>
                        {error}
                    </div>
                )}
                {success && (
                    <div className="auth-alert auth-alert--success">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                            <path d="M8 1a7 7 0 100 14A7 7 0 008 1zm3.22 5.97l-3.5 3.5a.75.75 0 01-1.06 0l-1.5-1.5a.75.75 0 111.06-1.06l.97.97 2.97-2.97a.75.75 0 111.06 1.06z"/>
                        </svg>
                        {success}
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
                            placeholder="Минимум 6 символов"
                            required
                        />
                    </div>
                    <div className="auth-field">
                        <label className="auth-label">
                            <span className="auth-label__icon">&#x276F;</span>
                            Уровень доступа
                        </label>
                        <select
                            className="auth-input auth-select"
                            value={role}
                            onChange={(e) => setRole(e.target.value)}
                        >
                            <option value="viewer">Наблюдатель (viewer)</option>
                            <option value="auditor">Аудитор (auditor)</option>
                            <option value="admin">Администратор (admin)</option>
                        </select>
                    </div>
                    <button className="auth-btn" type="submit" disabled={loading}>
                        {loading ? (
                            <span className="auth-btn__loading">
                                <span className="auth-spinner" />
                                Создание...
                            </span>
                        ) : (
                            <>
                                <span>Создать оператора</span>
                                <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
                                    <path d="M9 3v12M3 9h12" stroke="currentColor" strokeWidth="2" strokeLinecap="round"/>
                                </svg>
                            </>
                        )}
                    </button>
                </form>

                <div className="auth-footer">
                    <span className="auth-footer__text">Уже есть аккаунт?</span>
                    <Link to="/login" className="auth-footer__link">Войти</Link>
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
