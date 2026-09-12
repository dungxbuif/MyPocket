import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { BudgetsScreen, type BudgetsScreenProps } from './BudgetsScreen';
import type { BudgetProgress } from '../app/planning';

function budget(
  spent: number,
  overrides: Partial<BudgetProgress['budget']> & { period_start?: string; period_end?: string } = {},
): BudgetProgress {
  const amount = overrides.amount_vnd ?? 500000;
  return {
    budget: {
      id: overrides.id ?? 'budget-1',
      name: overrides.name ?? 'Cafe tháng này',
      period_type: overrides.period_type ?? 'monthly',
      amount_vnd: amount,
      category_ids: overrides.category_ids ?? [],
      all_categories: overrides.all_categories ?? true,
      version: overrides.version ?? 1,
    },
    period_start: overrides.period_start ?? '2026-09-01',
    period_end: overrides.period_end ?? '2026-09-30',
    spent_vnd: spent,
    remaining_vnd: amount - spent,
    percent: amount === 0 ? 0 : (spent / amount) * 100,
    alert_80: spent >= amount * 0.8,
    alert_100: spent >= amount,
  };
}

function props(budgets: BudgetProgress[] | null): BudgetsScreenProps {
  return {
    budgets, events: [], obligations: [], schedules: [], drafts: [], categories: [], wallets: [], transactions: [], online: true,
    onCreate: vi.fn(), onCreateEvent: vi.fn(), onCreateObligation: vi.fn(), onCreateSchedule: vi.fn(),
    onEdit: vi.fn(), onEditEvent: vi.fn(), onEditObligation: vi.fn(), onEditSchedule: vi.fn(),
    formatVND: value => `${value.toLocaleString('vi-VN')} đ`, formatDate: value => value,
    obligationDirectionLabel: value => value, recurrenceLabel: value => value, transactionTypeLabel: value => value,
  };
}

