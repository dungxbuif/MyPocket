import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { AccountScreen, type AccountScreenProps } from './AccountScreen';

const key = { id: 'key-1', user_id: 'owner', name: 'Integration', key_prefix: 'mpk_fixture', created_at: '2026-09-09T00:00:00Z' };
const created = { ...key, plaintext: 'mpk_nonsecret_test_fixture' };
const input: AccountScreenProps = {
  authState: { status: 'authenticated', user: { id: 'owner', email: 'owner@example.test', email_verified: true, display_name: 'Owner', avatar_url: '' } },
  assets: [], portfolioSummary: null, privacyMasked: false, online: true,
  onCreateAsset: vi.fn(), onAssetChanged: vi.fn(), onAssetArchived: vi.fn(), onLogout: vi.fn(), formatVND: String,
};
const ok = (body: unknown) => new Response(JSON.stringify(body), { status: 200 });
const failure = () => new Response(JSON.stringify({ error: { code: 'INTERNAL_FAILURE', message: 'DO_NOT_LEAK_RAW_EXCEPTION' }, correlation_id: 'req_key_failure' }), { status: 500 });

function mount(handler: (method: string, path: string) => Response) {
  const fetcher = vi.fn(async (url: RequestInfo | URL, options?: RequestInit) => {
    const path = String(url);
    return path.includes('/api-keys') ? handler(options?.method ?? 'GET', path) : ok({ allowed: false });
  });
  vi.stubGlobal('fetch', fetcher);
  render(<AccountScreen {...input} />);
  return fetcher;
}

afterEach(() => { vi.restoreAllMocks(); vi.unstubAllGlobals(); });

