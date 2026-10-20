import { render, screen, within } from '@testing-library/react';
import { expect, it, vi } from 'vitest';
import type { Report } from '../app/analytics';
import { buildDailyExpenseBars, OverviewScreen, type OverviewScreenProps } from './OverviewScreen';

const props: OverviewScreenProps = {
  online: true, wallets: [], dashboard: null, report: null, insider: null,
  privacyMasked: false, walletDetail: null, onCloseWalletDetail: vi.fn(),
  onManageWallets: vi.fn(), onRefreshInsider: vi.fn(),
  formatVND: value => `${new Intl.NumberFormat('vi-VN').format(value)} đ`, formatPercent: value => `${value}%`,
};

const septemberReport: Report = {
  summary: {
    income_vnd: 0,
    expense_vnd: 465000,
    net_income_vnd: -465000,
    from: '2026-09-01',
    to: '2026-09-30',
    generated_at: '2026-09-09T00:00:00Z',
    timezone: 'Asia/Ho_Chi_Minh',
    data_version: 1,
  },
  daily: [
    { date: '2026-09-01', income_vnd: 0, expense_vnd: 1000, net_income_vnd: -1000, cumulative_net_vnd: -1000 },
    { date: '2026-09-02', income_vnd: 0, expense_vnd: 2000, net_income_vnd: -2000, cumulative_net_vnd: -3000 },
    { date: '2026-09-03', income_vnd: 0, expense_vnd: 3000, net_income_vnd: -3000, cumulative_net_vnd: -6000 },
    { date: '2026-09-04', income_vnd: 0, expense_vnd: 4000, net_income_vnd: -4000, cumulative_net_vnd: -10000 },
    { date: '2026-09-05', income_vnd: 0, expense_vnd: 5000, net_income_vnd: -5000, cumulative_net_vnd: -15000 },
    { date: '2026-09-06', income_vnd: 0, expense_vnd: 6000, net_income_vnd: -6000, cumulative_net_vnd: -21000 },
    { date: '2026-09-07', income_vnd: 0, expense_vnd: 7000, net_income_vnd: -7000, cumulative_net_vnd: -28000 },
    { date: '2026-09-08', income_vnd: 0, expense_vnd: 8000, net_income_vnd: -8000, cumulative_net_vnd: -36000 },
    { date: '2026-09-09', income_vnd: 0, expense_vnd: 9000, net_income_vnd: -9000, cumulative_net_vnd: -45000 },
    { date: '2026-09-10', income_vnd: 0, expense_vnd: 10000, net_income_vnd: -10000, cumulative_net_vnd: -55000 },
    { date: '2026-09-11', income_vnd: 0, expense_vnd: 11000, net_income_vnd: -11000, cumulative_net_vnd: -66000 },
    { date: '2026-09-12', income_vnd: 0, expense_vnd: 12000, net_income_vnd: -12000, cumulative_net_vnd: -78000 },
    { date: '2026-09-13', income_vnd: 0, expense_vnd: 13000, net_income_vnd: -13000, cumulative_net_vnd: -91000 },
    { date: '2026-09-14', income_vnd: 0, expense_vnd: 14000, net_income_vnd: -14000, cumulative_net_vnd: -105000 },
    { date: '2026-09-15', income_vnd: 0, expense_vnd: 15000, net_income_vnd: -15000, cumulative_net_vnd: -120000 },
    { date: '2026-09-16', income_vnd: 0, expense_vnd: 16000, net_income_vnd: -16000, cumulative_net_vnd: -136000 },
    { date: '2026-09-17', income_vnd: 0, expense_vnd: 17000, net_income_vnd: -17000, cumulative_net_vnd: -153000 },
    { date: '2026-09-18', income_vnd: 0, expense_vnd: 18000, net_income_vnd: -18000, cumulative_net_vnd: -171000 },
    { date: '2026-09-19', income_vnd: 0, expense_vnd: 19000, net_income_vnd: -19000, cumulative_net_vnd: -190000 },
    { date: '2026-09-20', income_vnd: 0, expense_vnd: 20000, net_income_vnd: -20000, cumulative_net_vnd: -210000 },
    { date: '2026-09-21', income_vnd: 0, expense_vnd: 21000, net_income_vnd: -21000, cumulative_net_vnd: -231000 },
    { date: '2026-09-22', income_vnd: 0, expense_vnd: 22000, net_income_vnd: -22000, cumulative_net_vnd: -253000 },
    { date: '2026-09-23', income_vnd: 0, expense_vnd: 23000, net_income_vnd: -23000, cumulative_net_vnd: -276000 },
    { date: '2026-09-24', income_vnd: 0, expense_vnd: 24000, net_income_vnd: -24000, cumulative_net_vnd: -300000 },
    { date: '2026-09-25', income_vnd: 0, expense_vnd: 25000, net_income_vnd: -25000, cumulative_net_vnd: -325000 },
    { date: '2026-09-26', income_vnd: 0, expense_vnd: 26000, net_income_vnd: -26000, cumulative_net_vnd: -351000 },
    { date: '2026-09-27', income_vnd: 0, expense_vnd: 27000, net_income_vnd: -27000, cumulative_net_vnd: -378000 },
    { date: '2026-09-28', income_vnd: 0, expense_vnd: 28000, net_income_vnd: -28000, cumulative_net_vnd: -406000 },
    { date: '2026-09-29', income_vnd: 0, expense_vnd: 29000, net_income_vnd: -29000, cumulative_net_vnd: -435000 },
    { date: '2026-09-30', income_vnd: 0, expense_vnd: 30000, net_income_vnd: -30000, cumulative_net_vnd: -465000 },
  ],
};

