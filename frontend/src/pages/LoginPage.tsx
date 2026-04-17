import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { useAuth } from "../context/AuthContext";
import { Btn, Icon } from "../components/design";

export const LoginPage: React.FC = () => {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await login(username, password);
      toast.success("Вход выполнен успешно");
      navigate("/", { replace: true });
    } catch (err: any) {
      toast.error(err?.message || "Не удалось войти");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        background: "var(--bg)",
        color: "var(--fg)",
        padding: 20,
      }}
    >
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        style={{
          width: 400,
          background: "var(--bg-elev-1)",
          border: "1px solid var(--border)",
          borderRadius: "var(--r-lg)",
          padding: 32,
          boxShadow: "var(--sh-lg)",
        }}
      >
        {/* Logo */}
        <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 24 }}>
          <div
            style={{
              width: 38,
              height: 38,
              borderRadius: 8,
              background: "var(--accent)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              color: "var(--accent-fg)",
              boxShadow: "0 4px 12px oklch(0 0 0 / 0.3)",
            }}
          >
            <Icon name="shield" size={20} />
          </div>
          <div>
            <div style={{ fontSize: "var(--text-lg)", fontWeight: 600, letterSpacing: "-0.01em" }}>
              CyberRisk
            </div>
            <div
              style={{
                fontSize: "var(--text-2xs)",
                color: "var(--fg-faint)",
                fontFamily: "var(--font-mono)",
                textTransform: "uppercase",
                letterSpacing: "0.08em",
              }}
            >
              Enterprise Platform
            </div>
          </div>
        </div>

        <div style={{ borderTop: "1px solid var(--border)", paddingTop: 22, marginBottom: 24 }}>
          <h1
            style={{
              margin: 0,
              fontSize: "var(--text-xl)",
              fontWeight: 600,
              letterSpacing: "-0.02em",
            }}
          >
            Вход в систему
          </h1>
          <div style={{ fontSize: "var(--text-sm)", color: "var(--fg-muted)", marginTop: 4 }}>
            Введите корпоративные учётные данные
          </div>
        </div>

        <form onSubmit={onSubmit} style={{ display: "flex", flexDirection: "column", gap: 14 }}>
          <label>
            <div
              style={{
                fontSize: "var(--text-xs)",
                color: "var(--fg-dim)",
                marginBottom: 6,
                textTransform: "uppercase",
                letterSpacing: "0.06em",
                fontWeight: 500,
              }}
            >
              Пользователь
            </div>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                padding: "0 12px",
                height: 38,
                background: "var(--bg-elev-2)",
                border: "1px solid var(--border)",
                borderRadius: "var(--r-sm)",
              }}
            >
              <Icon name="user" size={14} color="var(--fg-dim)" />
              <input
                type="text"
                required
                autoFocus
                autoComplete="username"
                placeholder="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                style={{
                  flex: 1,
                  background: "transparent",
                  border: "none",
                  outline: "none",
                  color: "var(--fg)",
                  fontSize: "var(--text-sm)",
                }}
              />
            </div>
          </label>
          <label>
            <div
              style={{
                fontSize: "var(--text-xs)",
                color: "var(--fg-dim)",
                marginBottom: 6,
                textTransform: "uppercase",
                letterSpacing: "0.06em",
                fontWeight: 500,
              }}
            >
              Пароль
            </div>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: 8,
                padding: "0 12px",
                height: 38,
                background: "var(--bg-elev-2)",
                border: "1px solid var(--border)",
                borderRadius: "var(--r-sm)",
              }}
            >
              <Icon name="lock" size={14} color="var(--fg-dim)" />
              <input
                type="password"
                required
                autoComplete="current-password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                style={{
                  flex: 1,
                  background: "transparent",
                  border: "none",
                  outline: "none",
                  color: "var(--fg)",
                  fontSize: "var(--text-sm)",
                  letterSpacing: "0.1em",
                }}
              />
            </div>
          </label>

          <Btn
            type="submit"
            variant="primary"
            size="lg"
            fullWidth
            disabled={loading}
            icon={<Icon name="arrowR" size={14} />}
          >
            {loading ? "Вход…" : "Войти"}
          </Btn>
        </form>

        <div
          style={{
            textAlign: "center",
            marginTop: 20,
            paddingTop: 20,
            borderTop: "1px dashed var(--border)",
            fontSize: "var(--text-sm)",
            color: "var(--fg-dim)",
          }}
        >
          Нет аккаунта?{" "}
          <Link
            to="/register"
            style={{ color: "var(--accent)", textDecoration: "none", fontWeight: 500 }}
          >
            Зарегистрироваться
          </Link>
        </div>
      </motion.div>
    </div>
  );
};