describe('persisted budget display', () => {
  it('shows only the selected budget in the summary and keeps its ID across refreshes', async () => {
    const user = userEvent.setup();
    const first = budget(410000, { id: 'budget-food', name: 'Ăn uống tháng 9' });
    const second = budget(100000, {
      id: 'budget-travel',
      name: 'Di chuyển tuần 2',
      period_type: 'weekly',
      amount_vnd: 200000,
      period_start: '2026-09-07',
      period_end: '2026-09-13',
    });
    const { rerender } = render(<BudgetsScreen {...props([first, second])} />);

    const selector = screen.getByRole('combobox', { name: 'Ngân sách hiển thị' });
    let summary = screen.getByRole('region', { name: 'Ngân sách đã chọn' });
    let stats = within(summary).getByRole('group', { name: 'Số liệu ngân sách đã chọn' });
    expect(selector).toHaveValue('budget-food');
    expect(within(summary).getByRole('heading', { name: 'Ăn uống tháng 9' })).toBeVisible();
    expect(within(summary).getByText('2026-09-01 - 2026-09-30')).toBeVisible();
    expect(stats).toBeVisible();
    expect(stats).toHaveTextContent('500.000 đ');
    expect(stats).toHaveTextContent('410.000 đ');
    expect(stats).not.toHaveTextContent('200.000 đ');
    expect(stats).not.toHaveTextContent('100.000 đ');

    await user.selectOptions(selector, 'budget-travel');
    summary = screen.getByRole('region', { name: 'Ngân sách đã chọn' });
    stats = within(summary).getByRole('group', { name: 'Số liệu ngân sách đã chọn' });
    expect(within(summary).getByRole('heading', { name: 'Di chuyển tuần 2' })).toBeVisible();
    expect(within(summary).getByText('2026-09-07 - 2026-09-13')).toBeVisible();
    expect(stats).toBeVisible();
    expect(stats).toHaveTextContent('200.000 đ');
    expect(stats).toHaveTextContent('100.000 đ');
    expect(stats).not.toHaveTextContent('500.000 đ');
    expect(stats).not.toHaveTextContent('410.000 đ');

    rerender(<BudgetsScreen {...props([second, first])} />);
    expect(screen.getByRole('combobox', { name: 'Ngân sách hiển thị' })).toHaveValue('budget-travel');
    expect(within(screen.getByRole('region', { name: 'Ngân sách đã chọn' })).getByRole('heading', { name: 'Di chuyển tuần 2' })).toBeVisible();

    rerender(<BudgetsScreen {...props([first])} />);
    expect(screen.getByRole('combobox', { name: 'Ngân sách hiển thị' })).toHaveValue('budget-food');
    expect(within(screen.getByRole('region', { name: 'Ngân sách đã chọn' })).getByRole('heading', { name: 'Ăn uống tháng 9' })).toBeVisible();
  });

  it('does not present an unloaded budget list as an empty account balance', () => {
    render(<BudgetsScreen {...props(null)} />);
    expect(screen.getByRole('status')).toHaveTextContent('Chưa có dữ liệu ngân sách');
    expect(screen.queryByText('Số tiền bạn có thể chi')).not.toBeInTheDocument();
    expect(screen.queryByText('Chưa có ngân sách')).not.toBeInTheDocument();
  });

  it('opens the named budget with Enter and Space, retaining its category context', async () => {
    const input = props([budget(410000)]);
    render(<BudgetsScreen {...input} />);
    const row = screen.getByRole('button', { name: /Cafe tháng này/ });
    expect(row).toHaveTextContent('Tất cả các nhóm');
    row.focus();
    await userEvent.keyboard('{Enter}');
    await userEvent.keyboard(' ');
    expect(input.onEdit).toHaveBeenCalledTimes(2);
    expect(input.onEdit).toHaveBeenLastCalledWith(input.budgets![0]);
  });

  it('does not allow an offline budget to open for editing', async () => {
    const input = { ...props([budget(0)]), online: false };
    render(<BudgetsScreen {...input} />);
    const row = screen.getByRole('button', { name: /Cafe tháng này/ });
    expect(row).toBeDisabled();
    await userEvent.click(row);
    expect(input.onEdit).not.toHaveBeenCalled();
  });

  it('shows zero remaining when spending exactly reaches the budget', () => {
    render(<BudgetsScreen {...props([budget(500000)])} />);
    const summary = screen.getByRole('region', { name: 'Ngân sách đã chọn' });
    expect(within(summary).getByText('0 đ', { exact: true })).toBeVisible();
    expect(within(summary).queryByText('Vượt ngân sách')).not.toBeInTheDocument();
  });

  it('shows the explicit overspend amount instead of clamping it to zero', () => {
    render(<BudgetsScreen {...props([budget(650000)])} />);
    const summary = screen.getByRole('region', { name: 'Ngân sách đã chọn' });
    expect(within(summary).getByText('150.000 đ')).toBeVisible();
    expect(within(summary).getByText('Vượt ngân sách')).toBeVisible();
    expect(within(summary).queryByText('0 đ', { exact: true })).not.toBeInTheDocument();
  });

  it('counts remaining days by the Ho Chi Minh calendar date across 07:00', () => {
    vi.useFakeTimers();
    try {
      const input = props([budget(0)]);
      vi.setSystemTime(new Date('2026-09-28T18:00:00Z'));
      const { rerender } = render(<BudgetsScreen {...input} />);
      let stats = within(screen.getByRole('region', { name: 'Ngân sách đã chọn' }))
        .getByRole('group', { name: 'Số liệu ngân sách đã chọn' });
      expect(stats).toHaveTextContent('1 ngày');

      vi.setSystemTime(new Date('2026-09-29T18:00:00Z'));
      rerender(<BudgetsScreen {...input} />);
      stats = within(screen.getByRole('region', { name: 'Ngân sách đã chọn' }))
        .getByRole('group', { name: 'Số liệu ngân sách đã chọn' });
      expect(stats).toHaveTextContent('0 ngày');

      vi.setSystemTime(new Date('2026-09-30T01:00:00Z'));
      rerender(<BudgetsScreen {...input} />);
      stats = within(screen.getByRole('region', { name: 'Ngân sách đã chọn' }))
        .getByRole('group', { name: 'Số liệu ngân sách đã chọn' });
      expect(stats).toHaveTextContent('0 ngày');
    } finally {
      vi.useRealTimers();
    }
  });

  it('shows actual unspent totals rather than sample spending', () => {
    render(<BudgetsScreen {...props([budget(0)])} />);
    expect(screen.getByText('0 M đ')).toBeVisible();
    expect(screen.queryByText(/ví dụ minh họa/)).not.toBeInTheDocument();
  });

  it('shows an empty state without fabricated spendable money or days', () => {
    render(<BudgetsScreen {...props([])} />);
    expect(screen.getByText('Chưa có ngân sách')).toBeVisible();
    expect(screen.queryByRole('region', { name: 'Ngân sách đã chọn' })).not.toBeInTheDocument();
    expect(screen.queryByText('Số tiền bạn có thể chi')).not.toBeInTheDocument();
    expect(screen.queryByText(/\d+ ngày/)).not.toBeInTheDocument();
    expect(screen.queryByText('65 M đ')).not.toBeInTheDocument();
  });
});
