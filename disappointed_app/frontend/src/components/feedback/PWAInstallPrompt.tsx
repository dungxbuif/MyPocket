import { useEffect, useRef, useState } from 'react';
import { PillButton } from '../../app/components';

export type InstallPromptEvent = Event & {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: string }>;
};

export function PWAInstallPrompt({ prompt, onConsumed }: {
  prompt: InstallPromptEvent | null;
  onConsumed: () => void;
}) {
  const [dismissed, setDismissed] = useState(false);
  const [installed, setInstalled] = useState(() =>
    Boolean(window.matchMedia?.('(display-mode: standalone)').matches ||
      (navigator as Navigator & { standalone?: boolean }).standalone));
  const [helpOpen, setHelpOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const ref = useRef<HTMLElement>(null);
  const visible = !dismissed && !installed;

  useEffect(() => {
    const onInstalled = () => setInstalled(true);
    window.addEventListener('appinstalled', onInstalled);
    return () => window.removeEventListener('appinstalled', onInstalled);
  }, []);

  useEffect(() => {
    if (!visible || !ref.current) return;
    const root = document.documentElement;
    const measure = () => root.style.setProperty('--install-prompt-height', `${Math.ceil(ref.current?.getBoundingClientRect().height ?? 0)}px`);
    measure();
    const observer = typeof ResizeObserver === 'undefined' ? null : new ResizeObserver(measure);
    observer?.observe(ref.current);
    return () => {
      observer?.disconnect();
      root.style.removeProperty('--install-prompt-height');
    };
  }, [visible, helpOpen, error]);

  async function install() {
    if (!prompt) { setHelpOpen(true); return; }
    if (busy) return;
    setBusy(true);
    setError('');
    try {
      await prompt.prompt();
      const choice = await prompt.userChoice;
      if (choice.outcome === 'accepted') setDismissed(true);
    } catch {
      setError('Không mở được cài đặt. Hãy thử lại từ menu trình duyệt.');
    } finally {
      onConsumed();
      setBusy(false);
    }
  }

  if (!visible) return null;
  return (
    <aside ref={ref} className="pwa-install-prompt" role="dialog" aria-label="Cài MyPocket">
      <strong>Cài MyPocket</strong><span>Dùng nhanh hơn từ màn hình chính.</span>
      {helpOpen ? <><span>Safari: Chia sẻ → Thêm vào Màn hình chính</span><span>Chrome/Edge: menu → Cài ứng dụng hoặc Thêm vào màn hình chính (nếu được hỗ trợ).</span></> : null}
      {error ? <span role="alert">{error}</span> : null}
      <div className="flex gap-2">
        <PillButton disabled={busy} onClick={() => void install()}>{busy ? 'Đang mở' : 'Cài ứng dụng'}</PillButton>
        <PillButton onClick={() => setDismissed(true)}>Để sau</PillButton>
      </div>
    </aside>
  );
}
