import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

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
        <div style={styles.container}>
            <div style={styles.card}>
                <div style={styles.logoSection}>
                    <div style={styles.logoIcon}>
                        <svg width="48" height="48" viewBox="0 0 32 32" fill="none">
                            <rect width="32" height="32" rx="8" fill="url(#g)" />
                            <path d="M16 8L20 12H12L16 8Z M16 24L12 20H20L16 24Z M8 16L12 12V20L8 16Z M24 16L20 20V12L24 16Z" fill="white" opacity="0.9" />
                            <defs><linearGradient id="g" x1="0" y1="0" x2="32" y2="32"><stop offset="0%" stopColor="#e74c3c" /><stop offset="100%" stopColor="#c0392b" /></linearGradient></defs>
                        </svg>
                    </div>
                    <h1 style={styles.title}>CyberRisk</h1>
                    <p style={styles.subtitle}>Вход в систему</p>
                </div>

                {error && <div style={styles.error}>{error}</div>}

                <form onSubmit={handleSubmit} style={styles.form}>
                    <div style={styles.field}>
                        <label style={styles.label}>Имя пользователя</label>
                        <input
                            style={styles.input}
                            type="text"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="admin"
                            required
                            autoFocus
                        />
                    </div>
                    <div style={styles.field}>
                        <label style={styles.label}>Пароль</label>
                        <input
                            style={styles.input}
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            placeholder="******"
                            required
                        />
                    </div>
                    <button style={styles.button} type="submit" disabled={loading}>
                        {loading ? "Вход..." : "Войти"}
                    </button>
                </form>

                <p style={styles.link}>
                    Нет аккаунта?{" "}
                    <Link to="/register" style={styles.a}>Зарегистрироваться</Link>
                </p>
            </div>
        </div>
    );
};

const styles: Record<string, React.CSSProperties> = {
    container: {
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        minHeight: "100vh",
        background: "linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)",
    },
    card: {
        background: "white",
        borderRadius: "16px",
        padding: "48px 40px",
        width: "100%",
        maxWidth: "420px",
        boxShadow: "0 20px 60px rgba(0, 0, 0, 0.3)",
    },
    logoSection: {
        textAlign: "center" as const,
        marginBottom: "32px",
    },
    logoIcon: {
        marginBottom: "16px",
    },
    title: {
        fontSize: "28px",
        fontWeight: 700,
        color: "#2c3e50",
        margin: "0 0 4px 0",
    },
    subtitle: {
        fontSize: "14px",
        color: "#7f8c8d",
        margin: 0,
    },
    error: {
        background: "#fef2f2",
        color: "#dc2626",
        padding: "12px 16px",
        borderRadius: "8px",
        fontSize: "14px",
        marginBottom: "20px",
        border: "1px solid #fecaca",
    },
    form: {
        display: "flex",
        flexDirection: "column" as const,
        gap: "20px",
    },
    field: {
        display: "flex",
        flexDirection: "column" as const,
        gap: "6px",
    },
    label: {
        fontSize: "13px",
        fontWeight: 600,
        color: "#374151",
    },
    input: {
        padding: "12px 16px",
        borderRadius: "8px",
        border: "1px solid #d1d5db",
        fontSize: "15px",
        outline: "none",
        transition: "border-color 0.2s",
    },
    button: {
        padding: "14px",
        borderRadius: "8px",
        border: "none",
        background: "linear-gradient(135deg, #e74c3c, #c0392b)",
        color: "white",
        fontSize: "16px",
        fontWeight: 600,
        cursor: "pointer",
        marginTop: "8px",
    },
    link: {
        textAlign: "center" as const,
        marginTop: "24px",
        fontSize: "14px",
        color: "#6b7280",
    },
    a: {
        color: "#e74c3c",
        textDecoration: "none",
        fontWeight: 500,
    },
};
