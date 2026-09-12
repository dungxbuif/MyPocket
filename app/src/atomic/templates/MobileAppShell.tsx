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
}: {
  tab: PrototypeTab;
  masked: boolean;
  children: ReactNode;
  onTabChange: (tab: PrototypeTab) => void;
  onToggleMask: () => void;
  onAdd: () => void;
}) {
  return (
    <main className="min-h-screen bg-[#fbf9f9] text-[#1b1c1c]">
      <div className="mx-auto min-h-screen max-w-[430px] bg-[#fbf9f9] pb-28 shadow-[0_0_40px_rgb(0_0_0/0.08)]">
        <AppHeader masked={masked} onToggleMask={onToggleMask} />
        <section className="space-y-3 px-4">{children}</section>
        <BottomNavigation tab={tab} onTabChange={onTabChange} onAdd={onAdd} />
      </div>
    </main>
  );
}

