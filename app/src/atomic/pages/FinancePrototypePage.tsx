import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "@tanstack/react-router";
import { AccountPanel } from "../organisms/AccountPanel";
import { BudgetsPanel } from "../organisms/BudgetsPanel";
import { OverviewPanel } from "../organisms/OverviewPanel";
import { QuickAddSheet } from "../organisms/QuickAddSheet";
import { ReportsPanel } from "../organisms/ReportsPanel";
import { TransactionsPanel } from "../organisms/TransactionsPanel";
import { GroupManagementPanel } from "../organisms/GroupManagementPanel";
import { MobileAppShell } from "../templates/MobileAppShell";
import { APP_CONFIG, APP_ROUTES, API_ROUTES } from "../../config/app";
import {
  clearSession,
  extractFixtureLoginFromQueryParams,
  fetchHome,
  fetchProfile,
  getStoredToken,
  HomeOutput,
  isAuthCallbackPath,
  loginWithGoogleCode,
  loginWithGoogleFixture,
  writeSession,
  type LoginOutput,
  type UserProfile,
} from "../../services/auth";

export type PrototypeTab = "overview" | "transactions" | "budgets" | "reports" | "account";

type AuthPhase = "checking" | "authenticated" | "unauthenticated";

type AuthState = {
  phase: AuthPhase;
  errorMessage: string;
  user: UserProfile | null;
  home: HomeOutput | null;
};

export function FinancePrototypePage() {
  const [auth, setAuth] = useState<AuthState>({
    phase: "checking",
    errorMessage: "",
    user: null,
    home: null,
  });
  const location = useLocation();
  const navigate = useNavigate();
  const tab = tabFromPath(location.pathname);
  const [masked, setMasked] = useState(false);
  const [quickAddOpen, setQuickAddOpen] = useState(false);
  const [isLoggingIn, setIsLoggingIn] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const handleAuth = async (): Promise<void> => {
      const url = new URL(window.location.href);
      const query = new URLSearchParams(url.search);
      const statePath = url.pathname;

      try {
        if (statePath === "/auth/google") {
          window.location.replace(API_ROUTES.START_GOOGLE_AUTH);
          return;
        }
        if (isAuthCallbackPath(statePath)) {
          const directLogin = extractFixtureLoginFromQueryParams(query);
          if (directLogin) {
            await acceptLogin(directLogin);
          } else if (query.get("code")) {
            const loginOutput = await loginWithGoogleCode(query.get("code") ?? "");
            await acceptLogin(loginOutput);
          } else if (query.get("subject") && query.get("email")) {
            const loginOutput = await loginWithGoogleFixture(query);
            await acceptLogin(loginOutput);
          } else if (query.get("error")) {
            setAuth((prev) => ({ ...prev, phase: "unauthenticated", errorMessage: query.get("error") ?? "Đăng nhập Google bị hủy" }));
          } else {
            setAuth((prev) => ({ ...prev, phase: "unauthenticated", errorMessage: "Không nhận được thông tin xác thực từ Google" }));
          }
          void navigate({ to: "/", search: {}, replace: true });
          return;
        }

        const token = getStoredToken();
        if (!token) {
          setAuth({ phase: "unauthenticated", errorMessage: "", user: null, home: null });
          return;
        }
        const [profile, home] = await Promise.all([fetchProfile(token), fetchHome(token)]);
        if (!cancelled) {
          setAuth({ phase: "authenticated", errorMessage: "", user: profile, home });
        }
      } catch (error) {
        clearSession();
        if (!cancelled) {
          setAuth({
            phase: "unauthenticated",
            errorMessage: error instanceof Error ? error.message : "Không thể lấy thông tin tài khoản",
            user: null,
            home: null,
          });
        }
      }
    };

    void handleAuth();

    return () => {
      cancelled = true;
    };
  }, []);

  const acceptLogin = async (loginOutput: LoginOutput): Promise<void> => {
    writeSession({
      token: loginOutput.token,
      user: loginOutput.user,
      expiresAt: loginOutput.expires_at,
    });
    const token = loginOutput.token;
    const [profile, home] = await Promise.all([fetchProfile(token), fetchHome(token)]);
    setAuth({
      phase: "authenticated",
      errorMessage: "",
      user: profile,
      home,
    });
  };

  const handleGoogleLogin = (): void => {
    if (isLoggingIn) return;
    setIsLoggingIn(true);
    setAuth((current) => ({ ...current, errorMessage: "" }));
    const frontendCallback = encodeURIComponent(`${window.location.origin}${APP_ROUTES.AUTH_CALLBACK_PATH}`);
    window.location.href = `${API_ROUTES.START_GOOGLE_AUTH}?frontend_callback=${frontendCallback}`;
    setIsLoggingIn(false);
  };

  const handleLogout = (): void => {
    clearSession();
    setAuth({
      phase: "unauthenticated",
      errorMessage: "",
      user: null,
      home: null,
    });
  };

  if (auth.phase === "checking") {
    return <AuthLoadingScreen />;
  }

  if (auth.phase === "unauthenticated") {
    return <AuthGateScreen isLoading={isLoggingIn} errorMessage={auth.errorMessage} onGoogleLogin={handleGoogleLogin} />;
  }

  const userName = auth.user?.name || "Người dùng";

  return (
    <>
      <MobileAppShell
        tab={tab}
        masked={masked}
        onTabChange={(nextTab) => void navigate({ to: pathFromTab(nextTab) })}
        onToggleMask={() => setMasked((value) => !value)}
        onAdd={() => setQuickAddOpen(true)}
      >
        {tab === "overview" ? <OverviewPanel masked={masked} /> : null}
        {tab === "transactions" ? <TransactionsPanel /> : null}
        {tab === "budgets" ? <BudgetsPanel masked={masked} /> : null}
        {tab === "reports" ? <ReportsPanel masked={masked} /> : null}
        {tab === "account" ? (location.pathname.startsWith("/account/groups") ? <GroupManagementPanel /> : <AccountPanel user={auth.user} onLogout={handleLogout} />) : null}
      </MobileAppShell>
      {quickAddOpen ? <QuickAddSheet onClose={() => setQuickAddOpen(false)} /> : null}
    </>
  );
}

