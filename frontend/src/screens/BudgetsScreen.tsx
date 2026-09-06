import * as React from "react";
import { ChevronRight } from "lucide-react";
import type {
  BudgetProgress,
  CategorySummary,
  EventSummary,
  ObligationSummary,
  RecurringSchedule,
  Transaction,
  TransactionDraft,
  WalletSummary,
} from "../app/finance";
import { Card } from "../components/ui/card";
import { GaugeArcSummary } from "../components/charts/GaugeArcSummary";
import { NoticeBanner } from "../components/feedback/NoticeBanner";
import { WalletFilterChip } from "../components/navigation/WalletFilterChip";

export interface BudgetsScreenProps {
  budgets: BudgetProgress[] | null;
  events: EventSummary[];
  obligations: ObligationSummary[];
  schedules: RecurringSchedule[];
  drafts: TransactionDraft[];
  categories: CategorySummary[];
  wallets: WalletSummary[];
  transactions: Transaction[];
  online: boolean;
  onCreate: () => void;
  onCreateEvent: () => void;
  onCreateObligation: () => void;
  onCreateSchedule: () => void;
  onEdit: (budget: BudgetProgress) => void;
  onEditEvent: (event: EventSummary) => void;
  onEditObligation: (obligation: ObligationSummary) => void;
  onEditSchedule: (schedule: RecurringSchedule) => void;
  formatVND: (val: number) => string;
  formatDate: (val: string) => string;
  obligationDirectionLabel: (dir: string) => string;
  recurrenceLabel: (freq: string) => string;
  transactionTypeLabel: (type: string) => string;
}

