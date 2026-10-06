import { strict as assert } from "node:assert";
import { createServer } from "vite";

const server = await createServer({ server: { middlewareMode: true }, appType: "custom" });
try {
  const { transactionsSearch, walletIDFromTransactionsSearch } = await server.ssrLoadModule("/src/services/walletNavigation.ts");
  assert.deepEqual(transactionsSearch("wallet-1"), { wallet_id: "wallet-1" });
  assert.deepEqual(transactionsSearch(null), {});
  assert.equal(walletIDFromTransactionsSearch("?wallet_id=wallet-1"), "wallet-1");
  assert.equal(walletIDFromTransactionsSearch("wallet_id=wallet-2"), "wallet-2");
  assert.equal(walletIDFromTransactionsSearch("?wallet_id="), "");
  const fs = await import("node:fs/promises");
  const walletSource = await fs.readFile(new URL("../src/atomic/organisms/WalletManagementPanel.tsx", import.meta.url), "utf8");
  const transactionSource = await fs.readFile(new URL("../src/atomic/organisms/TransactionsPanel.tsx", import.meta.url), "utf8");
  assert.match(walletSource, /navigate\(\{ to: "\/transactions", search: transactionsSearch\(id\) \}\)/);
  assert.match(transactionSource, /walletIDFromTransactionsSearch\(location\.searchStr\)/);
  console.log("Wallet transaction navigation contract passed.");
} finally {
  await server.close();
}
