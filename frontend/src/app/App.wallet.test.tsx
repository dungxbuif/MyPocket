import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, it, vi } from 'vitest';
import { App } from './App';

afterEach(() => vi.unstubAllGlobals());

it('opens a real wallet card with a null transaction list as empty without unmounting the app', async () => {
  const wallet = { id: 'empty-wallet', name: 'Ví chưa giao dịch', type: 'cash', balance_vnd: 0, include_in_total: true, is_default_ai: false, version: 1 };
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
    const path = new URL(String(input), 'http://localhost').pathname;
    const payloads: Record<string, unknown> = {
      '/api/v1/me': { user: { id: 'wallet-owner', email: 'wallet@example.com', email_verified: true, display_name: 'Wallet', avatar_url: '' } },
      '/api/v1/wallets': { wallets: [wallet] },
      '/api/v1/categories': { categories: [] },
      '/api/v1/transactions': { transactions: [] },
      '/api/v1/budgets': { budgets: [] },
      '/api/v1/events': { events: [] },
      '/api/v1/obligations': { obligations: [] },
      '/api/v1/recurring-schedules': { schedules: [] },
      '/api/v1/transaction-drafts': { drafts: [] },
      '/api/v1/wallets/empty-wallet/detail': { status: 'ok', wallet, transactions: null, stale: false, generated_at: '2026-09-09T00:00:00Z', correlation_id: 'req_empty_wallet' },
    };
    return new Response(JSON.stringify(payloads[path] ?? { error: { code: 'INTERNAL_RETRYABLE' } }), { status: path in payloads ? 200 : 503 });
  }));
  render(<App />);
  await userEvent.click(await screen.findByRole('button', { name: /Ví chưa giao dịch/ }));
  const detail = await screen.findByLabelText('Chi tiết Ví chưa giao dịch');
  expect(within(detail).getByText('Chưa có giao dịch')).toBeVisible();
  expect(screen.getByRole('heading', { name: 'Ví của tôi' })).toBeVisible();
  await userEvent.click(within(detail).getByRole('button', { name: 'Đóng' }));
  expect(screen.queryByLabelText('Chi tiết Ví chưa giao dịch')).not.toBeInTheDocument();
});
