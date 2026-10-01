import { useEffect, useMemo, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { BaseSelect, BaseTextInput, FormField } from "../atoms/FormField";
import { SectionTitle } from "../atoms/SectionTitle";
import { StatusMessage } from "../atoms/StatusMessage";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";
import { formatVND } from "../utils/format";
import { createPortfolioAsset, createPortfolioTrade, fetchPortfolioSummary, type PortfolioAsset, type PortfolioSummary, updatePortfolioPrice } from "../../services/portfolio";

const today = new Date().toISOString().slice(0, 10);

export function PortfolioPanel({ refreshKey = 0 }: { refreshKey?: number }) {
  const [summary, setSummary] = useState<PortfolioSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [assetForm, setAssetForm] = useState({ symbol: "", name: "" });
  const [tradeForm, setTradeForm] = useState({ assetID: "", side: "buy" as "buy" | "sell", quantity: "", unitPrice: "", fee: "0", occurredAt: today });
  const [priceDrafts, setPriceDrafts] = useState<Record<string, string>>({});
  const [saving, setSaving] = useState(false);

  const load = async () => {
    try {
      setLoading(true);
      setSummary(await fetchPortfolioSummary());
      setError("");
    } catch (nextError) {
      setError(nextError instanceof Error ? nextError.message : "Không thể tải danh mục đầu tư.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void load(); }, [refreshKey]);

  const assets = summary?.assets ?? [];
  const selectedAsset = useMemo(() => assets.find(asset => asset.id === tradeForm.assetID), [assets, tradeForm.assetID]);

  const addAsset = async () => {
    if (saving || !assetForm.symbol.trim() || !assetForm.name.trim()) return;
    try {
      setSaving(true);
      const created = await createPortfolioAsset({ symbol: assetForm.symbol.trim(), name: assetForm.name.trim() });
      setAssetForm({ symbol: "", name: "" });
      setTradeForm(form => ({ ...form, assetID: form.assetID || created.id }));
      await load();
    } catch (nextError) {
      setError(nextError instanceof Error ? nextError.message : "Không thể thêm tài sản.");
    } finally {
      setSaving(false);
    }
  };

  const addTrade = async () => {
    if (saving || !tradeForm.assetID || !tradeForm.quantity || !tradeForm.unitPrice || !tradeForm.occurredAt) return;
    try {
      setSaving(true);
      await createPortfolioTrade(tradeForm.assetID, { side: tradeForm.side, quantity: tradeForm.quantity.trim(), unit_price: Number(tradeForm.unitPrice), fee: Number(tradeForm.fee || 0), occurred_at: new Date(`${tradeForm.occurredAt}T12:00:00`).toISOString() });
      setTradeForm(form => ({ ...form, quantity: "", unitPrice: "", fee: "0" }));
      await load();
    } catch (nextError) {
      setError(nextError instanceof Error ? nextError.message : "Không thể lưu giao dịch danh mục.");
    } finally {
      setSaving(false);
    }
  };

  const savePrice = async (asset: PortfolioAsset) => {
    if (saving) return;
    const value = priceDrafts[asset.id]?.trim() ?? "";
    try {
      setSaving(true);
      await updatePortfolioPrice(asset.id, value === "" ? null : Number(value));
      await load();
    } catch (nextError) {
      setError(nextError instanceof Error ? nextError.message : "Không thể cập nhật giá.");
    } finally {
      setSaving(false);
    }
  };

  if (loading && !summary) return <StatusMessage>Đang tải danh mục đầu tư...</StatusMessage>;
  if (!summary) return <div className="space-y-3"><StatusMessage tone="danger">{error || "Không thể tải danh mục đầu tư."}</StatusMessage><BaseButton variant="ghost" onClick={() => void load()}>Thử lại</BaseButton></div>;

  return <section className="space-y-3">
    <header className="grid grid-cols-[1fr_auto] items-center"><div><Text as="h1" size="lg" weight="bold">Danh mục đầu tư</Text><Text size="sm" tone="secondary">Theo dõi mua bán độc lập với số dư ví</Text></div><BaseButton variant="chip" size="sm" onClick={() => void load()}>Làm mới</BaseButton></header>
    {error ? <StatusMessage tone="danger">{error}</StatusMessage> : null}
    <SurfaceCard padding="md"><SectionTitle title="Tổng danh mục" /><div className="mt-3 grid grid-cols-2 gap-2"><Metric label="Giá vốn" value={summary.total_cost_basis} /><Metric label="Giá trị hiện tại" value={summary.total_market_value} unknown={summary.total_market_value == null} /><Metric label="Lãi/lỗ chưa thực hiện" value={summary.total_unrealized_pnl} unknown={summary.total_unrealized_pnl == null} /></div></SurfaceCard>
    <SurfaceCard padding="md"><SectionTitle title="Thêm tài sản" /><div className="mt-3 grid gap-3 sm:grid-cols-2"><FormField label="Mã"><BaseTextInput value={assetForm.symbol} onChange={event => setAssetForm({ ...assetForm, symbol: event.target.value })} placeholder="VN30F" /></FormField><FormField label="Tên"><BaseTextInput value={assetForm.name} onChange={event => setAssetForm({ ...assetForm, name: event.target.value })} placeholder="Quỹ chỉ số" /></FormField></div><BaseButton className="mt-3 w-full" disabled={saving || !assetForm.symbol.trim() || !assetForm.name.trim()} loading={saving} onClick={() => void addAsset()}>Thêm tài sản</BaseButton></SurfaceCard>
    <SurfaceCard padding="md"><SectionTitle title="Ghi giao dịch" /><div className="mt-3 space-y-3"><FormField label="Tài sản"><BaseSelect value={tradeForm.assetID} onChange={event => setTradeForm({ ...tradeForm, assetID: event.target.value })}><option value="">Chọn tài sản</option>{assets.map(asset => <option key={asset.id} value={asset.id}>{asset.symbol} · {asset.name}</option>)}</BaseSelect></FormField><div className="grid grid-cols-2 gap-3"><FormField label="Loại"><BaseSelect value={tradeForm.side} onChange={event => setTradeForm({ ...tradeForm, side: event.target.value as "buy" | "sell" })}><option value="buy">Mua</option><option value="sell">Bán</option></BaseSelect></FormField><FormField label="Khối lượng"><BaseTextInput value={tradeForm.quantity} onChange={event => setTradeForm({ ...tradeForm, quantity: event.target.value })} inputMode="decimal" placeholder="1.5" /></FormField><FormField label="Giá đơn vị"><BaseTextInput value={tradeForm.unitPrice} onChange={event => setTradeForm({ ...tradeForm, unitPrice: event.target.value })} inputMode="numeric" placeholder="100000" /></FormField><FormField label="Phí"><BaseTextInput value={tradeForm.fee} onChange={event => setTradeForm({ ...tradeForm, fee: event.target.value })} inputMode="numeric" /></FormField><FormField label="Ngày"><BaseTextInput type="date" value={tradeForm.occurredAt} onChange={event => setTradeForm({ ...tradeForm, occurredAt: event.target.value })} /></FormField></div><BaseButton className="w-full" disabled={saving || !selectedAsset || !tradeForm.quantity || !tradeForm.unitPrice} loading={saving} onClick={() => void addTrade()}>Lưu giao dịch</BaseButton></div></SurfaceCard>
    {assets.length === 0 ? <StatusMessage variant="plain">Chưa có tài sản. Thêm tài sản để bắt đầu ghi lịch sử mua bán.</StatusMessage> : <div className="space-y-3">{assets.map(asset => <AssetCard key={asset.id} asset={asset} position={summary.positions.find(position => position.asset_id === asset.id)} priceDraft={priceDrafts[asset.id] ?? (asset.latest_price != null ? String(asset.latest_price) : "")} onPriceDraft={value => setPriceDrafts(current => ({ ...current, [asset.id]: value }))} onSavePrice={() => void savePrice(asset)} />)}</div>}
  </section>;
}

function Metric({ label, value, unknown = false }: { label: string; value?: number; unknown?: boolean }) { return <SurfaceCard padding="sm" tone="muted" elevation="flat"><Text size="xs" tone="secondary">{label}</Text><Text numeric weight="bold" className="mt-1">{unknown ? "Chưa có giá" : formatVND(value ?? 0)}</Text></SurfaceCard>; }

function AssetCard({ asset, position, priceDraft, onPriceDraft, onSavePrice }: { asset: PortfolioAsset; position?: PortfolioSummary["positions"][number]; priceDraft: string; onPriceDraft: (value: string) => void; onSavePrice: () => void }) {
  return <SurfaceCard padding="md"><div className="flex items-start justify-between gap-3"><div><Text weight="bold">{asset.symbol}</Text><Text size="sm" tone="secondary">{asset.name}</Text></div><Text numeric size="sm" weight="semibold">{position?.quantity ?? "0"}</Text></div><div className="mt-3 grid grid-cols-2 gap-2"><Metric label="Giá vốn" value={position?.cost_basis ?? 0} /><Metric label="Bình quân" value={position?.average_cost ?? 0} /><Metric label="Giá trị" value={position?.market_value} unknown={position?.market_value == null} /><Metric label="Lãi/lỗ" value={position?.unrealized_pnl} unknown={position?.unrealized_pnl == null} /></div><div className="mt-3 flex items-end gap-2"><FormField label="Giá hiện tại"><BaseTextInput value={priceDraft} onChange={event => onPriceDraft(event.target.value)} inputMode="numeric" placeholder="Để trống nếu chưa biết" /></FormField><BaseButton variant="secondary" size="sm" onClick={onSavePrice}>Cập nhật</BaseButton></div></SurfaceCard>;
}
