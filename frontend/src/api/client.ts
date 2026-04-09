// src/api/client.ts
import {
    Asset,
    Threat,
    RiskPreviewRequest,
    RiskPreviewResponse,
    RiskOverviewPoint,
    Software,
    SoftwareCategory,
    AssetSoftwareAlternative,
} from "../types";

function getToken(): string | null {
    return localStorage.getItem("token");
}

async function request<T>(input: string, init?: RequestInit): Promise<T> {
    const token = getToken();
    const headers: Record<string, string> = {
        "Content-Type": "application/json",
        ...(init?.headers as Record<string, string> || {}),
    };
    if (token) {
        headers["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetch(input, {
        ...init,
        headers,
    });

    if (res.status === 401) {
        localStorage.removeItem("token");
        localStorage.removeItem("user");
        window.dispatchEvent(new Event("auth:logout"));
        throw new Error("Сессия истекла. Войдите снова.");
    }

    if (!res.ok) {
        let message = `HTTP ${res.status}`;
        try {
            const data = await res.json();
            if (data && typeof data.error === "string") {
                message = data.error;
            }
        } catch {
            // тело пустое или не JSON — оставляем дефолт
        }
        throw new Error(message);
    }

    if (res.status === 204) {
        return undefined as T;
    }

    const text = await res.text();
    if (!text) {
        return undefined as T;
    }

    return JSON.parse(text) as T;
}

// Auth API (no token needed)
export interface LoginResponse {
    token: string;
}

export interface UserResponse {
    id: number;
    username: string;
    role: string;
    is_active: boolean;
}

export async function loginAPI(username: string, password: string): Promise<LoginResponse> {
    const res = await fetch("/api/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
    });
    if (!res.ok) {
        let message = "Ошибка входа";
        try {
            const data = await res.json();
            if (data?.error) message = data.error;
        } catch {}
        throw new Error(message);
    }
    return res.json();
}

export async function registerAPI(username: string, password: string, role: string): Promise<UserResponse> {
    const res = await fetch("/api/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password, role }),
    });
    if (!res.ok) {
        let message = "Ошибка регистрации";
        try {
            const data = await res.json();
            if (data?.error) message = data.error;
        } catch {}
        throw new Error(message);
    }
    return res.json();
}

export const api = {
    // Assets
    getAssets(): Promise<Asset[]> {
        return request<Asset[]>("/api/assets");
    },

    getAsset(id: number): Promise<Asset> {
        return request<Asset>(`/api/assets/${id}`);
    },

    createAsset(asset: Partial<Asset>): Promise<Asset> {
        return request<Asset>("/api/assets", {
            method: "POST",
            body: JSON.stringify(asset),
        });
    },

    updateAsset(id: number, asset: Partial<Asset>): Promise<Asset> {
        return request<Asset>(`/api/assets/${id}`, {
            method: "PUT",
            body: JSON.stringify(asset),
        });
    },

    deleteAsset(id: number): Promise<void> {
        return request<void>(`/api/assets/${id}`, {
            method: "DELETE",
        });
    },

    async getAssetSoftwareAlternatives(id: number): Promise<AssetSoftwareAlternative[]> {
        try {
            return await request<AssetSoftwareAlternative[]>(`/api/assets/${id}/software/alternatives`);
        } catch (err) {
            return request<AssetSoftwareAlternative[]>(`/api/software/asset/${id}/alternatives`);
        }
    },

    // Threats
    getThreats(): Promise<Threat[]> {
        return request<Threat[]>("/api/threats");
    },

    // Risk
    getRiskOverview(): Promise<RiskOverviewPoint[]> {
        return request<RiskOverviewPoint[]>("/api/risk/overview");
    },

    previewRisk(body: RiskPreviewRequest): Promise<RiskPreviewResponse> {
        return request<RiskPreviewResponse>("/api/risk/preview", {
            method: "POST",
            body: JSON.stringify(body),
        });
    },

    // Software catalog
    getSoftware(): Promise<Software[]> {
        return request<Software[]>("/api/software");
    },

    getSoftwareCategories(): Promise<SoftwareCategory[]> {
        return request<SoftwareCategory[]>("/api/software/categories");
    },

    getRussianSoftware(): Promise<Software[]> {
        return request<Software[]>("/api/software/russian");
    },

    getCertifiedSoftware(): Promise<Software[]> {
        return request<Software[]>("/api/software/certified");
    },

    getSoftwareAlternatives(id: number): Promise<Software[]> {
        return request<Software[]>(`/api/software/${id}/alternatives`);
    },

    searchSoftware(params: { category?: string; russian?: boolean; certified?: boolean }): Promise<Software[]> {
        const query = new URLSearchParams();
        if (params.category) query.set("category", params.category);
        if (params.russian) query.set("russian", "true");
        if (params.certified) query.set("certified", "true");
        return request<Software[]>(`/api/software?${query.toString()}`);
    },
};
