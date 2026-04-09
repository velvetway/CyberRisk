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

async function request<T>(input: string, init?: RequestInit): Promise<T> {
    const res = await fetch(input, {
        headers: {
            "Content-Type": "application/json",
            ...(init?.headers || {}),
        },
        ...init,
    });

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
            // Для совместимости с новыми версиями API пробуем fallback-эндпоинт
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
