import * as React from "react";
import { KeyRound, Activity, BookOpen, ChevronRight } from "lucide-react";
import type { AuthState } from "../app/auth";
import type { AssetPosition, PortfolioSummary } from "../app/portfolio";
import {
  checkAuditAccess,
  createAPIKey,
  loadAPIKeys,
  loadAuditEvents,
  revokeAPIKey,
  type APIKeySummary,
  type AuditEvent,
  type CreatedAPIKey,
} from "../app/apiKeys";
import { Card } from "../components/ui/card";
import { GroupedCard } from "../components/cards/GroupedCard";
import { DestructiveActionRow } from "../components/cards/DestructiveActionRow";

export interface AccountScreenProps {
  authState: AuthState;
  assets: AssetPosition[];
  portfolioSummary: PortfolioSummary | null;
  privacyMasked: boolean;
  online: boolean;
  onCreateAsset: () => void;
  onAssetChanged: (asset: AssetPosition) => void;
  onAssetArchived: (assetID: string) => void;
  onLogout: () => void;
  formatVND: (val: number) => string;
}

export function AccountScreen({
  authState,
  assets,
  portfolioSummary,
  privacyMasked,
  online,
  onCreateAsset,
  onAssetChanged,
  onAssetArchived,
  onLogout,
  formatVND,
}: AccountScreenProps) {
  const email = authState.status === "authenticated" ? authState.user.email : "Chưa đăng nhập";
  const displayName =
    authState.status === "authenticated"
      ? authState.user.display_name || authState.user.email
      : "Tài khoản MyPocket";

  const [apiKeys, setAPIKeys] = React.useState<APIKeySummary[]>([]);
  const [apiKeyName, setAPIKeyName] = React.useState("AI Agent");
  const [createdKey, setCreatedKey] = React.useState<CreatedAPIKey | null>(null);
  const [apiKeyBusy, setAPIKeyBusy] = React.useState(false);
  const [auditAllowed, setAuditAllowed] = React.useState(false);
  const [auditEvents, setAuditEvents] = React.useState<AuditEvent[]>([]);
  const [auditCorrelationID, setAuditCorrelationID] = React.useState("");
  const [auditBusy, setAuditBusy] = React.useState(false);

  React.useEffect(() => {
    if (!online || authState.status !== "authenticated") return;
    void loadAPIKeys().then(setAPIKeys).catch(() => undefined);
  }, [authState.status, online]);

  React.useEffect(() => {
    if (!online || authState.status !== "authenticated") {
      setAuditAllowed(false);
      setAuditEvents([]);
      return;
    }
    let active = true;
    setAuditBusy(true);
    void checkAuditAccess()
      .then((result) => {
        if (!active) return;
        setAuditAllowed(result.allowed);
        if (result.allowed)
          return loadAuditEvents().then((events) => {
            if (active) setAuditEvents(events);
          });
        return undefined;
      })
      .catch(() => {
        if (active) setAuditAllowed(false);
      })
      .finally(() => {
        if (active) setAuditBusy(false);
      });
    return () => {
      active = false;
    };
  }, [authState.status, online]);

  async function handleCreateAPIKey() {
    if (!online || apiKeyName.trim() === "") return;
    setAPIKeyBusy(true);
    try {
      const created = await createAPIKey(apiKeyName.trim());
      setCreatedKey(created);
      setAPIKeyName("AI Agent");
      const updated = await loadAPIKeys();
      setAPIKeys(updated);
    } catch {
      // Ignored for now
    } finally {
      setAPIKeyBusy(false);
    }
  }

  async function handleRevokeAPIKey(id: string) {
    if (!online || apiKeyBusy) return;
    setAPIKeyBusy(true);
    try {
      await revokeAPIKey(id);
      const updated = await loadAPIKeys();
      setAPIKeys(updated);
    } catch {
      // Ignored for now
    } finally {
      setAPIKeyBusy(false);
    }
  }

  async function refreshAuditEvents() {
    if (!online || auditBusy) return;
    setAuditBusy(true);
    try {
      const events = await loadAuditEvents(auditCorrelationID.trim() || undefined);
      setAuditEvents(events);
    } catch {
      // Ignored for now
    } finally {
      setAuditBusy(false);
    }
  }

  return (
    <section className="content-stack">
      {/* Profile Card */}
      <section className="card user-profile-card flex flex-col items-center text-center p-5 rounded-[28px] bg-white shadow-xs">
        <div className="w-16 h-16 rounded-full bg-[#ff8800] text-white flex items-center justify-center font-bold text-2xl mb-2 shadow-xs">
          {displayName.slice(0, 1).toUpperCase()}
        </div>
        <span className="px-3 py-0.5 rounded-full bg-[#fff3e0] text-[#ff8800] text-[11px] font-bold tracking-wide uppercase mb-1">
          TÀI KHOẢN PREMIUM
        </span>
        <h2 className="text-lg font-bold text-[#111111]">{displayName}</h2>
        <p className="text-xs text-[#8e8e93]">{email}</p>
      </section>

      {/* Assets / Portfolio Card */}
      <section className="card list-card">
        <div className="section-title">
          <h2>Tài sản danh mục</h2>
          <button type="button" disabled={!online} onClick={onCreateAsset}>
            Thêm vị thế
          </button>
        </div>
        {portfolioSummary?.missing_price_count ? (
          <p className="notification-status">
            {portfolioSummary.missing_price_count} tài sản chưa có giá hiện tại
          </p>
        ) : null}
        {assets.length === 0 ? (
          <p className="empty-state">Chưa có tài sản</p>
        ) : (
          assets.map((asset) => (
            <div key={asset.id} className="planning-row">
              <span className="category-dot" />
              <div className="flex-1">
                <strong>{asset.symbol}</strong>
                <p>
                  {asset.units} {asset.asset_class} · {formatVND(asset.market_value_vnd)}
                </p>
              </div>
            </div>
          ))
        )}
      </section>

      {/* Devices Section */}
      <section className="card list-card">
        <div className="transaction-row">
          <div className="transaction-main">
            <span className="transaction-icon">📱</span>
            <div>
              <strong>iPhone</strong>
              <p className="text-[#2dbd4f]">Thiết bị này</p>
            </div>
          </div>
        </div>
      </section>

      {/* API Keys Card */}
      <section className="card list-card api-key-card">
        <div className="section-title">
          <h2>API keys</h2>
          <KeyRound size={18} />
        </div>
        <div className="manager-form api-key-form">
          <input
            aria-label="Tên API key"
            value={apiKeyName}
            onChange={(event) => setAPIKeyName(event.target.value)}
            placeholder="Tên key"
            disabled={!online || apiKeyBusy}
          />
          <button
            type="button"
            disabled={!online || apiKeyBusy || apiKeyName.trim() === ""}
            onClick={() => void handleCreateAPIKey()}
          >
            {apiKeyBusy ? "Đang xử lý" : "Tạo key"}
          </button>
        </div>

        {createdKey ? (
          <div className="api-key-secret">
            <span>Chỉ hiển thị một lần</span>
            <code>{createdKey.plaintext}</code>
            <button
              type="button"
              onClick={() => void navigator.clipboard?.writeText(createdKey.plaintext)}
            >
              Copy
            </button>
          </div>
        ) : null}

        {apiKeys.length === 0 ? (
          <p className="empty-state">{online ? "Chưa có API key" : "Cần online để quản lý API key"}</p>
        ) : (
          <div className="api-key-list">
            {apiKeys.map((key) => (
              <div className="api-key-row" key={key.id}>
                <div>
                  <strong>{key.name}</strong>
                  <p>
                    {key.key_prefix}... ·{" "}
                    {key.revoked_at
                      ? "Đã revoke"
                      : key.last_used_at
                      ? `Dùng ${new Date(key.last_used_at).toLocaleDateString("vi-VN")}`
                      : "Chưa dùng"}
                  </p>
                </div>
                <button
                  className="danger-text"
                  type="button"
                  disabled={!online || apiKeyBusy || Boolean(key.revoked_at)}
                  onClick={() => void handleRevokeAPIKey(key.id)}
                >
                  Revoke
                </button>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Audit Log Card */}
      {auditAllowed ? (
        <section className="card list-card audit-log-card">
          <div className="section-title">
            <h2>Nhật ký hệ thống</h2>
            <Activity size={18} />
          </div>
          <div className="manager-form audit-log-form">
            <input
              aria-label="Correlation ID"
              value={auditCorrelationID}
              onChange={(event) => setAuditCorrelationID(event.target.value)}
              placeholder="Correlation ID"
            />
            <button
              type="button"
              disabled={auditBusy || !online}
              onClick={() => void refreshAuditEvents()}
            >
              {auditBusy ? "Đang tải" : "Làm mới"}
            </button>
          </div>
          {auditEvents.length === 0 ? (
            <p className="empty-state">Không có sự kiện phù hợp</p>
          ) : (
            <div className="audit-log-list">
              {auditEvents.map((event) => (
                <article className="audit-log-row" key={event.id}>
                  <div>
                    <strong>{event.action}</strong>
                    <p>
                      {event.request_method ?? ""} {event.request_path ?? ""}
                    </p>
                  </div>
                  <span className={`audit-severity ${event.severity}`}>{event.outcome}</span>
                  <time dateTime={event.occurred_at}>
                    {new Date(event.occurred_at).toLocaleString("vi-VN")}
                  </time>
                  <code>{event.correlation_id}</code>
                </article>
              ))}
            </div>
          )}
        </section>
      ) : null}

      {auditBusy && !auditAllowed ? (
        <p className="notification-status">Đang kiểm tra quyền nhật ký…</p>
      ) : null}

      {/* Documentation Link */}
      <section className="card list-card">
        <a className="planning-row" href="/docs/" target="_blank" rel="noopener noreferrer">
          <BookOpen size={24} />
          <div>
            <strong>Tài liệu</strong>
            <p>Sơ đồ CSDL và tài liệu API</p>
          </div>
          <ChevronRight size={22} />
        </a>
      </section>

      {/* Logout Action */}
      <button className="wide-pill destructive" type="button" onClick={onLogout}>
        Đăng xuất
      </button>

      {/* Footer Version Info */}
      <div className="text-center py-4 text-xs text-[#8e8e93]">
        <span>Phiên bản 8.75.0 (Build 2026.09)</span>
      </div>
    </section>
  );
}
