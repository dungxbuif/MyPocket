import type { ReactNode } from "react";
import { AppHeader } from "../organisms/AppHeader";
import { BottomNavigation } from "../organisms/BottomNavigation";
import type { PrototypeTab } from "../pages/FinancePrototypePage";

export function MobileAppShell({
  tab,
  masked,
  children,
  onTabChange,
  onToggleMask,
  onAdd,
  showHeader = true,
}: {
  tab: PrototypeTab;
  masked: boolean;
  children: ReactNode;
  onTabChange: (tab: PrototypeTab) => void;
  onToggleMask: () => void;
  onAdd: () => void;
  showHeader?: boolean;
}) {
  return (
    <main className="min-h-screen bg-canvas text-ink">
      <div className="mx-auto min-h-screen max-w-[430px] bg-canvas pb-28 shadow-shell">
        {showHeader ? <AppHeader masked={masked} onToggleMask={onToggleMask} /> : null}
        <section className="space-y-3 px-4">{children}</section>
        <BottomNavigation tab={tab} onTabChange={onTabChange} onAdd={onAdd} />
      </div>
    </main>
  );
}
