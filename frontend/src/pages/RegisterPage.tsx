import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { Shield, User, Lock, UserPlus, ShieldCheck } from "lucide-react";
import "./AuthPages.css";

export const RegisterPage: React.FC = () => {
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [role, setRole] = useState("viewer");
    const [loading, setLoading] = useState(false);
    const { register } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (password.length < 6) {
            toast.error("Пароль должен быть не менее 6 символов");
            return;
        }

        setLoading(true);
        try {
            await register(username, password, role);
            toast.success("Регистрация выполнена");
            setTimeout(() => navigate("/login"), 1200);
        } catch (err: any) {
            toast.error(err.message || "Ошибка регистрации");
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
                    <h2 className="auth-card__title">Новый сотрудник</h2>
                    <p className="auth-card__desc">Создание учетной записи</p>
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
                                placeholder="Минимум 6 символов"
                                required
                            />
                        </div>
                    </div>
                    <div className="auth-field">
                        <label className="form-label">Доступ</label>
                        <div className="input-with-icon">
                            <ShieldCheck size={18} className="input-icon" style={{ zIndex: 2 }} />
                            <select
                                className="form-input"
                                value={role}
                                onChange={(e) => setRole(e.target.value)}
                                style={{ paddingLeft: '40px' }}
                            >
                                <option value="viewer">Наблюдатель</option>
                                <option value="auditor">Аудитор</option>
                                <option value="admin">Администратор</option>
                            </select>
                        </div>
                    </div>
                    <button className="btn btn-primary" type="submit" disabled={loading} style={{ width: '100%', marginTop: '10px' }}>
                        {loading ? (
                            <span className="loading-spinner" style={{ width: 18, height: 18 }} />
                        ) : (
                            <>
                                <UserPlus size={18} />
                                Зарегистрироваться
                            </>
                        )}
                    </button>
                </form>

                <div className="auth-footer">
                    <span className="auth-footer__text">Уже есть аккаунт?</span>
                    <Link to="/login" className="auth-footer__link">Войти</Link>
                </div>
            </motion.div>
        </div>
    );
};
