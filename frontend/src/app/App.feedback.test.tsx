import { act, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { saveFinanceMirror } from '../offline/db';
import * as offlineConflicts from '../offline/conflicts';
import { App } from './App';

const owner = {
  id: 'feedback-owner',
  email: 'feedback@example.com',
  email_verified: true,
  display_name: 'Feedback Owner',
  avatar_url: '',
};

const verificationWallet = {
  id: 'wallet-verification',
  name: 'Ví kiểm chứng',
  type: 'cash',
  balance_vnd: 880000,
  include_in_total: true,
  is_default_ai: true,
  version: 3,
} as const;

const reserveWallet = {
  id: 'wallet-reserve',
  name: 'Ví dự phòng',
  type: 'bank',
  balance_vnd: 240000,
  include_in_total: true,
  is_default_ai: false,
  version: 2,
} as const;

const offlineWallet = {
  id: 'wallet-offline',
  name: 'Ví ngoại tuyến đã lưu',
  type: 'cash',
  balance_vnd: 510000,
  include_in_total: true,
  is_default_ai: true,
  version: 5,
} as const;

const lateOnlineWallet = {
  id: 'wallet-online-late',
  name: 'Ví online đến muộn',
  type: 'bank',
  balance_vnd: 920000,
  include_in_total: true,
  is_default_ai: false,
  version: 6,
} as const;

const confirmedTransaction = {
  id: 'transaction-confirmed',
  type: 'expense',
  source_wallet_id: verificationWallet.id,
  category_id: 'category-food',
  amount_vnd: 120000,
  balance_after_vnd: 880000,
  occurred_at: '2026-09-09T02:30:00Z',
  note: 'Bữa trưa đã xác nhận',
  with_person: '',
  event_ref: '',
  excluded_from_reports: false,
  version: 4,
};

const confirmedInsider = {
  selected_category: { id: 'category-food', name: 'Ăn uống đã xác nhận', transaction_count: 4 },
  spent_vnd: 320000,
  average_daily_vnd: 40000,
  prior_average_daily_vnd: 35000,
  change_percent: 14.3,
  not_comparable: false,
  elapsed_days: 8,
  generated_at: '2026-09-09T03:00:00Z',
  timezone: 'Asia/Ho_Chi_Minh',
  from: '2026-09-01',
  to: '2026-09-09',
  data_version: 8,
};

const notice = {
  id: 'notice-budget',
  kind: 'budget_threshold',
  title: 'Ngân sách sắp chạm mức',
  body: 'Ngân sách ăn uống đã dùng 80%.',
  source_type: 'budget',
  source_id: 'budget-food',
  created_at: '2026-09-09T04:00:00Z',
};

const defaultPayloads: Record<string, unknown> = {
  '/api/v1/me': { user: owner },
  '/api/v1/wallets': { wallets: [verificationWallet, reserveWallet] },
  '/api/v1/categories': { categories: [{ id: 'category-food', kind: 'expense', name: 'Ăn uống', system_key: 'expense_food', is_system: true, version: 2 }] },
  '/api/v1/transactions': { transactions: [confirmedTransaction] },
  '/api/v1/budgets': { budgets: [{ budget: { id: 'budget-food', name: 'Ăn uống tháng 9', period_type: 'monthly', amount_vnd: 1000000, category_ids: ['category-food'], all_categories: false, version: 2 }, period_start: '2026-09-01', period_end: '2026-09-30', spent_vnd: 320000, remaining_vnd: 680000, percent: 32, alert_80: false, alert_100: false }] },
  '/api/v1/events': { events: [] },
  '/api/v1/obligations': { obligations: [] },
  '/api/v1/recurring-schedules': { schedules: [] },
  '/api/v1/transaction-drafts': { drafts: [] },
  '/api/v1/notifications': { notifications: [notice] },
  '/api/v1/dashboard': { report: { net_worth_vnd: 1120000, wallet_net_worth_vnd: 1120000, investment_market_value_vnd: 0, combined_net_worth_vnd: 1120000, missing_asset_price_count: 0, summary: { income_vnd: 0, expense_vnd: 120000, net_income_vnd: -120000, generated_at: '2026-09-09T03:00:00Z', timezone: 'Asia/Ho_Chi_Minh', from: '2026-09-01', to: '2026-09-09', data_version: 8 }, wallets: [{ id: verificationWallet.id, name: verificationWallet.name, balance_vnd: verificationWallet.balance_vnd, include_in_total: true }, { id: reserveWallet.id, name: reserveWallet.name, balance_vnd: reserveWallet.balance_vnd, include_in_total: true }], recent_transactions: [{ id: confirmedTransaction.id, type: confirmedTransaction.type, amount_vnd: confirmedTransaction.amount_vnd, note: confirmedTransaction.note, occurred_at: confirmedTransaction.occurred_at }] } },
  '/api/v1/reports/daily': { report: { summary: { income_vnd: 0, expense_vnd: 120000, net_income_vnd: -120000, generated_at: '2026-09-09T03:00:00Z', timezone: 'Asia/Ho_Chi_Minh', from: '2026-09-01', to: '2026-09-09', data_version: 8 }, daily: [{ date: '2026-09-09', income_vnd: 0, expense_vnd: 120000, net_income_vnd: -120000, cumulative_net_vnd: -120000 }] } },
  '/api/v1/reports/insider': { report: confirmedInsider },
  '/api/v1/assets': { assets: [] },
  '/api/v1/portfolio/summary': { summary: { investment_market_value_vnd: 0, missing_price_count: 0, included_position_count: 0, position_count: 0 } },
  '/api/v1/api-keys': { keys: [] },
  '/api/v1/audit/access': { allowed: false },
  '/api/v1/auth/logout': {},
};

function response(body: unknown, status = 200) {
  return new Response(JSON.stringify({ status: status < 400 ? 'ok' : 'error', correlation_id: 'req_feedback', ...(body as Record<string, unknown>) }), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

type Route = (url: URL, options?: RequestInit) => Response | Promise<Response>;

function renderApp(overrides: Record<string, Route> = {}) {
  const fetcher = vi.fn(async (input: RequestInfo | URL, options?: RequestInit) => {
    const url = new URL(String(input), 'http://localhost');
    const route = overrides[url.pathname];
    if (route) return route(url, options);
    if (url.pathname in defaultPayloads) return response(defaultPayloads[url.pathname]);
    return response({ error: { code: 'INTERNAL_FAILURE', message: 'UNEXPECTED_FIXTURE_ENDPOINT' } }, 503);
  });
  vi.stubGlobal('fetch', fetcher);
  render(<App />);
  return fetcher;
}

function walletDetail(wallet: typeof verificationWallet | typeof reserveWallet, note: string) {
  return {
    wallet: { id: wallet.id, name: wallet.name, balance_vnd: wallet.balance_vnd, include_in_total: wallet.include_in_total },
    transactions: [{ id: `${wallet.id}-transaction`, type: 'expense', amount_vnd: 25000, note, occurred_at: '2026-09-09T01:00:00Z' }],
    stale: false,
    generated_at: '2026-09-09T05:00:00Z',
    correlation_id: `req_${wallet.id}`,
  };
}

function deferredResponse() {
  let resolve!: (value: Response) => void;
  const promise = new Promise<Response>((done) => { resolve = done; });
  return { promise, resolve };
}

function deferredValue<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => { resolve = done; });
  return { promise, resolve };
}

afterEach(() => {
  Object.defineProperty(navigator, 'onLine', { configurable: true, value: true });
  localStorage.clear();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('App asynchronous operation feedback', () => {
  it('shows a safe wallet-detail failure and loads it only after an explicit retry', async () => {
    let attempts = 0;
    renderApp({
      '/api/v1/wallets/wallet-verification/detail': () => ++attempts === 1
        ? response({ error: { code: 'INTERNAL_RETRYABLE', message: 'PRIVATE_WALLET_FAILURE' } }, 503)
        : response(walletDetail(verificationWallet, 'Chi tiết đã tải lại')),
    });

    await userEvent.click(await screen.findByRole('button', { name: /Ví kiểm chứng/ }));
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Không tải được');
    expect(alert).toHaveTextContent('req_feedback');
    expect(screen.queryByText('PRIVATE_WALLET_FAILURE')).not.toBeInTheDocument();
    expect(attempts).toBe(1);

    await userEvent.click(within(alert).getByRole('button', { name: 'Thử lại' }));
    expect(await screen.findByLabelText('Chi tiết Ví kiểm chứng')).toBeVisible();
    expect(screen.getByText('Chi tiết đã tải lại')).toBeVisible();
    expect(attempts).toBe(2);
  });

  it('publishes only the latest wallet selection when deferred requests resolve in reverse order', async () => {
    const first = deferredResponse();
    const second = deferredResponse();
    renderApp({
      '/api/v1/wallets/wallet-verification/detail': () => first.promise,
      '/api/v1/wallets/wallet-reserve/detail': () => second.promise,
    });

    await userEvent.click(await screen.findByRole('button', { name: /Ví kiểm chứng/ }));
    await userEvent.click(screen.getByRole('button', { name: /Ví dự phòng/ }));
    await act(async () => second.resolve(response(walletDetail(reserveWallet, 'Kết quả mới'))));
    expect(await screen.findByLabelText('Chi tiết Ví dự phòng')).toBeVisible();
    await act(async () => first.resolve(response(walletDetail(verificationWallet, 'Kết quả cũ'))));

    expect(screen.getByLabelText('Chi tiết Ví dự phòng')).toBeVisible();
    expect(screen.queryByLabelText('Chi tiết Ví kiểm chứng')).not.toBeInTheDocument();
    expect(screen.queryByText('Kết quả cũ')).not.toBeInTheDocument();
  });

  it('does not reopen wallet detail when a newer pending selection resolves after the detail was closed', async () => {
    const pending = deferredResponse();
    renderApp({
      '/api/v1/wallets/wallet-verification/detail': () => Promise.resolve(response(walletDetail(verificationWallet, 'Đã xác nhận'))),
      '/api/v1/wallets/wallet-reserve/detail': () => pending.promise,
    });

    await userEvent.click(await screen.findByRole('button', { name: /Ví kiểm chứng/ }));
    const detail = await screen.findByLabelText('Chi tiết Ví kiểm chứng');
    await userEvent.click(screen.getByRole('button', { name: /Ví dự phòng/ }));
    await userEvent.click(within(detail).getByRole('button', { name: 'Đóng' }));
    await act(async () => pending.resolve(response(walletDetail(reserveWallet, 'Không được mở lại'))));

    expect(screen.queryByLabelText('Chi tiết Ví dự phòng')).not.toBeInTheDocument();
    expect(screen.queryByText('Không được mở lại')).not.toBeInTheDocument();
  });

  it('keeps successful wallets and transactions when the budget refresh fails', async () => {
    renderApp({
      '/api/v1/budgets': () => response({ error: { code: 'INTERNAL_RETRYABLE', message: 'PRIVATE_BUDGET_FAILURE' } }, 503),
    });

    expect(await screen.findByRole('button', { name: /Ví kiểm chứng/ })).toBeVisible();
    await userEvent.click(screen.getByRole('button', { name: 'Ngân sách' }));
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Không tải được một phần dữ liệu');
    expect(alert).toHaveTextContent('req_feedback');
    await userEvent.click(screen.getByRole('button', { name: 'Sổ giao dịch' }));
    expect(await screen.findByText('Bữa trưa đã xác nhận')).toBeVisible();
    expect(screen.queryByText('PRIVATE_BUDGET_FAILURE')).not.toBeInTheDocument();
  });

  it('keeps a notification unread after mark-read failure and changes it once after retry succeeds', async () => {
    let attempts = 0;
    renderApp({
      '/api/v1/notifications/notice-budget/read': () => ++attempts === 1
        ? response({ error: { code: 'INTERNAL_RETRYABLE', message: 'PRIVATE_NOTICE_FAILURE' } }, 503)
        : response({}),
    });

    const notificationButton = await screen.findByRole('button', { name: 'Thông báo' });
    await waitFor(() => expect(notificationButton).toHaveTextContent('1'));
    await userEvent.click(notificationButton);
    await userEvent.click(await screen.findByRole('button', { name: /Ngân sách sắp chạm mức/ }));
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Chưa đánh dấu đã đọc');
    expect(notificationButton).toHaveTextContent('1');
    expect(screen.queryByText('PRIVATE_NOTICE_FAILURE')).not.toBeInTheDocument();

    await userEvent.click(within(alert).getByRole('button', { name: 'Thử lại' }));
    await waitFor(() => expect(notificationButton).not.toHaveTextContent('1'));
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(attempts).toBe(2);
  });

  it('keeps confirmed Insider data visible when refresh fails and replaces it after explicit retry', async () => {
    let attempts = 0;
    renderApp({
      '/api/v1/reports/insider': () => {
        attempts += 1;
        if (attempts === 1) return response({ report: confirmedInsider });
        if (attempts === 2) return response({ error: { code: 'INTERNAL_RETRYABLE', message: 'PRIVATE_INSIDER_FAILURE' } }, 503);
        return response({ report: { ...confirmedInsider, selected_category: { id: 'category-travel', name: 'Di chuyển mới', transaction_count: 2 }, spent_vnd: 150000 } });
      },
    });

    expect(await screen.findByRole('heading', { name: /Ăn uống đã xác nhận/ })).toBeVisible();
    await userEvent.click(screen.getByRole('button', { name: 'Làm mới Money Insider' }));
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Không tải được Money Insider');
    expect(screen.getByRole('heading', { name: /Ăn uống đã xác nhận/ })).toBeVisible();
    expect(screen.queryByText('PRIVATE_INSIDER_FAILURE')).not.toBeInTheDocument();

    await userEvent.click(within(alert).getByRole('button', { name: 'Thử lại' }));
    expect(await screen.findByRole('heading', { name: /Di chuyển mới/ })).toBeVisible();
    expect(screen.queryByRole('heading', { name: /Ăn uống đã xác nhận/ })).not.toBeInTheDocument();
  });

  it('does not claim logout succeeded when the server session may still be active', async () => {
    renderApp({
      '/api/v1/auth/logout': () => response({ error: { code: 'INTERNAL_RETRYABLE', message: 'PRIVATE_LOGOUT_FAILURE' } }, 503),
    });

    await userEvent.click(await screen.findByRole('button', { name: 'Tài khoản' }));
    await userEvent.click(screen.getByRole('button', { name: 'Đăng xuất' }));
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Phiên trên máy chủ có thể vẫn còn hoạt động');
    expect(screen.getByRole('button', { name: 'Đăng xuất' })).toBeVisible();
    expect(screen.queryByRole('button', { name: 'Đăng nhập bằng Google' })).not.toBeInTheDocument();
    expect(screen.queryByText('PRIVATE_LOGOUT_FAILURE')).not.toBeInTheDocument();
    expect(localStorage.getItem('mypocket.current-user.v1')).toContain(owner.id);
  });

  it('suppresses a pending wallet result after a failed logout attempt invalidates the view lifecycle', async () => {
    const pending = deferredResponse();
    renderApp({
      '/api/v1/wallets/wallet-verification/detail': () => pending.promise,
      '/api/v1/auth/logout': () => response({ error: { code: 'INTERNAL_RETRYABLE' } }, 503),
    });

    await userEvent.click(await screen.findByRole('button', { name: /Ví kiểm chứng/ }));
    await userEvent.click(screen.getByRole('button', { name: 'Tài khoản' }));
    await userEvent.click(screen.getByRole('button', { name: 'Đăng xuất' }));
    await screen.findByText(/Phiên trên máy chủ có thể vẫn còn hoạt động/);
    await userEvent.click(screen.getByRole('button', { name: 'Tổng quan' }));
    await act(async () => pending.resolve(response(walletDetail(verificationWallet, 'Kết quả sau đăng xuất'))));

    expect(screen.queryByLabelText('Chi tiết Ví kiểm chứng')).not.toBeInTheDocument();
    expect(screen.queryByText('Kết quả sau đăng xuất')).not.toBeInTheDocument();
  });

  it('does not publish conflicts read after a failed logout invalidates the refresh owner', async () => {
    const pendingConflicts = deferredValue<Awaited<ReturnType<typeof offlineConflicts.listOpenConflicts>>>();
    vi.spyOn(offlineConflicts, 'listOpenConflicts').mockImplementation(() => pendingConflicts.promise);
    renderApp({
      '/api/v1/auth/logout': () => response({ error: { code: 'INTERNAL_RETRYABLE' } }, 503),
    });

    await screen.findByRole('button', { name: /Ví kiểm chứng/ });
    await waitFor(() => expect(offlineConflicts.listOpenConflicts).toHaveBeenCalledOnce());
    await userEvent.click(screen.getByRole('button', { name: 'Tài khoản' }));
    await userEvent.click(screen.getByRole('button', { name: 'Đăng xuất' }));
    await screen.findByText(/Phiên trên máy chủ có thể vẫn còn hoạt động/);
    await act(async () => pendingConflicts.resolve([{
      conflict_id: 'conflict-stale',
      mutation_id: 'mutation-stale',
      entity_type: 'transaction',
      entity_id: 'transaction-private',
      operation: 'update',
      base_version: 1,
      server_version: 2,
      local_payload: { note: 'Dữ liệu riêng đến muộn', amount_vnd: 45000 },
      server_payload: { note: 'Bản máy chủ', amount_vnd: 45000 },
      status: 'open',
      created_at: '2026-09-09T06:00:00Z',
    }]));

    expect(screen.queryByLabelText('Xung đột đồng bộ')).not.toBeInTheDocument();
    expect(screen.queryByText('Dữ liệu riêng đến muộn')).not.toBeInTheDocument();
  });

  it('does not let an obsolete online refresh overwrite offline hydration after connectivity changes', async () => {
    await saveFinanceMirror({
      userID: owner.id,
      wallets: [offlineWallet],
      categories: [],
      transactions: [],
      assets: [],
    });
    const pendingWallets = deferredResponse();
    renderApp({
      '/api/v1/wallets': () => pendingWallets.promise,
    });

    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/v1/wallets', expect.any(Object)));
    await act(async () => {
      Object.defineProperty(navigator, 'onLine', { configurable: true, value: false });
      window.dispatchEvent(new Event('offline'));
    });
    expect(await screen.findByRole('button', { name: /Ví ngoại tuyến đã lưu/ })).toBeVisible();
    await act(async () => pendingWallets.resolve(response({ wallets: [lateOnlineWallet] })));

    expect(screen.getByRole('button', { name: /Ví ngoại tuyến đã lưu/ })).toBeVisible();
    expect(screen.queryByRole('button', { name: /Ví online đến muộn/ })).not.toBeInTheDocument();
  });
});