function tabFromPath(pathname: string): PrototypeTab {
  if (pathname.startsWith("/transactions")) return "transactions";
  if (pathname.startsWith("/budgets")) return "budgets";
  if (pathname.startsWith("/reports")) return "reports";
  if (pathname.startsWith("/account")) return "account";
  return "overview";
}

function pathFromTab(tab: PrototypeTab): "/" | "/transactions" | "/budgets" | "/reports" | "/account" {
  const paths: Record<PrototypeTab, "/" | "/transactions" | "/budgets" | "/reports" | "/account"> = {
    overview: "/",
    transactions: "/transactions",
    budgets: "/budgets",
    reports: "/reports",
    account: "/account",
  };
  return paths[tab];
}

function AuthLoadingScreen() {
  return (
    <main className="min-h-screen bg-[#fbf9f9] text-[#1b1c1c]">
      <div className="mx-auto flex min-h-screen max-w-[430px] items-center justify-center bg-[#fbf9f9] px-4">
        <p>Đang kiểm tra phiên đăng nhập...</p>
      </div>
    </main>
  );
}

function AuthGateScreen({
  isLoading,
  errorMessage,
  onGoogleLogin,
}: {
  isLoading: boolean;
  errorMessage: string;
  onGoogleLogin: () => void;
}) {
  return (
    <main className="min-h-screen bg-[#fbf9f9] text-[#1b1c1c]">
      <div className="mx-auto flex min-h-screen max-w-[430px] flex-col items-center justify-center gap-4 px-6">
        <div className="grid h-16 w-16 place-items-center rounded-full bg-[#006e1c] text-2xl font-bold text-white">M</div>
        <h1 className="text-center text-2xl font-bold">MyPocket</h1>
        <p className="text-center text-sm text-[#3f4a3c]">Đăng nhập bằng Google để dùng đầy đủ tính năng.</p>
        {errorMessage ? <p className="rounded-2xl bg-[#ffe6e6] p-3 text-sm text-[#8d1717]">{errorMessage}</p> : null}
        <button
          type="button"
          onClick={() => void onGoogleLogin()}
          className="mt-2 rounded-full bg-[#006e1c] px-5 py-3 font-semibold text-white disabled:opacity-50"
          disabled={isLoading}
        >
          {isLoading ? "Đang đăng nhập..." : "Đăng nhập bằng Google"}
        </button>
      </div>
    </main>
  );
}
