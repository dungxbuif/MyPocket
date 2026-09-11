import * as React from "react";
import { ChevronRight } from "lucide-react";
import type {
  BudgetProgress,
  EventSummary,
  ObligationDirection,
  ObligationSummary,
  RecurrenceFrequency,
  RecurringSchedule,
  TransactionDraft,
} from "../app/planning";
import type { CategorySummary, Transaction, TransactionType, WalletSummary } from "../app/finance";
import { ActionCard } from "../app/components";
import { GaugeArcSummary } from "../components/charts/GaugeArcSummary";
import { WalletFilterChip } from "../components/navigation/WalletFilterChip";
import { Select } from "../components/ui/select";

const DAY_MS = 86400000;

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

function calendarDayNumber(day: string): number {
  const [year, month, date] = day.slice(0, 10).split("-").map(Number);
  return Date.UTC(year, month - 1, date) / DAY_MS;
}

function remainingCalendarDays(periodEnd: string, now = new Date()): number {
  return Math.max(0, calendarDayNumber(periodEnd) - calendarDayNumber(hoChiMinhDayKey(now)));
}

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
  obligationDirectionLabel: (dir: ObligationDirection) => string;
  recurrenceLabel: (freq: RecurrenceFrequency) => string;
  transactionTypeLabel: (type: TransactionType) => string;
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
  const rows = budgets ?? [];
  const [selectedID, setSelectedID] = React.useState("");
  const selectedBudget = rows.find((row) => row.budget.id === selectedID) ?? rows[0];
  const totalBudget = selectedBudget?.budget.amount_vnd ?? 0;
  const totalSpent = selectedBudget?.spent_vnd ?? 0;
  const availableAmount = selectedBudget?.remaining_vnd ?? 0;
  const daysLeft = selectedBudget
    ? remainingCalendarDays(selectedBudget.period_end)
    : 0;

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
      {budgets === null ? (
        <p className="empty-state" role="status">Chưa có dữ liệu ngân sách. Nếu chưa tải được, hãy kết nối mạng và tải lại.</p>
      ) : selectedBudget ? (
        <section className="card budget-hero" aria-label="Ngân sách đã chọn">
          <label className="form-row" htmlFor="displayed-budget">
            Ngân sách hiển thị
            <Select
              id="displayed-budget"
              value={selectedBudget.budget.id}
              onChange={(event) => setSelectedID(event.target.value)}
            >
              {rows.map((row) => (
                <option key={row.budget.id} value={row.budget.id}>{row.budget.name}</option>
              ))}
            </Select>
          </label>
          <h2>{selectedBudget.budget.name}</h2>
          <p>{formatDate(selectedBudget.period_start)} - {formatDate(selectedBudget.period_end)}</p>

          <GaugeArcSummary
            availableAmount={availableAmount}
            totalBudget={totalBudget}
            totalSpent={totalSpent}
            daysRemaining={daysLeft}
          />

          <div className="budget-stats" role="group" aria-label="Số liệu ngân sách đã chọn">
            <span>{formatVND(totalBudget)}<br />Tổng ngân sách</span>
            <span>{formatVND(totalSpent)}<br />Tổng đã chi</span>
            <span>{daysLeft} ngày<br />Còn lại</span>
          </div>

          <button className="primary-cta compact" type="button" disabled={!online} onClick={onCreate}>
            Tạo Ngân sách
          </button>
        </section>
      ) : null}

      {!online ? (
        <p className="offline-warning">Cần online để tạo hoặc sửa ngân sách. Dữ liệu đã tải vẫn có thể xem.</p>
      ) : null}

      {/* Category Budgets */}
      {budgets === null ? null : rows.length === 0 ? (
        <section className="card list-card">
          <p className="empty-state">Chưa có ngân sách</p>
        </section>
      ) : (
        rows.map((budget) => {
          const categoryNames = (budget.budget.category_ids ?? [])
            .map((id) => categories.find((c) => c.id === id)?.name)
            .filter(Boolean)
            .join(", ") || "Tất cả các nhóm";

          return (
            <ActionCard
              key={budget.budget.id}
              disabled={!online}
              onClick={() => onEdit(budget)}
              className="card budget-item-card p-3 my-2"
            >
              <div className="flex justify-between items-center mb-1">
                <span className="font-bold text-sm text-[#111111]">{budget.budget.name}</span>
                <span className="text-xs text-[#8e8e93] font-medium">
                  {formatVND(budget.budget.amount_vnd)}
                </span>
              </div>
              <p className="text-xs text-[#8e8e93]">{categoryNames}</p>
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
              {budget.alert_100 ? <p className="budget-alert">Đã vượt 100%</p> : budget.alert_80 ? <p className="budget-alert">Đã chạm 80%</p> : null}
            </ActionCard>
          );
        })
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