it('ends the seven-day preview on today in Ho Chi Minh time instead of future month rows', () => {
  const bars = buildDailyExpenseBars(septemberReport, new Date('2026-09-08T17:00:00Z'));

  expect(bars.map(item => item.label)).toEqual(['03/09', '04/09', '05/09', '06/09', '07/09', '08/09', '09/09']);
  expect(bars.map(item => item.amount)).toEqual([3000, 4000, 5000, 6000, 7000, 8000, 9000]);
  expect(bars.filter(item => item.isCurrent).map(item => item.label)).toEqual(['09/09']);
});

it('ends the preview on yesterday one second before Ho Chi Minh midnight', () => {
  const bars = buildDailyExpenseBars(septemberReport, new Date('2026-09-08T16:59:59Z'));

  expect(bars.map(item => item.label)).toEqual(['02/09', '03/09', '04/09', '05/09', '06/09', '07/09', '08/09']);
  expect(bars.filter(item => item.isCurrent).map(item => item.label)).toEqual(['08/09']);
});

it('uses the historical period end without marking its final day as today', () => {
  const historicalReport: Report = {
    ...septemberReport,
    summary: {
      ...septemberReport.summary,
      expense_vnd: 15000,
      net_income_vnd: -15000,
      to: '2026-09-05',
    },
    daily: septemberReport.daily!.slice(0, 5),
  };
  const bars = buildDailyExpenseBars(historicalReport, new Date('2026-09-08T17:00:00Z'));

  expect(bars.map(item => item.label)).toEqual(['01/09', '02/09', '03/09', '04/09', '05/09']);
  expect(bars.some(item => item.isCurrent)).toBe(false);
});

it('returns no preview rows when the report period is entirely in the future', () => {
  const futureReport: Report = {
    ...septemberReport,
    summary: { ...septemberReport.summary, from: '2026-10-01', to: '2026-10-03' },
    daily: [
      { date: '2026-10-01', income_vnd: 0, expense_vnd: 31000, net_income_vnd: -31000, cumulative_net_vnd: -31000 },
      { date: '2026-10-02', income_vnd: 0, expense_vnd: 32000, net_income_vnd: -32000, cumulative_net_vnd: -63000 },
      { date: '2026-10-03', income_vnd: 0, expense_vnd: 33000, net_income_vnd: -33000, cumulative_net_vnd: -96000 },
    ],
  };

  expect(buildDailyExpenseBars(futureReport, new Date('2026-09-08T17:00:00Z'))).toEqual([]);
});

it('returns an empty preview for a null report', () => {
  expect(buildDailyExpenseBars(null, new Date('2026-09-08T17:00:00Z'))).toEqual([]);
});

it('returns only available days for a short current period', () => {
  const shortReport: Report = { ...septemberReport, daily: septemberReport.daily!.slice(6, 9) };

  expect(buildDailyExpenseBars(shortReport, new Date('2026-09-08T17:00:00Z')).map(item => item.label)).toEqual([
    '07/09',
    '08/09',
    '09/09',
  ]);
});

it('does not report zero income and expense when no report has loaded', () => {
  render(<OverviewScreen {...props} />);
  const report = within(screen.getByLabelText('Báo cáo chi tiêu'));
  expect(report.queryAllByText('0 đ')).toHaveLength(0);
  expect(report.getByText(/Chưa tải được dữ liệu báo cáo/)).toBeVisible();
});

it('hides report totals and chart values when privacy is enabled', () => {
  render(<OverviewScreen {...props} privacyMasked report={{
    summary: { income_vnd: 1000000, expense_vnd: 250000, net_income_vnd: 750000, from: '2026-09-01', to: '2026-09-30', generated_at: '2026-09-09T00:00:00Z', timezone: 'Asia/Ho_Chi_Minh', data_version: 1 },
    daily: [{ date: '2026-09-09', income_vnd: 1000000, expense_vnd: 250000, net_income_vnd: 750000, cumulative_net_vnd: 750000 }],
  }} />);
  const report = screen.getByLabelText('Báo cáo chi tiêu');
  expect(report).not.toHaveTextContent('250.000');
  expect(report).not.toHaveTextContent('1.000.000');
  expect(report.querySelector('svg')).toBeNull();
  expect(report).toHaveTextContent('••••••');
});
