import * as React from "react";
import { RefreshCw } from "lucide-react";
import type { WalletSummary } from "../app/finance";
import type { Dashboard, Report, InsiderReport, WalletDetail } from "../app/analytics";
import { Card } from "../components/ui/card";
import { Button } from "../components/ui/button";
import { ComparisonBarChart } from "../components/charts/ComparisonBarChart";
import type { BarItem } from "../components/charts/ComparisonBarChart";
import { WalletRow } from "../components/finance/WalletRow";

function hoChiMinhDayKey(date: Date): string {
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone: "Asia/Ho_Chi_Minh",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(date);
  const value = (type: Intl.DateTimeFormatPartTypes) => parts.find((part) => part.type === type)?.value ?? "";
  return `${value("year")}-${value("month")}-${value("day")}`;
}

export function buildDailyExpenseBars(report: Report | null, now = new Date()): BarItem[] {
  const today = hoChiMinhDayKey(now);
  return (report?.daily ?? []).filter((item) => item.date <= today).slice(-7).map((item) => ({
    label: item.date.slice(8, 10) + "/" + item.date.slice(5, 7),
    amount: item.expense_vnd,
    amountFormatted: `${new Intl.NumberFormat("vi-VN").format(item.expense_vnd)} đ`,
    ...(item.date === today ? { isCurrent: true } : {}),
  }));
}

export interface OverviewScreenProps {
  online: boolean;
  wallets: WalletSummary[] | null;
  dashboard: Dashboard | null;
  report: Report | null;
  insider: InsiderReport | null;
  privacyMasked: boolean;
  walletDetail: WalletDetail | null;
  onCloseWalletDetail: () => void;
  onManageWallets: () => void;
  onRefreshInsider: () => void;
  onViewReports?: () => void;
  onWalletClick?: (walletId: string) => void;
  formatVND: (val: number) => string;
  formatPercent: (val: number) => string;
}

