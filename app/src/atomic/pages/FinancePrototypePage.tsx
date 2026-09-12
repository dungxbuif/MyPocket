import { useState } from "react";
import { AccountPanel } from "../organisms/AccountPanel";
import { BudgetsPanel } from "../organisms/BudgetsPanel";
import { OverviewPanel } from "../organisms/OverviewPanel";
import { QuickAddSheet } from "../organisms/QuickAddSheet";
import { ReportsPanel } from "../organisms/ReportsPanel";
import { TransactionsPanel } from "../organisms/TransactionsPanel";
import { MobileAppShell } from "../templates/MobileAppShell";

export type PrototypeTab = "overview" | "transactions" | "budgets" | "reports" | "account";

export function FinancePrototypePage() {
  const [tab, setTab] = useState<PrototypeTab>("overview");
  const [masked, setMasked] = useState(false);
  const [quickAddOpen, setQuickAddOpen] = useState(false);

  return (
    <>
      <MobileAppShell
        tab={tab}
        masked={masked}
        onTabChange={setTab}
        onToggleMask={() => setMasked((value) => !value)}
        onAdd={() => setQuickAddOpen(true)}
      >
        {tab === "overview" ? <OverviewPanel masked={masked} /> : null}
        {tab === "transactions" ? <TransactionsPanel /> : null}
        {tab === "budgets" ? <BudgetsPanel masked={masked} /> : null}
        {tab === "reports" ? <ReportsPanel masked={masked} /> : null}
        {tab === "account" ? <AccountPanel /> : null}
      </MobileAppShell>
      {quickAddOpen ? <QuickAddSheet onClose={() => setQuickAddOpen(false)} /> : null}
    </>
  );
}
