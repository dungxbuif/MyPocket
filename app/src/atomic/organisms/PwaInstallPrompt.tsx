import { useEffect, useRef, useState } from "react";
import { BaseButton } from "../atoms/BaseButton";
import { SurfaceCard } from "../atoms/SurfaceCard";
import { Text } from "../atoms/Text";

type InstallChoice = { outcome: "accepted" | "dismissed" };
type BeforeInstallPromptEvent = Event & { prompt: () => Promise<void>; userChoice: Promise<InstallChoice> };

export function PwaInstallPrompt() {
  const deferred = useRef<BeforeInstallPromptEvent | null>(null);
  const [available, setAvailable] = useState(false);
  const [closed, setClosed] = useState(false);

  useEffect(() => {
    if (window.matchMedia("(display-mode: standalone)").matches || (navigator as Navigator & { standalone?: boolean }).standalone) return;
    try { if (window.sessionStorage.getItem("mypocket.pwa.install.dismissed") === "1") setClosed(true); } catch { /* private mode */ }
    const onBeforeInstallPrompt = (event: Event) => {
      event.preventDefault();
      deferred.current = event as BeforeInstallPromptEvent;
      setAvailable(true);
    };
    const onInstalled = () => { deferred.current = null; setAvailable(false); };
    window.addEventListener("beforeinstallprompt", onBeforeInstallPrompt);
    window.addEventListener("appinstalled", onInstalled);
    return () => { window.removeEventListener("beforeinstallprompt", onBeforeInstallPrompt); window.removeEventListener("appinstalled", onInstalled); };
  }, []);

  if (!available || closed || !deferred.current) return null;
  const dismiss = () => {
    setClosed(true);
    try { window.sessionStorage.setItem("mypocket.pwa.install.dismissed", "1"); } catch { /* private mode */ }
  };
  const install = async () => {
    const event = deferred.current;
    if (!event) return;
    await event.prompt();
    const choice = await event.userChoice;
    deferred.current = null;
    setAvailable(false);
    if (choice.outcome === "dismissed") dismiss();
  };
  return <SurfaceCard role="status" aria-live="polite" padding="md" tone="form" className="fixed inset-x-4 bottom-24 z-40 mx-auto max-w-[398px]"><div className="flex items-start gap-3"><div className="min-w-0 flex-1"><Text weight="semibold">Cài MyPocket</Text><Text size="sm" tone="secondary" className="mt-1">Mở nhanh như ứng dụng và dùng tốt hơn trên điện thoại.</Text></div><BaseButton variant="ghost" size="sm" aria-label="Đóng gợi ý cài đặt" onClick={dismiss}>Để sau</BaseButton></div><BaseButton className="mt-3 w-full" onClick={() => void install()}>Cài đặt</BaseButton></SurfaceCard>;
}
