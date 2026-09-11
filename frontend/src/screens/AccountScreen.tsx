import * as React from "react";
import { KeyRound, Activity, BookOpen, ChevronRight } from "lucide-react";
import type { AuthState } from "../app/auth";
import type { AssetPosition, PortfolioSummary } from "../app/portfolio";
import {
  createAPIKey,
  loadAPIKeys,
  revokeAPIKey,
  type APIKeySummary,
  type CreatedAPIKey,
} from "../app/apiKeys";
import { checkAuditAccess, loadAuditEvents, type AuditEvent } from "../app/audit";
import { Card } from "../components/ui/card";
import { GroupedCard } from "../components/cards/GroupedCard";
import { DestructiveActionRow } from "../components/cards/DestructiveActionRow";
import { ActionButton } from "../app/components";
import { OperationError, operationFailure, type OperationFailure } from "../components/feedback/OperationError";
import { FilePickerInput } from "../components/inputs/FilePickerInput";
import { confirmDestructive, confirmImport, createExport, getExportDownload, getLifecycleJob, previewDestructive, uploadImport, type DestructivePreview, type LifecycleJob } from "../app/lifecycle";

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
  const [keyListBusy, setKeyListBusy] = React.useState(online);
  const [keyListError, setKeyListError] = React.useState<OperationFailure | null>(null);
  const [keyActionError, setKeyActionError] = React.useState<OperationFailure | null>(null);
  const [copyError, setCopyError] = React.useState<OperationFailure | null>(null);
  const [copied, setCopied] = React.useState(false);
  const keyListGeneration = React.useRef(0);
  const copyGeneration = React.useRef(0);
  React.useEffect(() => () => { ++copyGeneration.current; }, []);
  const [auditAllowed, setAuditAllowed] = React.useState(false);
  const [auditEvents, setAuditEvents] = React.useState<AuditEvent[]>([]);
  const [auditCorrelationID, setAuditCorrelationID] = React.useState("");
  const [auditBusy, setAuditBusy] = React.useState(false);
  const [lifecycleJob, setLifecycleJob] = React.useState<LifecycleJob | null>(null);
  const [lifecycleBusy, setLifecycleBusy] = React.useState(false);
  const [lifecycleError, setLifecycleError] = React.useState<OperationFailure | null>(null);
  const [destructiveKind, setDestructiveKind] = React.useState<"reset"|"delete"|null>(null);
  const [destructivePreview, setDestructivePreview] = React.useState<DestructivePreview|null>(null);
  const [typedConfirmation, setTypedConfirmation] = React.useState("");

  React.useEffect(() => {
    const generation = ++keyListGeneration.current;
    if (!online || authState.status !== "authenticated") {
      setKeyListBusy(false);
      return;
    }
    setKeyListBusy(true);
    setKeyListError(null);
    void loadAPIKeys().then(keys => { if (generation === keyListGeneration.current) setAPIKeys(keys); })
      .catch(error => { if (generation === keyListGeneration.current) setKeyListError(operationFailure(error, "Không tải được danh sách API key.")); })
      .finally(() => { if (generation === keyListGeneration.current) setKeyListBusy(false); });
    return () => { ++keyListGeneration.current; };
  }, [authState.status, online]);

  async function refreshKeys() {
    const generation = ++keyListGeneration.current;
    setKeyListBusy(true);
    setKeyListError(null);
    try {
      const keys = await loadAPIKeys();
      if (generation === keyListGeneration.current) { setAPIKeys(keys); setKeyActionError(null); }
    }
    catch (error) { if (generation === keyListGeneration.current) setKeyListError(operationFailure(error, "Chưa cập nhật được danh sách API key. Kết quả thao tác đã xác nhận vẫn được giữ lại.")); }
    finally { if (generation === keyListGeneration.current) setKeyListBusy(false); }
  }

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
    if (!online || apiKeyBusy || keyListBusy || apiKeyName.trim() === "") return;
    setAPIKeyBusy(true);
    setKeyActionError(null);
    setCopyError(null);
    setCopied(false);
    try {
      const created = await createAPIKey(apiKeyName.trim());
      ++copyGeneration.current;
      setCopied(false);
      setCopyError(null);
      setCreatedKey(created);
      const { plaintext: _plaintext, ...summary } = created;
      setAPIKeys(current => [summary, ...current.filter(key => key.id !== summary.id)]);
      setAPIKeyName("AI Agent");
      await refreshKeys();
    } catch (error) {
      setKeyActionError(operationFailure(error, "Chưa xác nhận được việc tạo API key. Kiểm tra lại danh sách trước khi tạo lại."));
    } finally {
      setAPIKeyBusy(false);
    }
  }

  async function handleRevokeAPIKey(id: string) {
    if (!online || apiKeyBusy || keyListBusy) return;
    setAPIKeyBusy(true);
    setKeyActionError(null);
    try {
      await revokeAPIKey(id);
      setAPIKeys(current => current.map(key => key.id === id ? { ...key, revoked_at: new Date().toISOString() } : key));
      if (createdKey?.id === id) { ++copyGeneration.current; setCreatedKey(null); setCopied(false); setCopyError(null); }
      await refreshKeys();
    } catch (error) {
      setKeyActionError(operationFailure(error, "Chưa xác nhận được việc thu hồi API key. Tải lại danh sách để kiểm tra trạng thái."));
    } finally {
      setAPIKeyBusy(false);
    }
  }

  async function copyKey() {
    if (!createdKey) return;
    const generation = ++copyGeneration.current;
    setCopied(false);
    setCopyError(null);
    try {
      if (!navigator.clipboard?.writeText) throw new Error('clipboard unavailable');
      await navigator.clipboard.writeText(createdKey.plaintext);
      if (generation === copyGeneration.current) setCopied(true);
    } catch (error) {
      if (generation === copyGeneration.current) setCopyError(operationFailure(error, "Không sao chép được API key. Bạn có thể chọn và sao chép chuỗi key đang hiển thị."));
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

  async function runLifecycle(action:()=>Promise<LifecycleJob>){if(lifecycleBusy)return;setLifecycleBusy(true);setLifecycleError(null);try{setLifecycleJob(await action())}catch(error){setLifecycleError(operationFailure(error,"Thao tác dữ liệu chưa hoàn tất."))}finally{setLifecycleBusy(false)}}
  async function chooseImport(file:File){await runLifecycle(()=>uploadImport(file))}
  async function refreshLifecycle(){if(!lifecycleJob)return;await runLifecycle(()=>getLifecycleJob(lifecycleJob))}
  async function downloadExport(){if(!lifecycleJob)return;setLifecycleBusy(true);try{const url=await getExportDownload(lifecycleJob);window.location.assign(url)}catch(error){setLifecycleError(operationFailure(error,"Không tạo được liên kết tải xuống."))}finally{setLifecycleBusy(false)}}
  async function openDestructive(kind:"reset"|"delete"){setLifecycleBusy(true);setLifecycleError(null);try{setDestructiveKind(kind);setDestructivePreview(await previewDestructive(kind));setTypedConfirmation("")}catch(error){setLifecycleError(operationFailure(error,"Không tải được bản xem trước."))}finally{setLifecycleBusy(false)}}
  async function applyDestructive(){if(!destructiveKind||!destructivePreview)return;await runLifecycle(()=>confirmDestructive(destructiveKind,typedConfirmation,destructivePreview.preview_token));setDestructiveKind(null);setDestructivePreview(null);setTypedConfirmation("")}

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
                  {asset.summary.quantity} {asset.unit} · {formatVND(asset.summary.market_value_vnd ?? 0)}
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
          <ActionButton disabled={!online || keyListBusy || apiKeyBusy} onClick={() => void refreshKeys()}>
            Tải lại danh sách key
          </ActionButton>
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
          <ActionButton
            type="button"
            disabled={!online || apiKeyBusy || keyListBusy || apiKeyName.trim() === ""}
            onClick={() => void handleCreateAPIKey()}
          >
            {apiKeyBusy ? "Đang xử lý" : "Tạo key"}
          </ActionButton>
        </div>

        {createdKey ? (
          <div className="api-key-secret">
            <span>Chỉ hiển thị một lần</span>
            <code>{createdKey.plaintext}</code>
            <ActionButton
              type="button"
              onClick={() => void copyKey()}
            >
              Copy
            </ActionButton>
          </div>
        ) : null}

        <OperationError failure={keyActionError} />
        <OperationError failure={copyError} />
        {copied ? <p role="status">Đã sao chép API key.</p> : null}
        <OperationError failure={keyListError} />
        {keyListBusy ? <p role="status">Đang tải danh sách API key…</p> : null}
        {apiKeys.length === 0 && !keyListBusy && !keyListError ? (
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
                <ActionButton
                  className="danger-text"
                  type="button"
                  disabled={!online || apiKeyBusy || keyListBusy || Boolean(key.revoked_at)}
                  onClick={() => void handleRevokeAPIKey(key.id)}
                >
                  Revoke
                </ActionButton>
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

      <GroupedCard title="Nhập và xuất dữ liệu">
        <div className="content-stack">
          <FilePickerInput aria-label="Chọn CSV để nhập" accept="text/csv,.csv" disabled={!online||lifecycleBusy} onFileSelected={(file)=>void chooseImport(file)} />
          <ActionButton disabled={!online||lifecycleBusy} onClick={()=>void runLifecycle(createExport)}>Tạo bản xuất CSV</ActionButton>
          {lifecycleJob ? <div role="status"><strong>{lifecycleJob.kind}: {lifecycleJob.status}</strong>{lifecycleJob.error_code?<p>{lifecycleJob.error_code}</p>:null}<ActionButton disabled={lifecycleBusy} onClick={()=>void refreshLifecycle()}>Cập nhật trạng thái</ActionButton>{lifecycleJob.kind==="import"&&lifecycleJob.status==="awaiting_confirmation"&&lifecycleJob.result?.confirmable?<ActionButton disabled={lifecycleBusy} onClick={()=>void runLifecycle(()=>confirmImport(lifecycleJob))}>Xác nhận nhập</ActionButton>:null}{lifecycleJob.kind==="import"&&lifecycleJob.result?.errors?.map(error=><p key={error.row}>Dòng {error.row}: {error.message}</p>)}{lifecycleJob.kind==="export"&&lifecycleJob.status==="completed"?<ActionButton disabled={lifecycleBusy} onClick={()=>void downloadExport()}>Tải CSV</ActionButton>:null}</div>:null}
          <OperationError failure={lifecycleError} onRetry={lifecycleJob?()=>void refreshLifecycle():undefined} busy={lifecycleBusy} />
        </div>
      </GroupedCard>

      <GroupedCard title="Vùng nguy hiểm">
        <p>Xóa dữ liệu tài chính nhưng giữ tài khoản.</p>
        <DestructiveActionRow label="Xem trước đặt lại dữ liệu" onClick={()=>void openDestructive("reset")} />
        <p>Vô hiệu hóa truy cập ngay và xóa dữ liệu.</p>
        <DestructiveActionRow label="Xem trước xóa tài khoản" onClick={()=>void openDestructive("delete")} />
        {destructiveKind&&destructivePreview?<div role="dialog" aria-label={destructiveKind==="reset"?"Xác nhận đặt lại":"Xác nhận xóa tài khoản"}><p>{Object.values(destructivePreview.affected_counts).reduce((sum,value)=>sum+value,0)} bản ghi sẽ bị ảnh hưởng.</p><label>Nhập {destructiveKind.toUpperCase()}<input aria-label="Chuỗi xác nhận" value={typedConfirmation} onChange={event=>setTypedConfirmation(event.target.value)} /></label><ActionButton onClick={()=>{setDestructiveKind(null);setDestructivePreview(null)}}>Hủy</ActionButton><ActionButton disabled={typedConfirmation!==destructiveKind.toUpperCase()||lifecycleBusy} onClick={()=>void applyDestructive()}>Xác nhận</ActionButton></div>:null}
      </GroupedCard>

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
