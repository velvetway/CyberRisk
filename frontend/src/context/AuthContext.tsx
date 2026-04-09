import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import { loginAPI, registerAPI } from "../api/client";

interface User {
    id: number;
    username: string;
    role: string;
}

interface AuthContextType {
    user: User | null;
    token: string | null;
    isAuthenticated: boolean;
    login: (username: string, password: string) => Promise<void>;
    register: (username: string, password: string, role: string) => Promise<void>;
    logout: () => void;
    canWrite: boolean;
    canDelete: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

function parseToken(token: string): User | null {
    try {
        const payload = JSON.parse(atob(token.split(".")[1]));
        return {
            id: payload.user_id,
            username: payload.username,
            role: payload.role,
        };
    } catch {
        return null;
    }
}

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [token, setToken] = useState<string | null>(() => localStorage.getItem("token"));
    const [user, setUser] = useState<User | null>(() => {
        const t = localStorage.getItem("token");
        return t ? parseToken(t) : null;
    });

    const logout = useCallback(() => {
        localStorage.removeItem("token");
        setToken(null);
        setUser(null);
    }, []);

    useEffect(() => {
        const handler = () => logout();
        window.addEventListener("auth:logout", handler);
        return () => window.removeEventListener("auth:logout", handler);
    }, [logout]);

    const login = async (username: string, password: string) => {
        const res = await loginAPI(username, password);
        localStorage.setItem("token", res.token);
        setToken(res.token);
        setUser(parseToken(res.token));
    };

    const register = async (username: string, password: string, role: string) => {
        await registerAPI(username, password, role);
    };

    const role = user?.role || "";
    const canWrite = role === "admin" || role === "auditor";
    const canDelete = role === "admin";

    return (
        <AuthContext.Provider
            value={{
                user,
                token,
                isAuthenticated: !!token,
                login,
                register,
                logout,
                canWrite,
                canDelete,
            }}
        >
            {children}
        </AuthContext.Provider>
    );
};

export function useAuth(): AuthContextType {
    const ctx = useContext(AuthContext);
    if (!ctx) throw new Error("useAuth must be used within AuthProvider");
    return ctx;
}