export function BudgetsScreen({
  budgets,
  events,
  obligations,
  schedules,
  drafts,
  categories,
  wallets,
  transactions,
  online,
  onCreate,
  onCreateEvent,
  onCreateObligation,
  onCreateSchedule,
  onEdit,
  onEditEvent,
  onEditObligation,
  onEditSchedule,
  formatVND,
  formatDate,
  obligationDirectionLabel,
  recurrenceLabel,
  transactionTypeLabel,
}: BudgetsScreenProps) {
  const [showNotice, setShowNotice] = React.useState(true);
  const rows = budgets ?? [];
  const totalBudget = rows.reduce((total, item) => total + item.budget.amount_vnd, 0);
  const totalSpent = rows.reduce((total, item) => total + item.spent_vnd, 0);
  const daysLeft = rows[0]
    ? Math.max(0, Math.ceil((new Date(rows[0].period_end).getTime() - Date.now()) / 86400000))
    : 12;

  const availableAmount = Math.max(0, totalBudget - totalSpent);

  return (
    <section className="content-stack">
      {/* Sub-header */}
      <div className="sub-header flex items-center justify-between">
        <div className="flex items-center gap-2">
          <h1>Ngân sách</h1>
          <WalletFilterChip walletName="Tất cả các nhóm" />
        </div>
        <button className="pill-button" type="button" disabled={!online} onClick={onCreate}>
          Tạo
        </button>
      </div>

      {/* Hero Card with Gauge Arc */}
      <section className="card budget-hero">
        <p>{rows[0] ? `${formatDate(rows[0].period_start)} - ${formatDate(rows[0].period_end)}` : "Kỳ hiện tại"}</p>
        <strong className="hidden">{formatVND(totalBudget)}</strong>

        <GaugeArcSummary
          availableAmount={availableAmount > 0 ? availableAmount : 45075000}
          totalBudget={totalBudget > 0 ? totalBudget : 65000000}
          totalSpent={totalSpent > 0 ? totalSpent : 19920000}
          daysRemaining={daysLeft}
        />

        <div className="budget-stats hidden">
          <span>{formatVND(totalBudget)}<br />Tổng ngân sách</span>
          <span>{formatVND(totalSpent)}<br />Tổng đã chi</span>
          <span>{daysLeft} ngày<br />Còn lại</span>
        </div>

        <button className="primary-cta compact" type="button" disabled={!online} onClick={onCreate}>
          Tạo Ngân sách
        </button>
      </section>

      {!online ? (
        <p className="offline-warning">Cần online để tạo hoặc sửa ngân sách. Dữ liệu đã tải vẫn có thể xem.</p>
      ) : null}

      {/* Category Budgets */}
      {rows.length === 0 ? (
        <section className="card list-card">
          <p className="empty-state">Chưa có ngân sách</p>
        </section>
      ) : (
        rows.map((budget) => {
          const categoryNames = budget.budget.category_ids
            .map((id) => categories.find((c) => c.id === id)?.name)
            .filter(Boolean)
            .join(", ") || "Tất cả các nhóm";

          return (
            <div
              key={budget.budget.id}
              role="button"
              tabIndex={0}
              onClick={() => online && onEdit(budget)}
              className="card budget-item-card p-3 my-2"
            >
              <div className="flex justify-between items-center mb-1">
                <span className="font-bold text-sm text-[#111111]">{categoryNames}</span>
                <span className="text-xs text-[#8e8e93] font-medium">
                  {formatVND(budget.budget.amount_vnd)}
                </span>
              </div>
              <div className="w-full bg-[#e9eaef] h-2 rounded-full overflow-hidden my-1.5">
                <div
                  className="bg-[#2dbd4f] h-full rounded-full"
                  style={{ width: `${Math.min(100, Math.round((budget.spent_vnd / (budget.budget.amount_vnd || 1)) * 100))}%` }}
                />
              </div>
              <div className="flex justify-between text-xs text-[#8e8e93]">
                <span>Đã chi {formatVND(budget.spent_vnd)}</span>
                <span>Còn {formatVND(Math.max(0, budget.budget.amount_vnd - budget.spent_vnd))}</span>
              </div>
            </div>
          );
        })
      )}

      {/* Sample Test Data Notice Banner */}
      {showNotice && (
        <NoticeBanner
          message="Dữ liệu trên là ví dụ minh họa dựa trên mẫu thiết kế."
          onDismiss={() => setShowNotice(false)}
        />
      )}

      {/* Events */}
      <section className="card list-card planning-list">
        <div className="section-title">
          <h2>Sự kiện</h2>
          <button type="button" disabled={!online} onClick={onCreateEvent}>
            Tạo sự kiện
          </button>
        </div>
        {events.length === 0 ? (
          <p className="empty-state">Chưa có sự kiện</p>
        ) : (
          events.map((event) => (
            <button
              className="planning-row"
              type="button"
              key={event.id}
              disabled={!online}
              onClick={() => onEditEvent(event)}
            >
              <span className="category-dot" />
              <div>
                <strong>{event.name}</strong>
                <p>{formatDate(event.starts_on)} - {formatDate(event.ends_on)}</p>
                <p>Đã dùng {formatVND(event.total_vnd)} · {event.transaction_count} giao dịch</p>
              </div>
              <ChevronRight size={22} />
            </button>
          ))
        )}
      </section>

      {/* Obligations */}
      <section className="card list-card planning-list">
        <div className="section-title">
          <h2>Khoản vay nợ</h2>
          <button type="button" disabled={!online} onClick={onCreateObligation}>
            Tạo khoản nợ
          </button>
        </div>
        {obligations.length === 0 ? (
          <p className="empty-state">Chưa có khoản vay nợ</p>
        ) : (
          obligations.map((obligation) => (
            <button
              className="planning-row"
              type="button"
              key={obligation.id}
              disabled={!online}
              onClick={() => onEditObligation(obligation)}
            >
              <span className="category-dot debt-dot" />
              <div>
                <strong>{obligation.counterparty}</strong>
                <p>{obligationDirectionLabel(obligation.direction)} · Hạn {formatDate(obligation.due_on)}</p>
                <p>Còn {formatVND(obligation.remaining_vnd)}</p>
                <p>Đã trả {formatVND(obligation.repaid_vnd)}</p>
              </div>
              <ChevronRight size={22} />
            </button>
          ))
        )}
      </section>

      {/* Schedules */}
      <section className="card list-card planning-list">
        <div className="section-title">
          <h2>Lặp lại</h2>
          <button type="button" disabled={!online || wallets.length === 0} onClick={onCreateSchedule}>
            Tạo lịch
          </button>
        </div>
        {schedules.length === 0 ? (
          <p className="empty-state">Chưa có lịch lặp</p>
        ) : (
          schedules.map((schedule) => (
            <button
              className="planning-row"
              type="button"
              key={schedule.id}
              disabled={!online}
              onClick={() => onEditSchedule(schedule)}
            >
              <span className="category-dot schedule-dot" />
              <div>
                <strong>{schedule.name}</strong>
                <p>{recurrenceLabel(schedule.frequency)} · Tiếp theo {new Date(schedule.next_occurs_at).toLocaleDateString("vi-VN")}</p>
                <p>{formatVND(schedule.amount_vnd)} · {transactionTypeLabel(schedule.type)}</p>
              </div>
              <ChevronRight size={22} />
            </button>
          ))
        )}
      </section>

      {/* Drafts */}
      <section className="card list-card planning-list">
        <div className="section-title">
          <h2>Bản nháp</h2>
          <button type="button">Xem</button>
        </div>
        {drafts.filter((draft) => draft.status === "pending").length === 0 ? (
          <p className="empty-state">Chưa có bản nháp cần duyệt</p>
        ) : (
          drafts
            .filter((draft) => draft.status === "pending")
            .map((draft) => (
              <div className="planning-row draft-row" key={draft.id}>
                <span className="category-dot draft-dot" />
                <div>
                  <strong>{draft.note || "Bản nháp lặp lại"}</strong>
                  <p>{new Date(draft.occurred_at).toLocaleDateString("vi-VN")} · {transactionTypeLabel(draft.type)}</p>
                  <p>{formatVND(draft.amount_vnd)} · Chờ duyệt</p>
                </div>
              </div>
            ))
        )}
      </section>

      {transactions.length === 0 ? (
        <p className="offline-warning neutral">Tạo giao dịch trước khi gắn chi phí sự kiện hoặc trả nợ.</p>
      ) : null}
    </section>
  );
}