export function OverviewScreen({
  online,
  wallets,
  dashboard,
  report,
  insider,
  privacyMasked,
  walletDetail,
  onCloseWalletDetail,
  onManageWallets,
  onRefreshInsider,
  onViewReports,
  onWalletClick,
  formatVND,
  formatPercent,
}: OverviewScreenProps) {
  const visibleWallets = wallets ?? [];
  const expenseBars = buildDailyExpenseBars(report);

  return (
    <section className="content-stack">
      {/* Wallets Card */}
      <Card className="wallet-card">
        {!online ? <p className="offline-warning neutral">Dữ liệu đang hiển thị từ lần đồng bộ cuối</p> : null}
        <div className="section-title">
          <h2>Ví của tôi</h2>
          <Button variant="ghost" size="sm" onClick={onManageWallets}>Xem tất cả</Button>
        </div>
        {visibleWallets.length === 0 ? <p className="empty-state">Chưa có ví</p> : null}
        {visibleWallets.map((wallet) => (
          <WalletRow
            key={wallet.id}
            id={wallet.id}
            name={wallet.name}
            type={wallet.type}
            balance={wallet.balance_vnd}
            privacyMasked={privacyMasked}
            onClick={() => onWalletClick?.(wallet.id)}
          />
        ))}
      </Card>

      {walletDetail ? <WalletDetailCard detail={walletDetail} privacyMasked={privacyMasked} formatVND={formatVND} onClose={onCloseWalletDetail} /> : null}

      {/* Investment Position Card */}
      {dashboard?.investment_market_value_vnd || dashboard?.missing_asset_price_count ? (
        <Card className="wallet-card">
          <div className="section-title">
            <h2>Tài sản đầu tư</h2>
          </div>
          <div className="wallet-row">
            <span className="wallet-row-icon">◆</span>
            <span className="wallet-row-name">Giá trị đầu tư</span>
            <span className="wallet-row-amount">
              {privacyMasked ? "••••••" : formatVND(dashboard.investment_market_value_vnd ?? 0)}
            </span>
          </div>
          <div className="wallet-row">
            <span className="wallet-row-icon">₫</span>
            <span className="wallet-row-name">Tổng tài sản</span>
            <span className="wallet-row-amount">
              {privacyMasked ? "••••••" : formatVND(dashboard.combined_net_worth_vnd ?? 0)}
            </span>
          </div>
          {dashboard.missing_asset_price_count ? (
            <p className="notification-status">{dashboard.missing_asset_price_count} tài sản chưa có giá hiện tại</p>
          ) : null}
        </Card>
      ) : null}

      {/* Money Insider */}
      <div className="section-heading">
        <h2>Money Insider</h2>
        <Button
          type="button"
          aria-label="Làm mới Money Insider"
          className="section-heading-action"
          onClick={onRefreshInsider}
        >
          <RefreshCw size={20} />
        </Button>
      </div>

      <section className="card insider-card">
        {!insider ? (
          <p className="empty-state">Đang tổng hợp chi tiêu...</p>
        ) : !insider.selected_category ? (
          <p className="empty-state">Chưa đủ dữ liệu chi tiêu tháng này</p>
        ) : (
          <>
            <h2>
              {insider.selected_category.name}{" "}
              <span className="info" title="Danh mục có nhiều giao dịch nhất">i</span>
            </h2>
            <p className="muted">
              Tổng đã chi{" "}
              <strong className="expense">
                {privacyMasked ? "••••••" : formatVND(insider.spent_vnd)}
              </strong>
            </p>
            <div className="insider-grid">
              <div>
                <h3>Tháng này</h3>
                <p className="muted">Trung bình trong {insider.elapsed_days} ngày</p>
                <strong>
                  {privacyMasked ? "••••••" : formatVND(insider.average_daily_vnd)}
                  <span>/ngày</span>
                </strong>
              </div>
              <div
                className="ring"
                aria-label={
                  insider.not_comparable
                    ? "Chưa đủ dữ liệu tháng trước"
                    : `${formatPercent(insider.change_percent ?? 0)} so với tháng trước`
                }
              >
                {insider.not_comparable ? "—" : formatPercent(insider.change_percent ?? 0)}
              </div>
            </div>
            <p className="insider-comparison">
              {insider.not_comparable
                ? "Chưa đủ dữ liệu tháng trước để so sánh"
                : `${formatPercent(insider.change_percent ?? 0)} so với trung bình mỗi ngày tháng trước`}
            </p>
          </>
        )}
      </section>

      {/* Monthly Report Card */}
      <div className="section-heading">
        <h2>Báo cáo tháng này</h2>
        <Button variant="ghost" size="sm" onClick={onViewReports}>Xem báo cáo</Button>
      </div>

      <section className="card report-card" aria-label="Báo cáo chi tiêu">
        {!report ? <p className="empty-state">Chưa tải được dữ liệu báo cáo. Mở báo cáo để thử lại.</p> : privacyMasked ? <p className="empty-state">Số liệu đang được ẩn.</p> : expenseBars.length > 0 ? (
          <ComparisonBarChart data={expenseBars} height={120} />
        ) : (
          <p className="empty-state">Chưa có dữ liệu chi tiêu để hiển thị biểu đồ</p>
        )}

        {report ? <div className="report-stats">
          <p>
            Tổng đã chi <strong className="expense">{privacyMasked ? "••••••" : formatVND(report.summary.expense_vnd)}</strong>
          </p>
          <p>
            Tổng thu <strong className="income">{privacyMasked ? "••••••" : formatVND(report.summary.income_vnd)}</strong>
          </p>
        </div> : null}
      </section>
    </section>
  );
}

function WalletDetailCard({
  detail,
  privacyMasked,
  formatVND,
  onClose,
}: {
  detail: WalletDetail;
  privacyMasked: boolean;
  formatVND: (value: number) => string;
  onClose: () => void;
}) {
  return (
    <Card className="wallet-detail" aria-label={`Chi tiết ${detail.wallet.name}`}>
      <div className="section-title">
        <h2>{detail.wallet.name}</h2>
        <Button variant="ghost" size="sm" onClick={onClose}>Đóng</Button>
      </div>
      <p>Trạng thái: {detail.wallet.include_in_total ? "Đang tính tổng" : "Không tính tổng"}</p>
      <strong>{privacyMasked ? "••••••" : formatVND(detail.wallet.balance_vnd)}</strong>
      {detail.transactions.length === 0 ? <p className="empty-state">Chưa có giao dịch</p> : detail.transactions.map((transaction) => (
        <div className="search-result" key={transaction.id}>
          <span>{transaction.note || "Giao dịch"}</span>
          <b>{privacyMasked ? "••••••" : formatVND(transaction.amount_vnd)}</b>
        </div>
      ))}
    </Card>
  );
}
