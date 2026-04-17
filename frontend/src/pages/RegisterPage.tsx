import React, { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import toast from "react-hot-toast";
import { useAuth } from "../context/AuthContext";
import { Btn, Icon } from "../components/design";

export const RegisterPage: React.FC = () => {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("viewer");
  const [loading, setLoading] = useState(false);

  const onSubmit = async (e: React.FormEvent) => {
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
      toast.error(err?.message || "Не удалось зарегистрироваться");
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
          width: 420,
          background: "var(--bg-elev-1)",
          border: "1px solid var(--border)",
          borderRadius: "var(--r-lg)",
          padding: 32,
          boxShadow: "var(--sh-lg)",
        }}
      >
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
            Создание аккаунта
          </h1>
          <div style={{ fontSize: "var(--text-sm)", color: "var(--fg-muted)", marginTop: 4 }}>
            Заполните данные учётной записи
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
                minLength={3}
                autoFocus
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
                minLength={6}
                placeholder="Минимум 6 символов"
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
              Роль
            </div>
            <div style={{ display: "flex", gap: 6 }}>
              {(
                [
                  { v: "viewer", l: "Viewer" },
                  { v: "auditor", l: "Auditor" },
                  { v: "admin", l: "Admin" },
                ] as const
              ).map((r) => (
                <button
                  key={r.v}
                  type="button"
                  onClick={() => setRole(r.v)}
                  style={{
                    flex: 1,
                    height: 34,
                    background: role === r.v ? "var(--accent-ghost)" : "var(--bg-elev-2)",
                    color: role === r.v ? "var(--accent)" : "var(--fg-muted)",
                    border: `1px solid ${role === r.v ? "var(--accent)" : "var(--border)"}`,
                    borderRadius: "var(--r-sm)",
                    fontSize: "var(--text-sm)",
                    cursor: "pointer",
                    transition: "var(--transition)",
                  }}
                >
                  {r.l}
                </button>
              ))}
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
            {loading ? "Создание…" : "Создать аккаунт"}
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
          Уже есть аккаунт?{" "}
          <Link
            to="/login"
            style={{ color: "var(--accent)", textDecoration: "none", fontWeight: 500 }}
          >
            Войти
          </Link>
        </div>
      </motion.div>
    </div>
  );
};
