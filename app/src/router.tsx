import { createRootRoute, createRoute, createRouter } from "@tanstack/react-router";
import { FinancePrototypePage } from "./atomic/pages/FinancePrototypePage";

const rootRoute = createRootRoute({
  component: FinancePrototypePage,
});

const indexRoute = createRoute({ getParentRoute: () => rootRoute, path: "/" });
const transactionsRoute = createRoute({ getParentRoute: () => rootRoute, path: "transactions" });
const budgetsRoute = createRoute({ getParentRoute: () => rootRoute, path: "budgets" });
const jarsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "jars",
  validateSearch: (search: Record<string, unknown>) => ({ month: typeof search.month === "string" ? search.month : undefined }),
});
const monthDetailRoute = createRoute({ getParentRoute: () => rootRoute, path: "months/$month" });
// Reports stay deliberately out of the active route tree until the reporting API is implemented.
// const reportsRoute = createRoute({ getParentRoute: () => rootRoute, path: "reports" });
const accountRoute = createRoute({ getParentRoute: () => rootRoute, path: "account" });
const groupsRoute = createRoute({ getParentRoute: () => rootRoute, path: "account/groups" });
const groupNewRoute = createRoute({ getParentRoute: () => rootRoute, path: "account/groups/new" });
const groupEditRoute = createRoute({ getParentRoute: () => rootRoute, path: "account/groups/$categoryId/edit" });
const walletsRoute = createRoute({ getParentRoute: () => rootRoute, path: "account/wallets" });
const authRoute = createRoute({ getParentRoute: () => rootRoute, path: "auth/google" });
const authCallbackRoute = createRoute({ getParentRoute: () => rootRoute, path: "auth/callback" });

const routeTree = rootRoute.addChildren([
  indexRoute,
  transactionsRoute,
  budgetsRoute,
  jarsRoute,
  monthDetailRoute,
  // reportsRoute,
  accountRoute,
  groupsRoute,
  groupNewRoute,
  groupEditRoute,
  walletsRoute,
  authRoute,
  authCallbackRoute,
]);

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
