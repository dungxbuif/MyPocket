import { apiRequest } from "./api";

export type PortfolioAsset = {
  id: string;
  owner_id: string;
  symbol: string;
  name: string;
  latest_price?: number;
  latest_price_at?: string;
};

export type PortfolioPosition = {
  asset_id: string;
  symbol?: string;
  name?: string;
  quantity: string;
  cost_basis: number;
  average_cost: number;
  realized_pnl: number;
  market_value?: number;
  unrealized_pnl?: number;
  latest_price?: number;
  latest_price_at?: string;
};

export type PortfolioTrade = {
  id: string;
  owner_id: string;
  asset_id: string;
  side: "buy" | "sell";
  quantity: string;
  unit_price: number;
  fee: number;
  occurred_at: string;
  note?: string;
};

export type PortfolioSummary = {
  assets: PortfolioAsset[];
  positions: PortfolioPosition[];
  trades: PortfolioTrade[];
  total_cost_basis: number;
  total_market_value?: number;
  total_unrealized_pnl?: number;
};

export const fetchPortfolioSummary = () => apiRequest<PortfolioSummary>("/api/v1/portfolio/summary");
export const fetchPortfolioAssets = () => apiRequest<PortfolioAsset[]>("/api/v1/portfolio/assets");
export const createPortfolioAsset = (input: { symbol: string; name: string }) => apiRequest<PortfolioAsset>("/api/v1/portfolio/assets", { method: "POST", body: JSON.stringify(input) });
export const updatePortfolioPrice = (assetID: string, latestPrice: number | null) => apiRequest<PortfolioAsset>(`/api/v1/portfolio/assets/${encodeURIComponent(assetID)}/price`, { method: "PATCH", body: JSON.stringify({ latest_price: latestPrice }) });
export const createPortfolioTrade = (assetID: string, input: { side: "buy" | "sell"; quantity: string; unit_price: number; fee: number; occurred_at: string; note?: string }) => apiRequest<PortfolioTrade>(`/api/v1/portfolio/assets/${encodeURIComponent(assetID)}/trades`, { method: "POST", body: JSON.stringify(input) });
export const fetchPortfolioTrades = (assetID: string) => apiRequest<PortfolioTrade[]>(`/api/v1/portfolio/assets/${encodeURIComponent(assetID)}/trades`);
