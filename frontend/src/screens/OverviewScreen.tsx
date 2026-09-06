import * as React from "react";
import { RefreshCw } from "lucide-react";
import type { WalletSummary, Dashboard, Report, InsiderReport, WalletDetail } from "../app/finance";
import { Card } from "../components/ui/card";
import { ComparisonBarChart } from "../components/charts/ComparisonBarChart";
import { TrendAreaLineChart } from "../components/charts/TrendAreaLineChart";

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
  onWalletClick?: (walletId: string) => void;
  formatVND: (val: number) => string;
  formatPercent: (val: number) => string;
  walletIcon: (type: string) => string;
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
  onWalletClick,
  formatVND,
  formatPercent,
  walletIcon,
}: OverviewScreenProps) {
  const visibleWallets = wallets ?? [];
  const [reportView, setReportView] = React.useState<"week" | "month">("month");

  // Mock bar data for period comparison
  const weeklyData = [
    { label: "Tuần trước", amount: 150000, amountFormatted: "150.000 đ" },
    { label: "Tuần này", amount: 250000, amountFormatted: "250.000 đ", isCurrent: true },
  ];

  const monthlyData = [
    { label: "Tháng trước", amount: 3200000, amountFormatted: "3.200.000 đ" },
    { label: "Tháng này", amount: 4800000, amountFormatted: "4.800.000 đ", isCurrent: true },
  ];

  const trendPoints = [
    { dateLabel: "01/08", actualAmount: 300000, baselineAmount: 400000 },
    { dateLabel: "08/08", actualAmount: 1200000, baselineAmount: 1100000 },
    { dateLabel: "15/08", actualAmount: 2100000, baselineAmount: 2200000 },
    { dateLabel: "23/08", actualAmount: 3500000, baselineAmount: 3100000 },
    { dateLabel: "31/08", actualAmount: 4800000, baselineAmount: 4200000 },
  ];

  return (
    <section className="content-stack">
      {/* Wallets Card */}
      <Card className="wallet-card">
        {!online ? <p className="offline-warning neutral">Dữ liệu đang hiển thị từ lần đồng bộ cuối</p> : null}
        <div className="section-title">
          <h2>Ví của tôi</h2>
          <button type="button" onClick={onManageWallets}>Xem tất cả</button>
        </div>
        {visibleWallets.length === 0 ? <p className="empty-state">Chưa có ví</p> : null}
        {visibleWallets.map((wallet) => (
          <div
            key={wallet.id}
            role="button"
            tabIndex={0}
            onClick={() => onWalletClick?.(wallet.id)}
            className="wallet-row"
          >
            <span className="wallet-row-icon">{walletIcon(wallet.type)}</span>
            <span className="wallet-row-name">{wallet.name}</span>
            <span className="wallet-row-amount">
              {privacyMasked ? "••••••" : formatVND(wallet.balance_vnd)}
            </span>
          </div>
        ))}
      </Card>

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
        <button
          type="button"
          aria-label="Làm mới Money Insider"
          className="section-heading-action"
          onClick={onRefreshInsider}
        >
          <RefreshCw size={20} />
        </button>
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
        <span className="section-heading-action">Xem báo cáo</span>
      </div>

      <section className="card report-card" aria-label="Báo cáo chi tiêu">
        <div className="segmented">
          <button
            type="button"
            className={reportView === "week" ? "active" : ""}
            onClick={() => setReportView("week")}
          >
            Tuần
          </button>
          <button
            type="button"
            className={reportView === "month" ? "active" : ""}
            onClick={() => setReportView("month")}
          >
            Tháng
          </button>
        </div>

        {/* Charts Presentation */}
        <ComparisonBarChart data={reportView === "week" ? weeklyData : monthlyData} height={120} />

        <div className="report-stats">
          <p>
            Tổng đã chi <strong className="expense">{formatVND(report?.summary.expense_vnd ?? 0)}</strong>
          </p>
          <p>
            Tổng thu <strong className="income">{formatVND(report?.summary.income_vnd ?? 0)}</strong>
          </p>
        </div>
      </section>
    </section>
  );
}
