import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { App } from './App';

afterEach(() => {
  Object.defineProperty(navigator, 'onLine', { configurable: true, value: true });
  vi.unstubAllGlobals();
});

const result = (label: string) => ({ results: [{ kind: 'wallet', id: label, label, detail: 'cash' }] });
const response = (body: unknown, status = 200) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });

async function openSearch(search: (query: string) => Promise<Response>) {
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(String(input), 'http://localhost');
    if (url.pathname === '/api/v1/search') return search(url.searchParams.get('q') ?? '');
    if (url.pathname === '/api/v1/me') return response({ user: { id: 'search-owner', email: 'search@example.com', email_verified: true, display_name: 'Search', avatar_url: '' } });
    if (url.pathname === '/api/v1/wallets') return response({ wallets: [] });
    if (url.pathname === '/api/v1/categories') return response({ categories: [] });
    if (url.pathname === '/api/v1/transactions') return response({ transactions: [] });
    if (url.pathname === '/api/v1/auth/logout') return response({});
    return response({ error: { code: 'INTERNAL_FAILURE', message: 'unrelated fixture endpoint' } }, 503);
  }));
  render(<App />);
  await userEvent.click(await screen.findByRole('button', { name: 'Tìm kiếm' }));
  return screen.getByLabelText('Tìm kiếm giao dịch và ví');
}

function deferred() {
  let resolve!: (response: Response) => void;
  const promise = new Promise<Response>(done => { resolve = done; });
  return { promise, resolve };
}

describe('global search feedback', () => {
  it('preserves the query on API failure and retries explicitly with safe error details', async () => {
    let calls = 0;
    const input = await openSearch(async () => ++calls === 1
      ? response({ error: { code: 'INTERNAL_FAILURE', message: 'PRIVATE_SERVER_DETAIL' }, correlation_id: 'req_search_failure' }, 503)
      : response(result('Ví tìm lại')));
    fireEvent.change(input, { target: { value: 'Ví tìm' } });
    expect(await screen.findByRole('alert')).toHaveTextContent('req_search_failure');
    expect(input).toHaveValue('Ví tìm');
    expect(screen.queryByText('PRIVATE_SERVER_DETAIL')).not.toBeInTheDocument();
    expect(screen.queryByText('Không tìm thấy kết quả.')).not.toBeInTheDocument();
    expect(calls).toBe(1);
    await userEvent.click(within(screen.getByRole('alert')).getByRole('button', { name: 'Thử lại' }));
    expect(await screen.findByText('Ví tìm lại')).toBeVisible();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('shows loading rather than empty until a successful empty response', async () => {
    const pending = deferred();
    const input = await openSearch(() => pending.promise);
    fireEvent.change(input, { target: { value: 'missing' } });
    expect(screen.getByRole('status')).toHaveTextContent('Đang tìm kiếm');
    expect(screen.queryByText('Không tìm thấy kết quả.')).not.toBeInTheDocument();
    await act(async () => pending.resolve(response({ results: [] })));
    expect(await screen.findByText('Không tìm thấy kết quả.')).toBeVisible();
  });

  it.each(['success', 'failure'])('ignores an obsolete %s after the newer query succeeds', async outcome => {
    const pending = deferred();
    const input = await openSearch(query => query === 'old' ? pending.promise : Promise.resolve(response(result('Kết quả mới'))));
    fireEvent.change(input, { target: { value: 'old' } });
    fireEvent.change(input, { target: { value: 'new' } });
    await screen.findByText('Kết quả mới');
    await act(async () => pending.resolve(outcome === 'success' ? response(result('Kết quả cũ')) : response({ error: { code: 'INTERNAL_FAILURE' } }, 503)));
    expect(screen.getByText('Kết quả mới')).toBeVisible();
    expect(screen.queryByText('Kết quả cũ')).not.toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('does not restore old results after the query is cleared', async () => {
    const pending = deferred();
    const input = await openSearch(() => pending.promise);
    fireEvent.change(input, { target: { value: 'old' } });
    await userEvent.click(screen.getByRole('button', { name: 'Xóa từ khóa tìm kiếm' }));
    await act(async () => pending.resolve(response(result('Không được trở lại'))));
    expect(input).toHaveValue('');
    expect(screen.queryByText('Không được trở lại')).not.toBeInTheDocument();
    expect(screen.queryByText('Không tìm thấy kết quả.')).not.toBeInTheDocument();
  });

  it('keeps the query but discards the closed panel request on reopening', async () => {
    const pending = deferred();
    let calls = 0;
    const input = await openSearch(() => ++calls === 1 ? pending.promise : Promise.resolve(response(result('Phiên tìm mới'))));
    fireEvent.change(input, { target: { value: 'same' } });
    await userEvent.click(screen.getByRole('button', { name: 'Tìm kiếm' }));
    await userEvent.click(screen.getByRole('button', { name: 'Tìm kiếm' }));
    expect(screen.getByLabelText('Tìm kiếm giao dịch và ví')).toHaveValue('same');
    await act(async () => pending.resolve(response(result('Phiên đã đóng'))));
    expect(await screen.findByText('Phiên tìm mới')).toBeVisible();
    expect(screen.queryByText('Phiên đã đóng')).not.toBeInTheDocument();
  });

  it('does not publish an in-flight response after going offline', async () => {
    const pending = deferred();
    const input = await openSearch(() => pending.promise);
    fireEvent.change(input, { target: { value: 'offline' } });
    await act(async () => {
      Object.defineProperty(navigator, 'onLine', { configurable: true, value: false });
      window.dispatchEvent(new Event('offline'));
    });
    await act(async () => pending.resolve(response(result('Đến quá muộn'))));
    expect(screen.queryByText('Đến quá muộn')).not.toBeInTheDocument();
    expect(screen.queryByText('Không tìm thấy kết quả.')).not.toBeInTheDocument();
    await waitFor(() => expect(screen.getByText('Tìm kiếm cần kết nối mạng.')).toBeVisible());
  });

  it('does not reveal pending search results after logout', async () => {
    const pending = deferred();
    const input = await openSearch(() => pending.promise);
    fireEvent.change(input, { target: { value: 'private' } });
    await userEvent.click(screen.getByRole('button', { name: 'Tài khoản' }));
    await userEvent.click(screen.getByRole('button', { name: 'Đăng xuất' }));
    await screen.findByRole('button', { name: 'Đăng nhập bằng Google' });
    await act(async () => pending.resolve(response(result('Kết quả riêng tư'))));
    expect(screen.queryByText('Kết quả riêng tư')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('Tìm kiếm giao dịch và ví')).not.toBeInTheDocument();
  });
});