describe('API key operation feedback', () => {
  it('ignores stale list results after reconnect and preserves the current loading lock', async () => {
    let reads = 0;
    let resolveOld!: (response: Response) => void;
    let resolveNew!: (response: Response) => void;
    vi.stubGlobal('fetch', vi.fn(async (url: RequestInfo | URL) => {
      if (!String(url).includes('/api-keys')) return ok({ allowed: false });
      reads++;
      if (reads === 1) return ok({ keys: [key] });
      return new Promise<Response>(resolve => { if (reads === 2) resolveOld = resolve; else resolveNew = resolve; });
    }));
    const view = render(<AccountScreen {...input} />);
    await screen.findByText('Integration');
    await userEvent.click(screen.getByRole('button', { name: 'Tải lại danh sách key' }));
    view.rerender(<AccountScreen {...input} online={false} />);
    expect(screen.queryByText('Đang tải danh sách API key…')).not.toBeInTheDocument();
    view.rerender(<AccountScreen {...input} online />);
    await waitFor(() => expect(reads).toBe(3));
    await act(async () => { resolveOld(ok({ keys: [] })); });
    expect(screen.getByText('Integration')).toBeVisible();
    expect(screen.getByRole('button', { name: 'Tạo key' })).toBeDisabled();
    await act(async () => { resolveNew(ok({ keys: [{ ...key, revoked_at: '2026-09-09T01:00:00Z' }] })); });
    expect(screen.getByRole('button', { name: 'Tạo key' })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'Revoke' })).toBeDisabled();
  });

  it('distinguishes a failed list request from an empty list and allows read retry', async () => {
    let failed = true;
    mount(() => failed ? failure() : ok({ keys: [key] }));
    expect(await screen.findByRole('alert')).toHaveTextContent('req_key_failure');
    expect(screen.queryByText('Chưa có API key')).not.toBeInTheDocument();
    expect(screen.queryByText('DO_NOT_LEAK_RAW_EXCEPTION')).not.toBeInTheDocument();
    failed = false;
    await userEvent.click(screen.getByRole('button', { name: 'Tải lại danh sách key' }));
    expect(await screen.findByText('Integration')).toBeVisible();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('keeps the requested name after create rejection and does not auto-retry', async () => {
    const fetcher = mount(method => method === 'GET' ? ok({ keys: [] }) : failure());
    await screen.findByText('Chưa có API key');
    await userEvent.clear(screen.getByLabelText('Tên API key'));
    await userEvent.type(screen.getByLabelText('Tên API key'), 'Flow của tôi');
    await userEvent.click(screen.getByRole('button', { name: 'Tạo key' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('req_key_failure');
    expect(screen.getByLabelText('Tên API key')).toHaveValue('Flow của tôi');
    expect(fetcher.mock.calls.filter(([, options]) => options?.method === 'POST')).toHaveLength(1);
    expect(screen.queryByText('Chỉ hiển thị một lần')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Tải lại danh sách key' })).toBeEnabled();
  });

  it('retains a confirmed created key when list refresh fails', async () => {
    let createdOnServer = false;
    mount(method => {
      if (method === 'POST') { createdOnServer = true; return ok({ key: created }); }
      return createdOnServer ? failure() : ok({ keys: [] });
    });
    await screen.findByText('Chưa có API key');
    await userEvent.click(screen.getByRole('button', { name: 'Tạo key' }));
    expect(await screen.findByText(created.plaintext)).toBeVisible();
    expect(await screen.findByText('Integration')).toBeVisible();
    expect(await screen.findByRole('alert')).toHaveTextContent('danh sách');
    expect(screen.getByRole('button', { name: 'Revoke' })).toBeEnabled();
  });

  it('keeps a key active after revoke rejection', async () => {
    mount(method => method === 'GET' ? ok({ keys: [key] }) : failure());
    await userEvent.click(await screen.findByRole('button', { name: 'Revoke' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('req_key_failure');
    expect(screen.getByRole('button', { name: 'Revoke' })).toBeEnabled();
    expect(screen.queryByText('Đã revoke', { exact: false })).not.toBeInTheDocument();
  });

  it('does not re-enable a confirmed revoked key when refresh fails', async () => {
    let revoked = false;
    mount(method => {
      if (method === 'POST') { revoked = true; return ok({}); }
      return revoked ? failure() : ok({ keys: [key] });
    });
    await userEvent.click(await screen.findByRole('button', { name: 'Revoke' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('danh sách');
    expect(screen.getByRole('button', { name: 'Revoke' })).toBeDisabled();
  });

  it.each([true, false])('reports clipboard outcome (success=%s)', async success => {
    const user = userEvent.setup();
    mount(method => method === 'POST' ? ok({ key: created }) : ok({ keys: [] }));
    await screen.findByText('Chưa có API key');
    await user.click(screen.getByRole('button', { name: 'Tạo key' }));
    await screen.findByText(created.plaintext);
    const copy = vi.spyOn(navigator.clipboard, 'writeText');
    if (success) copy.mockResolvedValue(); else copy.mockRejectedValue(new Error('clipboard denied'));
    await user.click(screen.getByRole('button', { name: 'Copy' }));
    if (success) expect(await screen.findByRole('status')).toHaveTextContent('Đã sao chép');
    else expect(await screen.findByRole('alert')).toHaveTextContent('Không sao chép được');
    expect(copy).toHaveBeenCalledWith(created.plaintext);
  });

  it('ignores completion of copying a key that has since been revoked', async () => {
    const user = userEvent.setup();
    mount((method, path) => method === 'POST' ? ok(path.endsWith('/revoke') ? {} : { key: created }) : ok({ keys: [key] }));
    await screen.findByText('Integration');
    await user.click(screen.getByRole('button', { name: 'Tạo key' }));
    await screen.findByText(created.plaintext);
    let finishCopy!: () => void;
    vi.spyOn(navigator.clipboard, 'writeText').mockImplementation(() => new Promise<void>(resolve => { finishCopy = resolve; }));
    await user.click(screen.getByRole('button', { name: 'Copy' }));
    await user.click(screen.getByRole('button', { name: 'Revoke' }));
    expect(screen.queryByText(created.plaintext)).not.toBeInTheDocument();
    await act(async () => { finishCopy(); });
    expect(screen.queryByText('Đã sao chép API key.')).not.toBeInTheDocument();
  });
});
