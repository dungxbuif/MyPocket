import { apiFetch } from "./apiClient";
import { archiveOfflineAsset, enqueueMutation, readOfflineSnapshot, upsertOfflineAsset } from "../offline/db";
import { readUserDataCache, writeUserDataCache } from "./userDataCache";

export type AssetType = "gold" | "stock" | "crypto" | "foreign_currency" | "other";
export type PricingMode = "manual" | "automatic";
export type TradeSide = "buy" | "sell";

export type AssetSummary = {
  quantity: string;
  cost_basis_vnd: number;
  realized_pnl_vnd: number;
  current_unit_price_vnd?: number | null;
  market_value_vnd?: number | null;
  unrealized_pnl_vnd?: number | null;
  unrealized_pnl_percent?: string | null;
  unrealized_not_comparable: boolean;
  valuation_status: "current" | "missing_price" | string;
  priced_at?: string;
  price_source?: string;
};

export type PricePoint = {
  id: string;
  asset_id: string;
  unit_price_vnd: number;
  priced_at: string;
  source: string;
  provider_quote_id?: string;
  created_at: string;
};

export type AssetTrade = {
  id: string;
  asset_id: string;
  side: TradeSide;
  quantity: string;
  unit_price_vnd: number;
  fee_vnd: number;
  occurred_at: string;
  quantity_after: string;
  cost_basis_after_vnd: number;
  realized_pnl_vnd: number;
  note: string;
  version: number;
};

export type AssetPosition = {
  id: string;
  type: AssetType;
  symbol: string;
  exchange: string;
  name: string;
  unit: string;
  reporting_currency: "VND";
  pricing_mode: PricingMode;
  provider_key?: string;
  provider_symbol?: string;
  include_in_net_worth: boolean;
  archived_at?: string;
  version: number;
  summary: AssetSummary;
  latest_price?: PricePoint;
  trades?: AssetTrade[];
  price_history?: PricePoint[];
};

export type PortfolioSummary = {
  investment_market_value_vnd: number;
  missing_price_count: number;
  included_position_count: number;
  position_count: number;
};

export type AssetCreateInput = {
  type: AssetType;
  symbol?: string;
  exchange?: string;
  name: string;
  unit: string;
  pricing_mode?: PricingMode;
  provider_key?: string;
  provider_symbol?: string;
  include_in_net_worth?: boolean;
};

export async function loadAssets(userID: string) {
  try {
    const response = await apiFetch<{ assets: AssetPosition[] }>("/api/v1/assets");
    writeUserDataCache(userID, "assets", response.assets);
    return response.assets;
  } catch (error) {
    if (!(error instanceof TypeError)) throw error;
    const cached = readUserDataCache<AssetPosition[]>(userID, "assets");
    if (cached) return cached;
    throw error;
  }
}

export async function loadPortfolioSummary(userID: string) {
  try {
    const response = await apiFetch<{ summary: PortfolioSummary }>("/api/v1/portfolio/summary");
    writeUserDataCache(userID, "portfolio-summary", response.summary);
    return response.summary;
  } catch (error) {
    if (!(error instanceof TypeError)) throw error;
    const cached = readUserDataCache<PortfolioSummary>(userID, "portfolio-summary");
    if (cached) return cached;
    throw error;
  }
}

export async function createAsset(input: AssetCreateInput) {
  const payload: AssetCreateInput = { pricing_mode: "manual", include_in_net_worth: true, ...input };
  return withOfflineFallback(
    () => apiFetch<{ asset: AssetPosition }>("/api/v1/assets", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }).then((response) => response.asset),
    async () => {
      const asset = optimisticAsset(payload);
      await enqueueMutation({ entity_type: "asset", entity_id: asset.id, operation: "create", base_version: 0, payload });
      await upsertOfflineAsset(asset);
      return asset;
    },
  );
}

export async function addAssetTrade(assetID: string, input: { side: TradeSide; quantity: string; unit_price_vnd: number; fee_vnd?: number; occurred_at: string; note?: string; base_version?: number }) {
  return withOfflineFallback(
    () => apiFetch<{ asset: AssetPosition }>(`/api/v1/assets/${encodeURIComponent(assetID)}/trades`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(input),
    }).then((response) => response.asset),
    async () => {
      await enqueueMutation({ entity_type: "asset", entity_id: assetID, operation: "add_trade", base_version: input.base_version ?? 0, payload: input as unknown as Record<string, unknown> });
      return cachedAsset(assetID);
    },
  );
}

export async function addAssetPrice(assetID: string, input: { unit_price_vnd: number; priced_at: string; source?: string; base_version?: number }) {
  const payload = { source: "manual", ...input };
  return withOfflineFallback(
    () => apiFetch<{ asset: AssetPosition }>(`/api/v1/assets/${encodeURIComponent(assetID)}/prices`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }).then((response) => response.asset),
    async () => {
      await enqueueMutation({ entity_type: "asset", entity_id: assetID, operation: "add_price", base_version: input.base_version ?? 0, payload });
      return cachedAsset(assetID);
    },
  );
}

export async function archiveAsset(assetID: string, baseVersion = 0) {
  await withOfflineFallback(
    () => apiFetch(`/api/v1/assets/${encodeURIComponent(assetID)}/archive`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ base_version: baseVersion }),
    }),
    async () => {
      await enqueueMutation({ entity_type: "asset", entity_id: assetID, operation: "archive", base_version: baseVersion, payload: {} });
      await archiveOfflineAsset(assetID, baseVersion);
    },
  );
}

function optimisticAsset(input: AssetCreateInput): AssetPosition {
  const now = new Date().toISOString();
  return {
    id: globalThis.crypto?.randomUUID?.() ?? `asset-${Date.now()}`,
    type: input.type,
    symbol: input.symbol?.toUpperCase().trim() ?? "",
    exchange: input.exchange?.toUpperCase().trim() ?? "",
    name: input.name.trim(),
    unit: input.unit,
    reporting_currency: "VND",
    pricing_mode: input.pricing_mode ?? "manual",
    include_in_net_worth: input.include_in_net_worth ?? true,
    version: 0,
    summary: {
      quantity: "0",
      cost_basis_vnd: 0,
      realized_pnl_vnd: 0,
      current_unit_price_vnd: null,
      market_value_vnd: null,
      unrealized_pnl_vnd: null,
      unrealized_pnl_percent: null,
      unrealized_not_comparable: true,
      valuation_status: "missing_price",
      priced_at: now,
      price_source: "offline",
    },
  };
}

async function cachedAsset(assetID: string) {
  const snapshot = await readOfflineSnapshot().catch(() => null);
  const offlineAsset = snapshot?.assets.find((item) => item.id === assetID);
  if (offlineAsset) return offlineAsset;
  throw new Error("Asset mutation queued offline");
}

async function withOfflineFallback<T>(onlineAction: () => Promise<T>, offlineAction: () => Promise<T>) {
  if (typeof navigator !== "undefined" && !navigator.onLine) {
    return offlineAction();
  }
  try {
    return await onlineAction();
  } catch (error) {
    if (isNetworkFailure(error)) return offlineAction();
    throw error;
  }
}

function isNetworkFailure(error: unknown) {
  return error instanceof TypeError || (error instanceof Error && /Failed to fetch|NetworkError|offline/i.test(error.message));
}
