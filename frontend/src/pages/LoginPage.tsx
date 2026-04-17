import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { Shield, User, Lock, LogIn } from "lucide-react";
import "./AuthPages.css";

export const LoginPage: React.FC = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [loading, setLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        try {
            await login(username, password);
            toast.success("Вход выполнен успешно");
            navigate("/", { replace: true });
        } catch (err: any) {
            toast.error(err.message || "Ошибка входа");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="auth-page">
            <motion.div 
                className="auth-card card"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.4 }}
            >
                <div className="auth-card__header">
                    <div className="auth-logo">
                        <div className="auth-logo__icon">
                            <Shield size={32} color="var(--command)" strokeWidth={2.5} />
                        </div>
                        <div className="auth-logo__text">
                            <span className="auth-logo__title">CYBERRISK</span>
                            <span className="auth-logo__sub">ENTERPRISE SYSTEM</span>
                        </div>
                    </div>
                    <div className="auth-card__divider" />
                    <h2 className="auth-card__title">Вход в систему</h2>
                    <p className="auth-card__desc">Введите корпоративные учётные данные</p>
                </div>

                <form onSubmit={handleSubmit} className="auth-form">
                    <div className="auth-field">
                        <label className="form-label">Пользователь</label>
                        <div className="input-with-icon">
                            <User size={18} className="input-icon" />
                            <input
                                className="form-input"
                                type="text"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                placeholder="username"
                                required
                                autoFocus
                                autoComplete="username"
                            />
                        </div>
                    </div>
                    <div className="auth-field">
                        <label className="form-label">Пароль</label>
                        <div className="input-with-icon">
                            <Lock size={18} className="input-icon" />
                            <input
                                className="form-input"
                                type="password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                placeholder="••••••••"
                                required
                                autoComplete="current-password"
                            />
                        </div>
                    </div>
                    <button className="btn btn-primary" type="submit" disabled={loading} style={{ width: '100%', marginTop: '10px' }}>
                        {loading ? (
                            <span className="loading-spinner" style={{ width: 18, height: 18 }} />
                        ) : (
                            <>
                                <LogIn size={18} />
                                Войти
                            </>
                        )}
                    </button>
                </form>

                <div className="auth-footer">
                    <span className="auth-footer__text">Нет аккаунта?</span>
                    <Link to="/register" className="auth-footer__link">Зарегистрироваться</Link>
                </div>
            </motion.div>
        </div>
    );
};
